package end_to_end

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"github.com/hashicorp/go-multierror"
)

type Command struct {
	Stdin  io.WriteCloser
	Stderr io.ReadCloser
	Stdout io.ReadCloser

	ScanStdout *bufio.Scanner
	ScanStderr *bufio.Scanner

	name string
	args []string
	cmd  *exec.Cmd
}

func NewCommand(name string, args []string) (command *Command, err error) {
	var result *multierror.Error

	command = &Command{
		name: name,
		args: args,
		cmd:  exec.Command(name, args...),
	}

	if command.Stdin, err = command.cmd.StdinPipe(); err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StdinPipe: %v", err))
	}
	if command.Stderr, err = command.cmd.StderrPipe(); err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StderrPipe: %v", err))
	}
	if command.Stdout, err = command.cmd.StdoutPipe(); err != nil {
		result = multierror.Append(result, fmt.Errorf("unable to get command StdoutPipe: %v", err))
	}

	command.ScanStdout = bufio.NewScanner(command.Stdout)
	command.ScanStdout.Split(bufio.ScanWords)
	command.ScanStderr = bufio.NewScanner(command.Stderr)
	command.ScanStderr.Split(bufio.ScanWords)

	return command, result.ErrorOrNil()
}

func (c *Command) Start() error {
	return c.cmd.Start()
}

func (c *Command) Wait() error {
	return c.cmd.Wait()
}

func (c *Command) StdoutScan() (stdout string, err error) {
	var result *multierror.Error
	var buf bytes.Buffer

	for c.ScanStdout.Scan() {
		if _, err := buf.WriteString(c.ScanStdout.Text()); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return buf.String(), result.ErrorOrNil()
}

func (c *Command) StderrScan() (stderr string, err error) {
	var result *multierror.Error
	var buf bytes.Buffer

	for c.ScanStderr.Scan() {
		if _, err := buf.WriteString(c.ScanStderr.Text()); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return buf.String(), result.ErrorOrNil()
}
