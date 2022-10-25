package lang

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
)

type Command struct {
	cmd     *exec.Cmd
	started bool
	outchan chan []byte
	outopen bool
}

func NewCommand(name string, args ...string) (c Command, err error) {
	c.cmd = exec.Command(name, args...)
	return c, c.cmd.Err
}

func (c *Command) Start() (err error) {
	if c.started {
		return errors.New("ng: command already started")
	}

	var outpipe, errpipe io.ReadCloser

	if outpipe, err = c.cmd.StdoutPipe(); err != nil {
		return err
	}
	if errpipe, err = c.cmd.StderrPipe(); err != nil {
		return err
	}

	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.outchan = make(chan []byte)

	stdout := bufio.NewScanner(outpipe)
	stderr := bufio.NewScanner(errpipe)

	go func() {
		for stdout.Scan() {
			c.outchan <- stdout.Bytes()
		}

		for stderr.Scan() {
			c.outchan <- stderr.Bytes()
		}

		close(c.outchan)
	}()

	return nil
}

func (c *Command) Available() bool {
	return c.outopen
}

func (c *Command) Output() []byte {
	var outbytes []byte
	outbytes, c.outopen = <-c.outchan
	if !c.outopen {
		return nil
	}

	return outbytes
}

func (c *Command) WaitFor() error {
	return c.cmd.Wait()
}

func (c *Command) Run() ([]byte, error) {
	return c.cmd.CombinedOutput()
}
