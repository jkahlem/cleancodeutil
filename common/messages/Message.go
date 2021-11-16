package messages

import (
	"fmt"
	"mime"
	"strconv"
	"strings"

	"returntypes-langserver/common/errors"
)

const HeaderSeperator = "\r\n"

type MessageWrapper struct {
	Header Header
	Body   string
}

type Header struct {
	ContentLength int
	ContentType   MIMEType
}

type MIMEType struct {
	mediaType string
	params    map[string]string
}

func (mimeType *MIMEType) MediaType() string {
	return mimeType.mediaType
}

func (mimeType *MIMEType) Charset() string {
	if val, ok := mimeType.params["charset"]; ok {
		return val
	}
	return ""
}

func (mimeType MIMEType) String() string {
	format := mime.FormatMediaType(mimeType.mediaType, mimeType.params)
	return format
}

// Creates a message header.
func NewHeader(mimeTypeStr string) (Header, errors.Error) {
	mimeType, err := ParseMIMEType(mimeTypeStr)
	return Header{
		ContentLength: 0,
		ContentType:   mimeType,
	}, err
}

// Parses a mime type of a string.
func ParseMIMEType(str string) (MIMEType, errors.Error) {
	mediaType, params, err := mime.ParseMediaType(str)
	if err != nil {
		return MIMEType{}, errors.Wrap(err, "Error", "Could not parse message header")
	}
	return MIMEType{
		mediaType: mediaType,
		params:    params,
	}, nil
}

// Parses settings of the LSP message header.
func (header *Header) ParseSettingFromLine(line string) errors.Error {
	index := strings.Index(line, ": ")
	if index >= 0 {
		fieldName, value := line[:index], line[index+2:]
		switch fieldName {
		case "Content-Length":
			length, err := strconv.Atoi(value)
			if err != nil {
				return errors.Wrap(err, "Error", "Could not parse message header")
			}
			header.ContentLength = length
		case "Content-Type":
			mimeType, err := ParseMIMEType(value)
			if err != nil {
				return errors.Wrap(err, "Error", "Could not parse message header")
			}
			header.ContentType = mimeType
		}
	} else {
		return errors.New("Error", "Header line is malformed")
	}
	return nil
}

// Returns the header in it's string format according to the LSP specification.
func (header Header) String() string {
	return fmt.Sprintf("Content-Length: %d\r\nContent-Type: %s\r\n%s", header.ContentLength, header.ContentType, HeaderSeperator)
}

// Writes data to the message body while updating the ContentLength field of the header accordingly.
func (wrapper *MessageWrapper) Write(p []byte) (n int, err errors.Error) {
	wrapper.Body += string(p)
	wrapper.Header.ContentLength = len([]byte(wrapper.Body))
	return len(p), nil
}

// Returns the message in it's string format according to the LSP specification.
func (wrapper MessageWrapper) String() string {
	return fmt.Sprintf("%s%s", wrapper.Header, wrapper.Body)
}

func (wrapper MessageWrapper) Bytes() []byte {
	return []byte(wrapper.String())
}
