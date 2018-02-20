package test

import (
	"fmt"
	"strings"
)

var _ Condition = &SplitStringCondition{}
var _ Condition = &StringCondition{}

type Condition interface {
	TestConditon(stdout []string) (pass bool)
	Expected(stdout []string) string
}

type SplitStringCondition struct {
	line  int
	split int
	match string
}

type StringCondition struct {
	line  int
	match string
}

func (s *SplitStringCondition) TestConditon(stdout []string) (pass bool) {
	if len(stdout) < s.line+1 {
		return false
	}

	words := strings.Fields(stdout[s.line])
	if len(words) < s.split+1 {
		return false
	}

	return words[s.split] == s.match
}

func (s *SplitStringCondition) Expected(stdout []string) string {
	if len(stdout) < s.line+1 {
		return gotNothing(s.match)
	}

	words := strings.Fields(stdout[s.line])
	if len(words) < s.split+1 {
		return gotNothing(s.match)
	}
	return fmt.Sprintf("expected='%s', got='%s'", s.match, words[s.split])
}

func (s *StringCondition) TestConditon(stdout []string) (pass bool) {
	if len(stdout) < s.line+1 {
		return false
	}
	return stdout[s.line] == s.match
}

func (s *StringCondition) Expected(stdout []string) string {
	if len(stdout) < s.line+1 {
		return gotNothing(s.match)
	}
	return fmt.Sprintf("expected='%s', got='%s'", s.match, stdout[s.line])
}

func gotNothing(expected string) string {
	return fmt.Sprintf("expected='%s', got=nothing", expected)
}
