package crawler

import (
	"io"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
)

// Defines a connection to the crawler. Uses standard-IO of the crawler process for transmission.
type connection struct {
	process *utils.Process
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

// "Connects" to the crawler by spawning a new crawler process.
func (conn *connection) Connect() errors.Error {
	if conn.process != nil {
		return errors.New(CrawlerErrorTitle, "A connection does already exist")
	} else if err := conn.connect(); err != nil {
		return err
	}
	return nil
}

func (conn *connection) connect() errors.Error {
	conn.process = utils.NewProcess("java", "-jar", configuration.CrawlerPath())
	stdin, err := conn.process.Stdin()
	if err != nil {
		return err
	}
	stdout, err := conn.process.Stdout()
	if err != nil {
		return err
	}
	conn.stdin = stdin
	conn.stdout = stdout

	if err := conn.process.Start(); err != nil {
		conn.Close()
		return err
	}
	go func() {
		if err := conn.process.Wait(); err != nil {
			conn.Close()
		}
	}()
	return nil
}

// Returns true if a connection to the crawler is present (so the crawler process is still running).
func (conn *connection) IsConnected() bool {
	return conn.process != nil && conn.process.IsRunning()
}

// Reads bytes from the standard input stream of the crawler process.
func (conn *connection) Read(bytes []byte) (int, error) {
	if conn.stdout == nil {
		return 0, errors.Wrap(io.ErrClosedPipe, CrawlerErrorTitle, "Stream does not exist")
	}
	n, err := conn.stdout.Read(bytes)
	return n, err
}

// Writes bytes to the standard output stream of the crawler process.
func (conn *connection) Write(bytes []byte) (int, error) {
	if conn.stdout == nil {
		return 0, errors.Wrap(io.ErrClosedPipe, CrawlerErrorTitle, "Stream does not exist")
	}
	return conn.stdin.Write(bytes)
}

// Closes the crawler connection.
func (conn *connection) Close() errors.Error {
	if conn.stdin != nil {
		if err := conn.stdin.Close(); err != nil {
			return errors.Wrap(err, CrawlerErrorTitle, "Could not close stream")
		}
		conn.stdin = nil
	}
	if conn.stdout != nil {
		if err := conn.stdout.Close(); err != nil {
			return errors.Wrap(err, CrawlerErrorTitle, "Could not close stream")
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

// Returns true as the crawler connection is recoverable (by respawning the crawler process).
func (conn *connection) IsRecoverable() bool {
	return true
}
