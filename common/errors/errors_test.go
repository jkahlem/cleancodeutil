package errors

import (
	"testing"

	goErrors "github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
)

type testErrorInterface interface {
	Error() string
	TestErrorMessage() string
}

func TestNew(t *testing.T) {
	// given
	const testErrorTitle = "IO Error"
	const testErrorMessage = "File does not exist"

	// when
	err := New(testErrorTitle, testErrorMessage)
	_, isGoError := err.Unwrap().(*goErrors.Error)

	// then
	assertError(t, testErrorTitle, testErrorMessage, err)
	assert.True(t, isGoError)
}

func TestWrap(t *testing.T) {
	// given
	const testErrorTitle = "Parser Error"
	const testErrorMessage = "Could not parse"
	errToWrap := New("IO Error", "File does not exist")

	// when
	wrapperErr := Wrap(errToWrap, testErrorTitle, testErrorMessage)

	// then
	assertError(t, testErrorTitle, testErrorMessage, wrapperErr)
	assert.Equal(t, errToWrap, wrapperErr.Unwrap())
}

func TestAsForStructError(t *testing.T) {
	// given
	structErr := testErrorStruct{msg: "TestErrorStructMsg"}
	wrappedStructErr := Wrap(structErr, "A", "B")

	// when
	structErrDestination := testErrorStruct{}
	isStructErr := As(wrappedStructErr, &structErrDestination)

	// then
	assert.True(t, isStructErr)
	assert.Equal(t, structErr.msg, structErrDestination.msg)
}

func TestAsForPtrError(t *testing.T) {
	// given
	ptrErr := &testErrorPtr{msg: "TestErrorPtrMsg"}
	wrappedPtrErr := Wrap(ptrErr, "A", "B")

	// when
	ptrErrorDestination := &testErrorPtr{}
	isPtrErr := As(wrappedPtrErr, &ptrErrorDestination)

	// then
	assert.True(t, isPtrErr)
	assert.Equal(t, ptrErr.msg, ptrErrorDestination.msg)
}

func TestAsForStructErrorAsInterface(t *testing.T) {
	// given
	structErr := testErrorStruct{msg: "TestErrorStructMsg"}
	wrappedStructErr := Wrap(structErr, "A", "B")

	// when
	var structErrDestination testErrorInterface
	isStructErr := As(wrappedStructErr, &structErrDestination)

	// then
	assert.True(t, isStructErr)
	assert.Equal(t, structErr.msg, structErrDestination.TestErrorMessage())
}

func TestAsForPtrErrorAsInterface(t *testing.T) {
	// given
	ptrErr := &testErrorPtr{msg: "TestErrorPtrMsg"}
	wrappedPtrErr := Wrap(ptrErr, "C", "D")

	// when
	var ptrErrorDestination testErrorInterface
	isPtrErr := As(wrappedPtrErr, &ptrErrorDestination)

	// then
	assert.True(t, isPtrErr)
	assert.Equal(t, ptrErr.msg, ptrErrorDestination.TestErrorMessage())
}

func TestAsWrappingNil(t *testing.T) {
	// given
	wrappedNil := Wrap(nil, "A", "B")

	// when
	var errDestination testErrorInterface
	returnValue := As(wrappedNil, &errDestination)

	// then
	assert.False(t, returnValue)
}

// Test relevant structures
type testErrorStruct struct {
	msg string
}

func (t testErrorStruct) Error() string {
	return ""
}
func (t testErrorStruct) TestErrorMessage() string {
	return t.msg
}

type testErrorPtr struct {
	msg string
}

func (t *testErrorPtr) Error() string {
	return ""
}
func (t *testErrorPtr) TestErrorMessage() string {
	return t.msg
}

// Helper functions
func assertError(t *testing.T, expectedTitle, expectedMessage string, err error) {
	assert.NotNil(t, err)

	if customErr, ok := err.(Error); ok {
		assert.Equal(t, expectedTitle, customErr.Title())
		assert.Equal(t, expectedMessage, customErr.Message())
		assert.NotEqual(t, "", customErr.Stacktrace())
		assert.NotNil(t, customErr.Unwrap())
	} else {
		assert.Fail(t, "Error is not the custom error type")
	}
}
