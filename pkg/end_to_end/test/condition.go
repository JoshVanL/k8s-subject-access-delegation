package test

import (
	"strings"
)

var _ Condition = &SplitStringCondition{}

type Condition interface {
	TestConditon(stdout []string) (pass bool)
}

type SplitStringCondition struct {
	line  int
	split int
	match string
}

func (s *SplitStringCondition) TestConditon(stdout []string) (pass bool) {
	words := strings.Fields(stdout[s.line])
	return words[s.split] == s.match
}
