package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/hashicorp/go-multierror"
)

func ParseTime(stamp string) (time.Time, error) {
	var result *multierror.Error

	args := strings.Split(stamp, " ")
	t, err := parseDurationAgruments(args)
	if err == nil {
		return t, nil
	}
	result = multierror.Append(result, err)

	t, err = parseTimeStamp(stamp)
	if err == nil {
		return t, nil
	}
	result = multierror.Append(result, err)

	t, err = dateparse.ParseAny(stamp)
	if err == nil {
		return t, nil
	}
	result = multierror.Append(result, err)

	return time.Time{}, result.ErrorOrNil()
}

func parseTimeStamp(stamp string) (time.Time, error) {
	if err := matchStr(strings.ToLower(stamp), "^[0-2]?[0-9](:[0-5][0-9])?(pm|am)?$"); err != nil {
		return time.Time{}, fmt.Errorf("failed to parse for a normal time: %s: %v", stamp, err)
	}

	if err := matchStr(strings.ToLower(stamp), "^[0-9]?[0-9]:[0-5][0-9]$"); err == nil {
		split := strings.Split(stamp, ":")
		hour, err := strconv.Atoi(split[0])
		if err != nil {
			return time.Time{}, err
		}

		minutes, err := strconv.Atoi(split[1])
		if err != nil {
			return time.Time{}, err
		}

		if hour > 23 {
			return time.Time{}, fmt.Errorf("bad hour: %s", stamp)
		}

		timestamp := time.Now()

		timestamp = timestamp.Add(time.Hour * time.Duration((hour - time.Now().Hour())))
		timestamp = timestamp.Add(time.Minute * time.Duration((minutes - time.Now().Minute())))

		if hour < time.Now().Hour() {
			timestamp = timestamp.Add(time.Hour * 24)
		}

		if minutes < time.Now().Minute() {
			timestamp = timestamp.Add(time.Minute * 60)
		}

		return timestamp, nil
	}

	if err := matchStr(strings.ToLower(stamp), "^[0-9]?[0-9]:[0-5][0-9](am|pm)$"); err == nil {
		split := strings.Split(stamp, ":")
		hour, err := strconv.Atoi(split[0])
		if err != nil {
			return time.Time{}, err
		}

		if hour > 11 {
			return time.Time{}, fmt.Errorf("bad hour: %s", stamp)
		}

		minutes, err := strconv.Atoi(split[1][:2])
		if err != nil {
			return time.Time{}, err
		}

		if split[1][2:] == "pm" {
			hour += 12
		}

		timestamp := time.Now()

		timestamp = timestamp.Add(time.Hour * time.Duration((hour - time.Now().Hour())))
		timestamp = timestamp.Add(time.Minute * time.Duration((minutes - time.Now().Minute())))

		if hour < time.Now().Hour() {
			timestamp = timestamp.Add(time.Hour * 24)
		}

		if minutes < time.Now().Minute() {
			timestamp = timestamp.Add(time.Minute * 60)
		}

		return timestamp, nil
	}

	return time.Time{}, fmt.Errorf("failed to parse time: %s", stamp)
}

func parseDurationAgruments(args []string) (time.Time, error) {
	var result *multierror.Error
	total := time.Now()

	for _, arg := range args {
		if seconds, err := strconv.Atoi(arg); err == nil {
			total = total.Add(time.Second * time.Duration(seconds))
			continue
		}

		if err := matchStr(strings.ToLower(arg), "^forever$"); err == nil {
			total = time.Unix(1<<63-62135596801, 999999999)
			break
		}

		if err := matchStr(strings.ToLower(arg), "^now$"); err == nil {
			continue
		}

		parseTime, err := matchNum(arg, "^[0-9]+(.[0-9]+|)n$")
		if err == nil {
			total = total.Add(time.Nanosecond * time.Duration(parseTime))
			continue
		}

		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)nanoseconds$")
		if err == nil {
			total = total.Add(time.Nanosecond * time.Duration(parseTime))
			continue
		}

		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)s$")
		if err == nil {
			total = total.Add(time.Second * time.Duration(parseTime))
			continue
		}
		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)seconds$")
		if err == nil {
			total = total.Add(time.Second * time.Duration(parseTime))
			continue
		}

		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)m$")
		if err == nil {
			total = total.Add(time.Minute * time.Duration(parseTime))
			continue
		}

		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)minutes$")
		if err == nil {
			total = total.Add(time.Minute * time.Duration(parseTime))
			continue
		}

		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)h$")
		if err == nil {
			total = total.Add(time.Hour * time.Duration(parseTime))
			continue
		}
		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)hours$")
		if err == nil {
			total = total.Add(time.Hour * time.Duration(parseTime))
			continue
		}

		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)d$")
		if err == nil {
			total = total.Add(time.Hour * time.Duration(parseTime*24))
			continue
		}
		parseTime, err = matchNum(arg, "^[0-9]+(.[0-9]+|)days$")
		if err == nil {
			total = total.Add(time.Hour * time.Duration(parseTime*24))
			continue
		}

		result = multierror.Append(result, fmt.Errorf("could not parse argument: %s", arg))
	}

	return total, result.ErrorOrNil()
}

func matchNum(str, regex string) (num float64, err error) {
	r := regexp.MustCompile(regex)
	match := r.FindStringSubmatch(str)
	if len(match) > 0 {
		num, err = getNum(match[0])
		if err != nil {
			return -1, err
		}

		return num, nil
	}

	return -1, fmt.Errorf("'%s' didn't match regex '%s'", str, regex)
}

func matchStr(str, regex string) error {
	r := regexp.MustCompile(regex)
	match := r.FindStringSubmatch(str)
	if len(match) < 1 {
		return fmt.Errorf("'%s' didn't match regex '%s'", str, regex)
	}

	return nil
}

func getNum(str string) (num float64, err error) {
	str = str[:len(str)-1]
	num, err = strconv.ParseFloat(str, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to convert arg to integer: %v", err)
	}

	return num, nil
}
