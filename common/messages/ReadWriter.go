// Gives support for sending/receiving messages according to the Base Protocol of LSP
// For simplification, this base protocol is re-used for the predictor
package messages

import (
	"fmt"
	"io"
	"sync"

	"returntypes-langserver/common/errors"
	"returntypes-langserver/common/log"
	"returntypes-langserver/common/utils"
)

// A read writer for messages of the LSP base protocol.
type ReadWriter struct {
	acceptedMediaTypes []string
	writingMIMEType    string
	readWriter         *utils.BufferedReadWriter
	mutex              sync.Mutex
}

// Creates a new read writer for messages with the default mime type "text/plain".
func NewReadWriter(readWriter io.ReadWriter) *ReadWriter {
	bufferedReadWriter := utils.NewBufferedReadWriter(readWriter, readWriter)
	return &ReadWriter{readWriter: bufferedReadWriter, writingMIMEType: "text/plain"}
}

// Accepts messages with the given mime type.
// By default, accepts any mime type, so if this function is called, only the mime types set using this function are accepted.
func (r *ReadWriter) AcceptMediaType(mediaType string) {
	r.acceptedMediaTypes = append(r.acceptedMediaTypes, mediaType)
}

// Sets the mime type used for writing messages.
func (r *ReadWriter) SetWritingMimeType(mimeType string) {
	r.writingMIMEType = mimeType
}

// Reads a message from the underlying stream.
func (r *ReadWriter) ReadMessage() (string, errors.Error) {
	msg, err := r.readMessage()
	log.Print(log.Messager, "<< Read Message:\n%s\n", msg.String())
	return msg.Body, err
}

// Reads and parses a message from the underlying stream.
func (r *ReadWriter) readMessage() (MessageWrapper, errors.Error) {
	header, err := r.readHeader()
	if err != nil {
		return MessageWrapper{}, err
	} else if !r.isMessageAccepted(header) {
		return MessageWrapper{}, errors.New("Error", fmt.Sprintf("The mime type %s is not accepted", header.ContentType))
	}
	body, err := r.readBody(header)
	if err != nil {
		return MessageWrapper{}, err
	}
	return MessageWrapper{
		Header: header,
		Body:   body,
	}, nil
}

// Returns true if the message is accepted.
func (r *ReadWriter) isMessageAccepted(header Header) bool {
	if len(header.ContentType.mediaType) == 0 {
		// default to the accepted media types
		return true
	}
	for _, acceptedMediaType := range r.acceptedMediaTypes {
		if acceptedMediaType == header.ContentType.mediaType {
			return true
		}
	}
	return false
}

// Reads and parses the message header from the stream.
func (r *ReadWriter) readHeader() (Header, errors.Error) {
	header := Header{}
	for {
		line, err := r.readWriter.ReadString('\n')
		if err != nil {
			return header, errors.Wrap(err, "Error", "Could not read line of message")
		} else if line == HeaderSeperator {
			return header, nil
		}
		line = line[:len(line)-2]
		if err = header.ParseSettingFromLine(line); err != nil {
			return header, errors.Wrap(err, "Error", "Could not read line of message")
		}
	}
}

// Reads and parses the message body from the stream using the content length property.
func (r *ReadWriter) readBody(header Header) (string, errors.Error) {
	bytes := make([]byte, header.ContentLength)
	readBytes, err := io.ReadFull(r.readWriter, bytes)
	if err != nil {
		return "", errors.Wrap(err, "Error", "Could not read message body")
	}
	return string(bytes[:readBytes]), nil
}

// Writes a message with the given content using the setted mime type to the underlying stream.
func (r *ReadWriter) WriteMessage(content []byte) errors.Error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	msg, err := r.createEmptyMessage(r.writingMIMEType)
	if err != nil {
		return err
	}

	_, err = msg.Write(content)
	if err != nil {
		return err
	}

	if _, err := r.readWriter.Write(msg.Bytes()); err != nil {
		return errors.Wrap(err, "Error", "Could not write message")
	} else if err := r.readWriter.Flush(); err != nil {
		return errors.Wrap(err, "Error", "Could not write message")
	}
	log.Print(log.Messager, ">> Write Message:\n%s\n", msg.String())
	return nil
}

// Creates an empty message.
func (r *ReadWriter) createEmptyMessage(mimeType string) (MessageWrapper, errors.Error) {
	header, err := NewHeader(mimeType)
	if err != nil {
		return MessageWrapper{}, err
	}

	return MessageWrapper{
		Header: header,
	}, nil
}

func (r *ReadWriter) Reset() {
	r.readWriter.Reset()
}
