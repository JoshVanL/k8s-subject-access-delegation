package test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/end_to_end/command"
)

type TestingSuite struct {
	log *logrus.Entry

	Blocks []TestBlock `yaml:"testing_blocks"`
}

type TestBlock struct {
	Name string `yaml:"name"`

	Commands []Command `yaml:"commands"`
}

type Command struct {
	Command   *command.Command
	Program   string `yaml:"program"`
	Arguments string `yaml:"arguments"`

	Background  bool                   `yaml:"background"`
	Delay       int                    `yaml:"delay"`
	SplitString []SplitStringCondition `yaml:"split_string_conditions"`
	String      []StringCondition      `yaml:"string_conditions"`

	stdout []string
}

func NewSuite(filepath string, log *logrus.Entry) (suite *TestingSuite, err error) {
	suite, err = readTestFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading test file failed: %v", err)
	}

	suite.log = log

	return suite, nil
}

func readTestFile(filename string) (*TestingSuite, error) {
	var result *multierror.Error

	path, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to find absolute file path: %v", err)
	}

	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading test file: %v", err)
	}

	suite := new(TestingSuite)
	if err := yaml.UnmarshalStrict(f, suite); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test file: %v", err)
	}

	for _, block := range suite.Blocks {
		for i, cmd := range block.Commands {
			c, err := command.New(cmd.Program, cmd.Arguments)
			if err != nil {
				result = multierror.Append(result, err)
				continue
			}
			block.Commands[i].Command = c
		}
	}

	return suite, result.ErrorOrNil()
}

func (t *TestingSuite) RunTests() error {
	var result *multierror.Error

	for _, block := range t.Blocks {
		if err := t.run(&block); err != nil {
			result = multierror.Append(result, err)
			t.log.Errorf("Testing block '%s' failed: %v", block.Name, err)
		}
	}

	return result.ErrorOrNil()
}

func (t *TestingSuite) run(block *TestBlock) error {
	var result *multierror.Error
	var backgroundCmds []*command.Command

	t.log.Infof("== Testing block: %s ==", block.Name)
	for _, cmd := range block.Commands {
		t.log.Infof("Running command: $ %s", cmd.Command.String())

		var wg sync.WaitGroup

		wg.Add(3)

		time.Sleep(time.Second * time.Duration(cmd.Delay))

		go func() {
			if err := cmd.Command.Run(); err != nil {
				result = multierror.Append(result, err)
			}
			if cmd.Background {
				backgroundCmds = append(backgroundCmds, cmd.Command)
			}

			wg.Done()
		}()

		go func() {
			for str := range cmd.Command.Stdout() {
				cmd.stdout = append(cmd.stdout, str)
				fmt.Printf("%s\n", str)
			}
			wg.Done()
		}()

		go func() {
			for str := range cmd.Command.Stderr() {
				cmd.stdout = append(cmd.stdout, str)
				fmt.Printf("%s\n", str)
			}
			wg.Done()
		}()

		if cmd.Background {
			time.Sleep(time.Second * 5)
		} else {
			wg.Wait()
		}

		for _, condition := range cmd.SplitString {
			if condition.TestConditon(cmd.stdout) {
				t.log.Infof("Condition passed.")
			} else {
				err := fmt.Errorf("Condition failed in test block '%s': %s", block.Name, condition.Expected(cmd.stdout))
				result = multierror.Append(result, err)
			}
		}

		for _, condition := range cmd.String {
			if condition.TestConditon(cmd.stdout) {
				t.log.Infof("Condition passed.")
			} else {
				err := fmt.Errorf("Condition failed in test block '%s': %s", block.Name, condition.Expected(cmd.stdout))
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
