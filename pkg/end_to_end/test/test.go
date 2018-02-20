package test

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/end_to_end/command"
)

type TestingSuite struct {
	log *logrus.Entry

	blocks []*TestBlock
}

type TestBlock struct {
	name string

	commands []*Command
}

type Command struct {
	command   *command.Command
	program   string
	arguments []string

	background bool
	delay      int
	conditions []Condition

	stdout []string
}

func NewSuit(log *logrus.Entry) (suite *TestingSuite, err error) {

	suite = &TestingSuite{
		log: log,
	}
	suite.blocks, err = suite.allTestBlocks()

	return suite, err
}

func (t *TestingSuite) RunTests() error {
	var result *multierror.Error

	for _, block := range t.blocks {
		if err := t.run(block); err != nil {
			result = multierror.Append(result, err)
			t.log.Errorf("Testing block '%s' failed: %v", block.name, err)
		}
	}

	return result.ErrorOrNil()
}

func NewTestBlock(name string, tests []*Command) (testBlock *TestBlock, err error) {
	var result *multierror.Error

	block := &TestBlock{
		name: name,
	}

	for _, test := range tests {
		cmd, err := command.New(test.program, test.arguments)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		test.command = cmd

		block.commands = append(block.commands, test)
	}

	return block, result.ErrorOrNil()
}

func (t *TestingSuite) run(block *TestBlock) error {
	var result *multierror.Error
	var backgroundCmds []*command.Command

	t.log.Infof("== Testing block: %s ==", block.name)
	for _, cmd := range block.commands {
		t.log.Infof("Running command: $ %s", cmd.command.String())

		var wg sync.WaitGroup

		wg.Add(3)

		time.Sleep(time.Second * time.Duration(cmd.delay))

		go func() {
			if err := cmd.command.Run(); err != nil {
				result = multierror.Append(result, err)
			}
			if cmd.background {
				backgroundCmds = append(backgroundCmds, cmd.command)
			}

			wg.Done()
		}()

		go func() {
			for str := range cmd.command.Stdout() {
				cmd.stdout = append(cmd.stdout, str)
				fmt.Printf("%s\n", str)
			}
			wg.Done()
		}()

		go func() {
			for str := range cmd.command.Stderr() {
				cmd.stdout = append(cmd.stdout, str)
				fmt.Printf("%s\n", str)
			}
			wg.Done()
		}()

		if cmd.background {
			time.Sleep(time.Second * 5)
		} else {
			wg.Wait()
		}

		for _, condition := range cmd.conditions {
			if condition.TestConditon(cmd.stdout) {
				t.log.Infof("Condition passed.")
			} else {
				err := fmt.Errorf("Condition failed in test block '%s': %s", block.name, condition.Expected(cmd.stdout))
				result = multierror.Append(result, err)
			}
		}
	}

	for _, cmd := range backgroundCmds {
		if err := cmd.Kill(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
