package end_to_end

import (
	"fmt"
	"sync"

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

	commands []*command.Command
}

type CommandArguments struct {
	program   string
	arguments []string
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
		}
	}

	return result.ErrorOrNil()
}

func (t *TestingSuite) allTestBlocks() (blocks []*TestBlock, err error) {
	var result *multierror.Error

	block, err := initialStartup()
	if err != nil {
		result = multierror.Append(result, err)
	}
	blocks = append(blocks, block)

	block, err = cleanUp()
	if err != nil {
		result = multierror.Append(result, err)
	}
	blocks = append(blocks, block)

	return blocks, result.ErrorOrNil()
}

func NewTestBlock(name string, programs []*CommandArguments) (test *TestBlock, err error) {
	var result *multierror.Error

	test = &TestBlock{
		name: name,
	}

	for _, program := range programs {
		cmd, err := command.New(program.program, program.arguments)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}
		test.commands = append(test.commands, cmd)
	}

	return test, result.ErrorOrNil()
}

func (t *TestingSuite) run(block *TestBlock) error {
	var result *multierror.Error
	var wg sync.WaitGroup

	t.log.Infof("-- Testing block: %s --", block.name)
	for _, cmd := range block.commands {
		t.log.Infof("Running command: $ %s", cmd.String())

		wg.Add(3)

		go func() {
			if err := cmd.Run(); err != nil {
				result = multierror.Append(result, err)
			}
			wg.Done()
		}()

		go func() {
			for stdout := range cmd.Stdout() {
				fmt.Printf(stdout)
			}
			wg.Done()
		}()

		go func() {
			for stderr := range cmd.Stderr() {
				fmt.Printf(stderr)
			}
			wg.Done()
		}()

		wg.Wait()
	}

	return result.ErrorOrNil()
}
