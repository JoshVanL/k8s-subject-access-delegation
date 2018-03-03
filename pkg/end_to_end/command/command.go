package command

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/hashicorp/go-multierror"
)

type Command struct {
	stderrReader io.ReadCloser
	stdoutReader io.ReadCloser

	scanStdout *bufio.Scanner
	scanStderr *bufio.Scanner

	name string
	args []string
	cmd  *exec.Cmd

	complete chan struct{}
	stdoutCh chan string
	stderrCh chan string
}

func New(name string, args []string) (*Command, error) {
	var result *multierror.Error
	command := &Command{
		name:     name,
		args:     args,
		cmd:      exec.Command(name, args...),
		complete: make(chan struct{}),
		stdoutCh: make(chan string),
		stderrCh: make(chan string),
	}

	errReader, err := command.cmd.StderrPipe()
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StderrPipe: %v", err))
	}
	outReader, err := command.cmd.StdoutPipe()
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StdoutPipe: %v", err))
	}

	command.stderrReader = errReader
	command.stdoutReader = outReader
	command.scanStdout = bufio.NewScanner(command.stdoutReader)
	command.scanStdout.Split(bufio.ScanLines)
	command.scanStderr = bufio.NewScanner(command.stderrReader)
	command.scanStderr.Split(bufio.ScanLines)

	return command, result.ErrorOrNil()
}

func (c *Command) Run() error {
	var result *multierror.Error
	var wg sync.WaitGroup

	if err := c.Start(); err != nil {
		return err
	}

	wg.Add(2)

	go func() {
		for c.scanStdout.Scan() {
			c.stdoutCh <- c.scanStdout.Text()
		}

		close(c.stdoutCh)
		wg.Done()
	}()

	go func() {
		for c.scanStderr.Scan() {
			c.stderrCh <- c.scanStderr.Text()
		}

		wg.Done()
		close(c.stderrCh)
	}()

	wg.Wait()

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
	c.stderrCh = make(chan string)
	c.stdoutCh = make(chan string)
}

func (c *Command) Stdout() chan string {
	return c.stdoutCh
}

func (c *Command) Stderr() chan string {
	return c.stderrCh
}

func (c *Command) Kill() error {
	return c.cmd.Process.Signal(syscall.SIGINT)
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s", c.name, strings.Join(c.args, " "))
}
