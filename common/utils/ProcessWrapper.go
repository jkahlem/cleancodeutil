package utils

import (
	"fmt"
	"io"
	"os/exec"
	"returntypes-langserver/common/debug/errors"
)

const ProcessErrorTitle = "Process Error"

type Process struct {
	cmd *exec.Cmd
}

// Wraps the exec.Command functionalities of the official go lib while converting errors to the stacktraceable errors.
func NewProcess(name string, args ...string) *Process {
	g := &Process{}
	g.cmd = exec.Command(name, args...)
	str := g.cmd.String()
	fmt.Println(str)
	return g
}

// Returns a reader for the stdout pipe
func (g *Process) Stdout() (io.ReadCloser, errors.Error) {
	pipe, err := g.cmd.StdoutPipe()
	if err != nil {
		return pipe, errors.Wrap(err, ProcessErrorTitle, "Could not create io pipes")
	}
	return pipe, nil
}

// Returns a writer for the stdin pipe
func (g *Process) Stdin() (io.WriteCloser, errors.Error) {
	pipe, err := g.cmd.StdinPipe()
	if err != nil {
		return pipe, errors.Wrap(err, ProcessErrorTitle, "Could not create io pipes")
	}
	return pipe, nil
}

// Returns a reader for the stderr pipe
func (p *Process) Stderr() (io.ReadCloser, errors.Error) {
	pipe, err := p.cmd.StderrPipe()
	if err != nil {
		return pipe, errors.Wrap(err, ProcessErrorTitle, "Could not create io pipes")
	}
	return pipe, nil
}

// Starts the process similar to command.Start without waiting for the process to end.
func (p *Process) Start() errors.Error {
	if err := p.cmd.Start(); err != nil {
		return errors.Wrap(err, ProcessErrorTitle, "An error occured while running the process")
	}
	return nil
}

// Wait for the process in the current thread.
func (p *Process) Wait() errors.Error {
	if err := p.cmd.Wait(); err != nil {
		return errors.Wrap(err, ProcessErrorTitle, "An error occured while running the process")
	}
	return nil
}

// Closes the process.
func (p *Process) Close() errors.Error {
	if p.cmd != nil && p.cmd.Process != nil && !p.cmd.ProcessState.Exited() {
		if err := p.cmd.Process.Kill(); err != nil {
			return errors.Wrap(err, ProcessErrorTitle, "Could not stop the process")
		}
	}
	return nil
}

// Returns true if the process is still running.
func (p *Process) IsRunning() bool {
	return p.cmd != nil && (p.cmd.ProcessState == nil || !p.cmd.ProcessState.Exited())
}
