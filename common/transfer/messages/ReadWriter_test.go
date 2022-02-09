package messages

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"returntypes-langserver/common/utils"

	"github.com/stretchr/testify/assert"
)

func TestMessageParsing(t *testing.T) {
	// given
	strRw := createStringReadWriterWithInput(TestRequest)
	readWriter := NewReadWriter(strRw)
	readWriter.AcceptMediaType("text/plain")

	// when
	msgWrapper, err := readWriter.readMessage()

	// then
	assert.NoError(t, err)
	assert.Equal(t, "text/plain", msgWrapper.Header.ContentType.mediaType)
	assert.Equal(t, 8, msgWrapper.Header.ContentLength)
	assert.Equal(t, "12345678", msgWrapper.Body)
}

func TestMessageParsingWithNotAcceptedMimeType(t *testing.T) {
	// given
	strRw := createStringReadWriterWithInput(TestRequest)
	readWriter := NewReadWriter(strRw)
	readWriter.AcceptMediaType("application/json")

	// when
	_, err := readWriter.ReadMessage()

	// then
	assert.Error(t, err)
}

func TestMessageSending(t *testing.T) {
	// given
	strRw, strBuilder := createStringReadWriterWithOutput()
	readWriter := NewReadWriter(strRw)
	readWriter.SetWritingMimeType("application/json")
	expectedMessage := "Content-Length: 14\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n" +
		"{\"name\":\"asd\"}"
	structAsJson, _ := json.Marshal(testStruct{
		Name: "asd",
	})

	// when
	err := readWriter.WriteMessage(structAsJson)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, strBuilder.String())
}

// Test relevant structures

type testStruct struct {
	Name string `json:"name"`
}

// Helpers

const TestRequest = "Content-Length: 8\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" +
	"12345678"

func createStringReadWriterWithInput(input string) io.ReadWriter {
	strReader := strings.NewReader(input)
	strWriter := &strings.Builder{}
	return utils.WrapReadWriter(strReader, strWriter)
}

func createStringReadWriterWithOutput() (io.ReadWriter, *strings.Builder) {
	strReader := strings.NewReader("")
	strWriter := &strings.Builder{}
	return utils.WrapReadWriter(strReader, strWriter), strWriter
}
