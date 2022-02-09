package crawler

import (
	"io"
	"os/exec"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
)

// Wraps functionalities for running the crawler process and interacting with it.
type crawlerProcess struct {
	cmd *exec.Cmd
}

// Creates and prepares a new crawler process wrapper without starting it yet.
func NewProcess() *crawlerProcess {
	c := &crawlerProcess{}
	c.cmd = exec.Command("java", "-jar", configuration.CrawlerPath())
	return c
}

func (c *crawlerProcess) Stdout() (io.ReadCloser, errors.Error) {
	pipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return pipe, errors.Wrap(err, CrawlerErrorTitle, "Could not create io pipes")
	}
	return pipe, nil
}

func (c *crawlerProcess) Stdin() (io.WriteCloser, errors.Error) {
	pipe, err := c.cmd.StdinPipe()
	if err != nil {
		return pipe, errors.Wrap(err, CrawlerErrorTitle, "Could not create io pipes")
	}
	return pipe, nil
}

// Starts the crawler similar to command.Start without waiting for the process to end.
func (c *crawlerProcess) Start() errors.Error {
	if err := c.cmd.Start(); err != nil {
		return errors.Wrap(err, CrawlerErrorTitle, "An error occured while running the crawler")
	}
	return nil
}

// Wait for the crawler in the current thread.
func (c *crawlerProcess) Wait() errors.Error {
	if err := c.cmd.Wait(); err != nil {
		return errors.Wrap(err, CrawlerErrorTitle, "An error occured while running the crawler")
	}
	return nil
}

// Closes the crawler process.
func (c *crawlerProcess) Close() errors.Error {
	if c.cmd != nil && c.cmd.Process != nil && !c.cmd.ProcessState.Exited() {
		if err := c.cmd.Process.Kill(); err != nil {
			return errors.Wrap(err, CrawlerErrorTitle, "Could not stop the crawler process")
		}
	}
	return nil
}

// Returns true if the crawler process is still running.
func (c *crawlerProcess) IsRunning() bool {
	return c.cmd != nil && (c.cmd.ProcessState == nil || !c.cmd.ProcessState.Exited())
}
