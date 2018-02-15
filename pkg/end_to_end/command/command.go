package command

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Command struct {
	stdinWriter  io.WriteCloser
	stderrReader io.ReadCloser
	stdoutReader io.ReadCloser

	scanStdout *bufio.Scanner
	scanStderr *bufio.Scanner

	name string
	args []string
	cmd  *exec.Cmd

	stdoutBuffer chan string
	stderrBuffer chan string
	complete     chan struct{}
}

func New(name string, args []string) (command *Command, err error) {
	var result *multierror.Error

	command = &Command{
		name:         name,
		args:         args,
		cmd:          exec.Command(name, args...),
		stdoutBuffer: make(chan string, 256),
		stderrBuffer: make(chan string, 256),
		complete:     make(chan struct{}),
	}

	if command.stdinWriter, err = command.cmd.StdinPipe(); err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StdinPipe: %v", err))
	}
	if command.stderrReader, err = command.cmd.StderrPipe(); err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StderrPipe: %v", err))
	}
	if command.stdoutReader, err = command.cmd.StdoutPipe(); err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StdoutPipe: %v", err))
	}

	command.scanStdout = bufio.NewScanner(command.stdoutReader)
	command.scanStdout.Split(bufio.ScanLines)
	command.scanStderr = bufio.NewScanner(command.stderrReader)
	command.scanStderr.Split(bufio.ScanLines)

	return command, result.ErrorOrNil()
}

func (c *Command) Run() error {
	var result *multierror.Error
	var stdoutError error
	var stderrError error
	var wg sync.WaitGroup

	if err := c.Start(); err != nil {
		return err
	}

	wg.Add(2)

	go func() {
		if err := c.stdoutScan(); err != nil {
			stdoutError = err
		}
		close(c.stdoutBuffer)
		wg.Done()
	}()

	go func() {
		if err := c.stderrScan(); err != nil {
			stderrError = err
		}
		wg.Done()
		close(c.stderrBuffer)
	}()

	wg.Wait()

	result = multierror.Append(result, stdoutError, stderrError)

	if err := c.cmd.Wait(); err != nil {
		result = multierror.Append(result, err)
	}

	close(c.complete)

	return result.ErrorOrNil()
}

func (c *Command) Start() error {
	return c.cmd.Start()
}

func (c *Command) Wait() {
	<-c.complete
}

func (c *Command) ReReady() {
	c.complete = make(chan struct{})
	c.stdoutBuffer = make(chan string, 256)
	c.stderrBuffer = make(chan string, 256)
}

func (c *Command) Stdout() chan string {
	return c.stdoutBuffer
}

func (c *Command) Stderr() chan string {
	return c.stderrBuffer
}

func (c *Command) stdoutScan() error {
	var result *multierror.Error

	for c.scanStdout.Scan() {
		fmt.Printf(fmt.Sprintf("%s\n", c.scanStdout.Text()))
	}

	return result.ErrorOrNil()
}

func (c *Command) stderrScan() error {
	var result *multierror.Error

	for c.scanStderr.Scan() {
		fmt.Printf(fmt.Sprintf("%s\n", c.scanStderr.Text()))
	}

	return result.ErrorOrNil()
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}
