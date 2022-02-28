package external

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/debug/log"
)

const GitErrorTitle string = "Git Error"

// Wraps functionalities for running the crawler process and interacting with it.
type gitProcess struct {
	cmd *exec.Cmd
}

// Creates and prepares a new crawler process wrapper without starting it yet.
func NewProcess(options Options) *gitProcess {
	c := &gitProcess{}
	c.cmd = exec.Command("git", optionsToArgs(options)...)
	return c
}

func optionsToArgs(options Options) []string {
	args := make([]string, 0, 4)
	args = append(args, "clone")
	if options.Filter != nil {
		args = append(args, buildFilter(*options.Filter))
	}
	args = append(args, options.URI)
	args = append(args, options.OutputDir)
	return args
}

func buildFilter(filter Filter) string {
	if len(filter.SizeLimit) > 0 {
		return fmt.Sprintf("--filter=blob:limit=%s", filter.SizeLimit)
	}
	return ""
}

func (c *gitProcess) Stdout() (io.ReadCloser, errors.Error) {
	pipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return pipe, errors.Wrap(err, GitErrorTitle, "Could not create io pipes")
	}
	return pipe, nil
}

func (c *gitProcess) Stdin() (io.WriteCloser, errors.Error) {
	pipe, err := c.cmd.StdinPipe()
	if err != nil {
		return pipe, errors.Wrap(err, GitErrorTitle, "Could not create io pipes")
	}
	return pipe, nil
}

// Starts the crawler similar to command.Start without waiting for the process to end.
func (c *gitProcess) Start() errors.Error {
	if err := c.cmd.Start(); err != nil {
		return errors.Wrap(err, GitErrorTitle, "An error occured while running the crawler")
	}
	return nil
}

// Wait for the crawler in the current thread.
func (c *gitProcess) Wait() errors.Error {
	if err := c.cmd.Wait(); err != nil {
		return errors.Wrap(err, GitErrorTitle, "An error occured while running the crawler")
	}
	return nil
}

// Closes the crawler process.
func (c *gitProcess) Close() errors.Error {
	if c.cmd != nil && c.cmd.Process != nil && !c.cmd.ProcessState.Exited() {
		if err := c.cmd.Process.Kill(); err != nil {
			return errors.Wrap(err, GitErrorTitle, "Could not stop the crawler process")
		}
	}
	return nil
}

// Returns true if the crawler process is still running.
func (c *gitProcess) IsRunning() bool {
	return c.cmd != nil && (c.cmd.ProcessState == nil || !c.cmd.ProcessState.Exited())
}

type Git struct {
	process *gitProcess
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

type Options struct {
	OutputDir string
	URI       string
	Filter    *Filter
}

type Filter struct {
	SizeLimit string
}

func (g *Git) Clone(options Options) errors.Error {
	return g.runProcess(options)
}

func (g *Git) runProcess(options Options) errors.Error {
	g.process = NewProcess(options)
	stdin, err := g.process.Stdin()
	if err != nil {
		return err
	}
	stdout, err := g.process.Stdout()
	if err != nil {
		return err
	}
	g.stdin = stdin
	g.stdout = stdout

	if err := g.process.Start(); err != nil {
		g.Close()
		return err
	}
	go func() {
		reader := bufio.NewReader(g.stdout)
		var err2 error
		var str string
		for str, err2 = reader.ReadString('\n'); len(str) > 0 || err2 == nil; {
			log.Info(str)
		}
		if err2 != nil {
			log.Error(errors.Wrap(err2, "Error", "Error"))
		}
	}()
	err = g.process.Wait()
	g.Close()
	return err
}

// Returns true if a connection to the crawler is present (so the crawler process is still running).
func (conn *Git) IsConnected() bool {
	return conn.process != nil && conn.process.IsRunning()
}

// Reads bytes from the standard input stream of the crawler process.
func (conn *Git) Read(bytes []byte) (int, error) {
	if conn.stdout == nil {
		return 0, errors.Wrap(io.ErrClosedPipe, GitErrorTitle, "Stream does not exist")
	}
	n, err := conn.stdout.Read(bytes)
	return n, err
}

// Writes bytes to the standard output stream of the crawler process.
func (conn *Git) Write(bytes []byte) (int, error) {
	if conn.stdout == nil {
		return 0, errors.Wrap(io.ErrClosedPipe, GitErrorTitle, "Stream does not exist")
	}
	return conn.stdin.Write(bytes)
}

// Closes the crawler connection.
func (conn *Git) Close() errors.Error {
	if conn.stdin != nil {
		if err := conn.stdin.Close(); err != nil {
			return errors.Wrap(err, GitErrorTitle, "Could not close stream")
		}
		conn.stdin = nil
	}
	if conn.stdout != nil {
		if err := conn.stdout.Close(); err != nil {
			return errors.Wrap(err, GitErrorTitle, "Could not close stream")
		}
		conn.stdout = nil
	}
	if conn.process != nil && conn.process.IsRunning() {
		if err := conn.process.Close(); err != nil {
			return err
		}
	}
	return nil
}
