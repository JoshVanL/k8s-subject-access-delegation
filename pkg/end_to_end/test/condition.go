package test

import (
	"strings"
)

var _ Condition = &SplitStringCondition{}
var _ Condition = &StringCondition{}

type Condition interface {
	TestConditon(stdout []string) (pass bool)
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
	words := strings.Fields(stdout[s.line])
	return words[s.split] == s.match
}

func (s *StringCondition) TestConditon(stdout []string) (pass bool) {
	return stdout[s.line] == s.match
}
