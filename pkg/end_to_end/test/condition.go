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
	Line  int    `yaml:"line"`
	Split int    `yaml:"split"`
	Match string `yaml:"match:`
}

type StringCondition struct {
	Line  int    `yaml:"line"`
	Match string `yaml:"match"`
}

func (s *SplitStringCondition) TestConditon(stdout []string) (pass bool) {
	if len(stdout) < s.Line+1 {
		return false
	}

	words := strings.Fields(stdout[s.Line])
	if len(words) < s.Split+1 {
		return false
	}

	return words[s.Split] == s.Match
}

func (s *SplitStringCondition) Expected(stdout []string) string {
	if len(stdout) < s.Line+1 {
		return gotNothing(s.Match, stdout)
	}

	words := strings.Fields(stdout[s.Line])
	if len(words) < s.Split+1 {
		return gotNothing(s.Match, stdout)
	}

	return fmt.Sprintf("expected='%s', got='%s'", s.Match, words[s.Split])
}

func (s *StringCondition) TestConditon(stdout []string) (pass bool) {
	if len(stdout) < s.Line+1 {
		return false
	}
	return stdout[s.Line] == s.Match
}

func (s *StringCondition) Expected(stdout []string) string {
	if len(stdout) < s.Line+1 {
		return gotNothing(s.Match, stdout)
	}
	return fmt.Sprintf("expected='%s', got='%s'", s.Match, stdout[s.Line])
}

func gotNothing(expected string, stdout []string) string {
	return fmt.Sprintf("expected='%s', got(all)=\n%s", expected, strings.Join(stdout, "\n"))
}
