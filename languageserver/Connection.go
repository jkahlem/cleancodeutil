package languageserver

import (
	"os"
	"returntypes-langserver/common/debug/errors"
)

// A connection wrapping this application's standard io streams.
type connection struct{}

func (conn *connection) Connect() errors.Error {
	// Do nothing. Only used for implementing Connection interface
	return nil
}

func (conn *connection) IsConnected() bool {
	// Do nothing. Only used for implementing Connection interface
	return true
}

func (conn *connection) Read(bytes []byte) (int, error) {
	return os.Stdin.Read(bytes)
}

func (conn *connection) Write(bytes []byte) (int, error) {
	return os.Stdout.Write(bytes)
}

func (conn *connection) Close() errors.Error {
	// Do nothing. Only used for implementing Connection interface
	return nil
}

func (conn *connection) IsRecoverable() bool {
	return false
}
