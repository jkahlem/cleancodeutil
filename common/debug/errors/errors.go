// The errors package extends the golang errors and the go-errors package to:
// - have a consistent format (title: description)
// - contain the error message of wrapped errors in an error's own error message
// - have a stacktrace (provided by the go-errors package)
//
// This error type should be used as possible in the whole application to have
// consistent error handling. Errors from go packages/external packages should always
// be wrapped in the errors of this packages using the Wrap function.
package errors

import (
	golangErrors "errors"
	"fmt"

	goErrors "github.com/go-errors/errors"
)

// Used for stacktraces for errors created with New function (as goErrors.New does not support skipping stack frames)
type dummyErrorForCreationByNew struct{}

func (e *dummyErrorForCreationByNew) Error() string {
	return ""
}

type Error interface {
	// The formatted message of the error
	Error() string
	// The title (prefix) for the error message
	Title() string
	// The plain error message
	Message() string
	// The stacktrace of the error
	ErrorWithStacktrace() string
	// the wrapped error object
	Unwrap() error
	// For checks
	Is(error) bool
}

type customError struct {
	wrappedErr error
	title      string
	message    string
}

func (e *customError) ErrorWithStacktrace() string {
	return fmt.Sprintf("%s\n%s", e.Error(), e.Stacktrace())
}

// Prints the error message containing a title, the message and the wrapped errors
func (e *customError) Error() string {
	// If it contains a dummy error, no need to show the error message of the dummy error.
	// (as it was created by new)
	return fmt.Sprintf("%s: %s%s", e.title, e.message, e.errorMsgOfWrappedError())
}

// creates a list of error messages of all wrapped errors
func (e *customError) errorMsgOfWrappedError() string {
	if e.wrappedErr == nil {
		return ""
	} else if wrappedGoErr, ok := e.wrappedErr.(*goErrors.Error); !ok {
		// Case: The wrapped error is not a goError. It is either this new Error type or an unknown error type.
		return "\n  contains: " + e.wrappedErr.Error()
	} else if nextErr := golangErrors.Unwrap(wrappedGoErr); nextErr != nil {
		// Case: The wrapped error is a goError containing another error. As this new error type is usually created using
		//       the New and Wrap functions, the wrappedErr chain should always end with a goError used just for wrapping
		//       another error (either a real error when using Wrap or the dummy error created by New)
		if _, isDummyError := nextErr.(*dummyErrorForCreationByNew); isDummyError {
			// the underlying error is a dummy error: No message as it contains no information
			return ""
		}
		// the underlying error is a real error, so print it's error message
		return "\n  contains: " + nextErr.Error()
	} else {
		// otherwise just print the errors error message.
		return wrappedGoErr.Error()
	}
}

// Returns the stacktrace of the error (pointing to where the error was created or an external package error was wrapped).
func (e *customError) Stacktrace() string {
	// Use the stacktrace of the lowest go-error
	goError := e.findGoError()
	if goError != nil {
		return string(goError.Stack())
	}
	return ""
}

// Returns the title of the error.
func (e *customError) Title() string {
	return e.title
}

// Returns only the error message of this error.
func (e *customError) Message() string {
	return e.message
}

// Finds a go-errors error from errors wrapped by this error
func (e *customError) findGoError() *goErrors.Error {
	for unwrapped := golangErrors.Unwrap(e); unwrapped != nil; unwrapped = golangErrors.Unwrap(unwrapped) {
		if goError, ok := unwrapped.(*goErrors.Error); ok {
			return goError
		}
	}
	return nil
}

// Returns the wrapped error or nil.
func (e *customError) Unwrap() error {
	return e.wrappedErr
}

func (e *customError) Is(err error) bool {
	if x, ok := err.(interface {
		Title() string
		Message() string
	}); ok {
		return x.Title() == e.title && x.Message() == e.message
	}
	return false
}

type ErrorIdentifier struct {
	title   string
	message string
}

func (e ErrorIdentifier) Title() string {
	return e.title
}

func (e ErrorIdentifier) Message() string {
	return e.message
}

func (e ErrorIdentifier) Error() string {
	return e.title + ": " + e.message
}

func ErrorId(title, message string, args ...interface{}) ErrorIdentifier {
	return ErrorIdentifier{
		title:   title,
		message: fmt.Sprintf(message, args...),
	}
}

// Wraps the given error in the new error type. Does nothing if err is nil.
func Wrap(err error, title, message string, args ...interface{}) Error {
	if err == nil {
		return nil
	}
	return &customError{
		wrappedErr: wrapInGoErrorIfUnknownError(err),
		title:      title,
		message:    fmt.Sprintf(message, args...),
	}
}

func WrapWithId(err error, id ErrorIdentifier) Error {
	return Wrap(err, id.title, id.message)
}

// Creates a new error. The args parameters are formatting arguments for the passed message.
func New(title, message string, args ...interface{}) Error {
	return &customError{
		wrappedErr: wrapInGoErrorIfUnknownError(&dummyErrorForCreationByNew{}),
		title:      title,
		message:    fmt.Sprintf(message, args...),
	}
}

func NewById(id ErrorIdentifier) Error {
	return New(id.title, id.message)
}

// For getting an error of a type. Mirrors golang's errors.As but without panicking.
func As(err error, target interface{}) bool {
	defer recover()
	return golangErrors.As(err, target)
}

// For checking the error type of an error. Mirrors golang's errors.Is.
func Is(err, target error) bool {
	return golangErrors.Is(err, target)
}

// For stacktracing, wrap errors in go-errors Errors. Do nothing if it is an go-errors Error or this custom error type.
func wrapInGoErrorIfUnknownError(err error) error {
	if _, ok := err.(*customError); ok {
		return err
	} else if _, ok := err.(*goErrors.Error); ok {
		return err
	}
	return goErrors.Wrap(err, 2)
}
