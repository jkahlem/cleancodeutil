package messages

import (
	"encoding/json"
	"io"

	"returntypes-langserver/common/debug/errors"
)

// Reads/Writes json values as messages from/to a stream.
type JsonStreamReadWriter struct {
	readWriter io.ReadWriter
	decoder    *json.Decoder
}

// Creates a new json readwriter.
func NewJson(rw io.ReadWriter) *JsonStreamReadWriter {
	return &JsonStreamReadWriter{
		readWriter: rw,
		decoder:    json.NewDecoder(rw),
	}
}

// Reads a json message from the stream.
func (rw *JsonStreamReadWriter) ReadMessage() (string, errors.Error) {
	if rw.readWriter != nil {
		var obj interface{}
		rw.decoder.Decode(&obj)
		out, err := json.Marshal(obj)
		if err != nil {
			return "", errors.Wrap(err, "Error", "Could not parse json object")
		} else {
			return string(out), nil
		}
	}
	return "", nil
}

// Writes a message to the stream.
func (rw *JsonStreamReadWriter) WriteMessage(content []byte) errors.Error {
	if rw.readWriter != nil {
		if _, err := rw.readWriter.Write(content); err != nil {
			return errors.Wrap(err, "Error", "Could not write message")
		}
		return nil
	}
	return nil
}

func (r *JsonStreamReadWriter) Reset() {
	// do nothing
	return
}
