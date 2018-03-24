package test

import (
	"testing"
)

func Test_SplitStringCondition(t *testing.T) {
	s := &SplitStringCondition{
		Line:  1,
		Split: 1,
		Match: "hello!",
	}

	matches := [][]string{[]string{"", "a hello!"}, []string{"foo bar car", "foo hello! foo bar ", "", "foo", "bar"}}
	for _, match := range matches {
		if !s.TestConditon(match) {
			t.Errorf("expected to pass, did not: %s", match)
		}
	}

	nonMatches := [][]string{[]string{"", "", "a hello!"}, []string{"foo bar car", "foo hello foo bar ", "", "foo", "bar"}, []string{""}}
	for _, match := range nonMatches {
		if s.TestConditon(match) {
			t.Errorf("expected to fail, did not: %s", match)
		}
	}
}

func Test_StringCondition(t *testing.T) {
	s := &StringCondition{
		Line:  1,
		Match: "hello!",
	}

	matches := [][]string{[]string{"", "hello!"}, []string{"foo bar car", "hello!", "", "foo", "bar"}}
	for _, match := range matches {
		if !s.TestConditon(match) {
			t.Errorf("expected to pass, did not: %s", match)
		}
	}

	nonMatches := [][]string{[]string{"", "", "hello!"}, []string{"foo bar car", "foo hello foo bar ", "", "foo", "bar"}, []string{""}}
	for _, match := range nonMatches {
		if s.TestConditon(match) {
			t.Errorf("expected to fail, did not: %s", match)
		}
	}

}
