package utils

//TODO make this better

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func Test_ParseTime_Human_Read_NoError(t *testing.T) {
	var stamp, tmp1, tmp2, tmp3, tmp4 string
	var day, hour, min, sec, double time.Duration
	var corrTime time.Time

	seed := rand.NewSource(time.Now().UnixNano())
	gen := rand.New(seed)

	num := gen.Intn(1000)
	sec = time.Duration(num) * time.Second
	stamp = fmt.Sprintf("%d", num)
	corrTime = time.Now().Add(sec)
	test_stamp(stamp, corrTime, true, t)

	corrTime = time.Unix(1<<63-62135596801, 999999999)
	test_stamp("FOREVER", corrTime, true, t)
	test_stamp("Forever", corrTime, true, t)
	test_stamp("forever", corrTime, true, t)
	test_stamp("FoReVeR", corrTime, true, t)

	corrTime = time.Now()
	test_stamp("NOW", corrTime, true, t)
	test_stamp("Now", corrTime, true, t)
	test_stamp("now", corrTime, true, t)
	test_stamp("NoW", corrTime, true, t)

	stamp, day = gen_day(gen)
	corrTime = time.Now().Add(day)
	test_stamp(stamp, corrTime, true, t)

	stamp, hour = gen_hour(gen)
	corrTime = time.Now().Add(hour)
	test_stamp(stamp, corrTime, true, t)

	stamp, min = gen_min(gen)
	corrTime = time.Now().Add(min)
	test_stamp(stamp, corrTime, true, t)

	stamp, sec = gen_sec(gen)
	corrTime = time.Now().Add(sec)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, double = gen_day(gen)
	corrTime = time.Now().Add(day + double)
	stamp = fmt.Sprintf("now %s %s", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)

	tmp1, hour = gen_hour(gen)
	tmp2, double = gen_hour(gen)
	corrTime = time.Now().Add(hour + double)
	stamp = fmt.Sprintf("%s now %s", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)

	tmp1, min = gen_min(gen)
	tmp2, double = gen_min(gen)
	corrTime = time.Now().Add(min + double)
	stamp = fmt.Sprintf("%s %s", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, hour = gen_hour(gen)
	corrTime = time.Now().Add(day + hour)
	stamp = fmt.Sprintf("%s now %s", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour)
	stamp = fmt.Sprintf("%s %s", tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, min = gen_min(gen)
	corrTime = time.Now().Add(day + min)
	stamp = fmt.Sprintf("%s %s now", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + min)
	stamp = fmt.Sprintf("%s %s", tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, sec = gen_sec(gen)
	corrTime = time.Now().Add(day + sec)
	stamp = fmt.Sprintf("%s %s", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + sec)
	stamp = fmt.Sprintf("%s %s", tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, hour = gen_hour(gen)
	tmp2, min = gen_min(gen)
	corrTime = time.Now().Add(hour + min)
	stamp = fmt.Sprintf("%s now %s", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(hour + min)
	stamp = fmt.Sprintf("%s %s", tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, hour = gen_hour(gen)
	tmp2, sec = gen_sec(gen)
	corrTime = time.Now().Add(hour + sec)
	stamp = fmt.Sprintf("%s %s", tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(hour + sec)
	stamp = fmt.Sprintf("%s now %s", tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, hour = gen_hour(gen)
	tmp3, min = gen_min(gen)
	corrTime = time.Now().Add(day + hour + min)
	stamp = fmt.Sprintf("%s %s %s", tmp1, tmp2, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min)
	stamp = fmt.Sprintf("%s %s %s", tmp1, tmp3, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min)
	stamp = fmt.Sprintf("%s %s now %s", tmp3, tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min)
	stamp = fmt.Sprintf("%s %s %s", tmp3, tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min)
	stamp = fmt.Sprintf("%s %s %s", tmp2, tmp1, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min)
	stamp = fmt.Sprintf("%s %s %s", tmp2, tmp3, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, hour = gen_hour(gen)
	tmp3, sec = gen_sec(gen)
	corrTime = time.Now().Add(day + hour + sec)
	stamp = fmt.Sprintf("%s %s now %s", tmp1, tmp2, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp1, tmp3, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp3, tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + sec)
	stamp = fmt.Sprintf("now %s %s %s", tmp3, tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp2, tmp1, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp2, tmp3, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, hour = gen_hour(gen)
	tmp2, min = gen_min(gen)
	tmp3, sec = gen_sec(gen)
	corrTime = time.Now().Add(min + hour + sec)
	stamp = fmt.Sprintf("%s now %s %s", tmp1, tmp2, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(min + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp1, tmp3, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(min + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp3, tmp1, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(min + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp3, tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(min + hour + sec)
	stamp = fmt.Sprintf("%s %s %s", tmp2, tmp1, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(min + hour + sec)
	stamp = fmt.Sprintf("%s %s now %s", tmp2, tmp3, tmp1)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, hour = gen_hour(gen)
	tmp3, min = gen_min(gen)
	tmp4, sec = gen_sec(gen)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp2, tmp3, tmp4)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp2, tmp4, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp3, tmp2, tmp4)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp3, tmp4, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp4, tmp3, tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp4, tmp1, tmp3, tmp2)
	test_stamp(stamp, corrTime, true, t)

	tmp1, day = gen_day(gen)
	tmp2, hour = gen_hour(gen)
	tmp3, min = gen_min(gen)
	num = gen.Intn(1000)
	sec = time.Duration(num) * time.Second
	tmp4 = fmt.Sprintf("%d", num)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp2, tmp3, tmp4)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp2, tmp4, tmp3)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp3, tmp2, tmp4)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp1, tmp3, tmp4, tmp2)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp4, tmp3, tmp2, tmp1)
	test_stamp(stamp, corrTime, true, t)
	corrTime = time.Now().Add(day + hour + min + sec)
	stamp = fmt.Sprintf("%s %s %s %s", tmp4, tmp1, tmp3, tmp2)
	test_stamp(stamp, corrTime, true, t)

	corrTime = time.Unix(1<<63-62135596801, 999999999)
	stamp = fmt.Sprintf("%s forever %s %s %s", tmp4, tmp1, tmp3, tmp2)
	test_stamp(stamp, corrTime, true, t)

	corrTime = time.Now().Add(time.Hour).Add(time.Minute)
	stamp = fmt.Sprintf("%d:%02d", ((time.Now().Hour() + 1) % 24), ((time.Now().Minute() + 1) % 60))
	test_stamp(stamp, corrTime, false, t)

	corrTime = time.Now().Add(time.Hour).Add(time.Minute)
	if corrTime.Hour() <= 12 {
		corrTime = corrTime.Add(time.Hour * 12)
	}
	hur := time.Now().Hour()
	if hur >= 12 {
		hur -= 12
	}
	stamp = fmt.Sprintf("%d:%02dpm", hur+1, ((time.Now().Minute() + 1) % 60))
	test_stamp(stamp, corrTime, false, t)

	return
}

func Test_ParseTime_Human_Read_Error(t *testing.T) {
	test_error_stamp("foo", t)
	test_error_stamp("bar", t)
	test_error_stamp("1s5h", t)
	test_error_stamp("dsajflkdjflkjdklfjldsjf", t)
	test_error_stamp("1000000sh", t)
	test_error_stamp("s100", t)
	test_error_stamp("1000000mh", t)
	test_error_stamp("100s500", t)
	test_error_stamp("###100j", t)
	test_error_stamp("Forever!", t)
	test_error_stamp("Foorever", t)
	test_error_stamp("forever?", t)
	test_error_stamp("?forever", t)
	test_error_stamp("Now!", t)
	test_error_stamp("Noow", t)
	test_error_stamp("now?", t)
	test_error_stamp("?now", t)
}

func gen_day(gen *rand.Rand) (stamp string, duration time.Duration) {
	day := gen.Intn(10000)
	return fmt.Sprintf("%dd", day), time.Duration(time.Hour * 24 * time.Duration(day))
}

func gen_hour(gen *rand.Rand) (stamp string, duration time.Duration) {
	hour := gen.Intn(10000)
	return fmt.Sprintf("%dh", hour), time.Duration(time.Hour * time.Duration(hour))
}

func gen_min(gen *rand.Rand) (stamp string, duration time.Duration) {
	min := gen.Intn(10000)
	return fmt.Sprintf("%dm", min), time.Duration(time.Minute * time.Duration(min))
}

func gen_sec(gen *rand.Rand) (stamp string, duration time.Duration) {
	sec := gen.Intn(10000)
	return fmt.Sprintf("%ds", sec), time.Duration(time.Second * time.Duration(sec))
}

func test_stamp(stamp string, corrTime time.Time, duration bool, t *testing.T) {
	result, isDuration, err := ParseTime(stamp)
	if err != nil {
		t.Errorf("unexpected error from stamp '%s': %v", stamp, err)
		return
	}

	if isDuration != duration {
		t.Errorf("expected is duration=%t got=%t (%s)", duration, isDuration, stamp)
	}

	// 0.1 Seconds, reasonable error time for computation
	if math.Abs(float64(result.Sub(corrTime).Nanoseconds())) > (0.1 * float64(time.Duration(time.Second))) {
		t.Errorf("time didn't match expected. exp=%+v got=%+v diff=%+v", corrTime, result, corrTime.Sub(result))
	}

	return
}

func test_error_stamp(stamp string, t *testing.T) {
	_, _, err := ParseTime(stamp)
	if err == nil {
		t.Errorf("expected err, got none: %s", stamp)
	}
}
