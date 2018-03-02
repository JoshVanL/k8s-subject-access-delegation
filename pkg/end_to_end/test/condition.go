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
	Line  int    `yaml:"Line"`
	Split int    `yaml:"Split"`
	Match string `yaml:"Match:`
}

type StringCondition struct {
	Line  int    `yaml:"Line"`
	Match string `yaml:"Match"`
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
		return gotNothing(s.Match)
	}

	words := strings.Fields(stdout[s.Line])
	if len(words) < s.Split+1 {
		return gotNothing(s.Match)
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
		return gotNothing(s.Match)
	}
	return fmt.Sprintf("expected='%s', got='%s'", s.Match, stdout[s.Line])
}

func gotNothing(expected string) string {
	return fmt.Sprintf("expected='%s', got=nothing", expected)
}
