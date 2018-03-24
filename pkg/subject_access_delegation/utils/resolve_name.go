package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	valid = regexp.MustCompile(`^([0-9a-zA-Z]|\*)+((\.|\-|\*|[0-9a-zA-Z])*([0-9a-zA-Z]|\*)+)*$`)
)

func ValidName(name string) bool {
	return valid.MatchString(name)
}

func MatchName(name, regex string) (bool, error) {

	if regex == "*" {
		return true, nil
	}

	regex = format(regex)

	if !ValidName(name) {
		return false, fmt.Errorf("not a valid name '%s'. Must contain only alphanumerics, '-', '.' and '*'", name)
	}

	r, err := regexp.Compile(regex)
	if err != nil {
		return false, fmt.Errorf("failed to compile regular expression: %v", err)
	}

	if !r.MatchString(name) {
		return false, nil
	}

	return true, nil
}

func format(regex string) string {
	regex = strings.NewReplacer(`.`, `\.`).Replace(regex)
	regex = strings.NewReplacer(`-`, `\-`).Replace(regex)
	regex = strings.NewReplacer(`*`, `(\.|\-|\*|[0-9a-zA-Z])*`).Replace(regex)

	return fmt.Sprintf("^%s$", regex)
}
