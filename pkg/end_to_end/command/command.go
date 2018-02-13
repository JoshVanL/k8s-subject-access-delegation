package command

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
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

	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer
	complete     chan struct{}
}

func New(name string, args []string) (command *Command, err error) {
	var result *multierror.Error

	command = &Command{
		name:         name,
		args:         args,
		cmd:          exec.Command(name, args...),
		stdoutBuffer: &bytes.Buffer{},
		stderrBuffer: &bytes.Buffer{},
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
	command.scanStderr = bufio.NewScanner(command.stderrReader)

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
		wg.Done()
	}()

	go func() {
		if err := c.stderrScan(); err != nil {
			stderrError = err
		}
		wg.Done()
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
}

func (c *Command) Stdout() string {
	return c.stdoutBuffer.String()
}

func (c *Command) Stderr() string {
	return c.stderrBuffer.String()
}

func (c *Command) stdoutScan() error {
	var result *multierror.Error

	for c.scanStdout.Scan() {
		if _, err := c.stdoutBuffer.WriteString(c.scanStdout.Text()); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (c *Command) stderrScan() error {
	var result *multierror.Error

	for c.scanStderr.Scan() {
		if _, err := c.stderrBuffer.Write(c.scanStderr.Bytes()); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
