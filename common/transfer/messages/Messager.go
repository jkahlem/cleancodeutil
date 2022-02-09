package messages

import "returntypes-langserver/common/debug/errors"

type Messager interface {
	// Reads a message
	ReadMessage() (string, errors.Error)
	// Writes a message
	WriteMessage(content []byte) errors.Error
	// Resets the messager
	Reset()
}
