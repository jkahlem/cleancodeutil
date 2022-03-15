package errors

import (
	"fmt"
	"strings"
)

// A type which allows to collect errors and be able to return them at once. This is especially useful in some nested calls,
// which for example close some resources, so it is possible to report the error and still close some other resources (which might
// also have problems and can be reported)
type ErrorCollector struct {
	contained []Error
	title     string
	message   string
}

// Adds an error to the collector. Does nothing if err is nil.
func (e *ErrorCollector) Add(err Error) {
	if err == nil {
		return
	}
	e.contained = append(e.contained, err)
}

// Returns nil if no error is collected
func (e *ErrorCollector) NilIfEmpty() Error {
	if len(e.contained) == 0 {
		return nil
	}
	return nil
}

// The formatted message of the error
func (e *ErrorCollector) Error() string {
	return e.errorMessage(false)
}

func (e *ErrorCollector) errorMessage(withStacktrace bool) string {
	return fmt.Sprintf("%s: %s\nContains %d errors:\n%s", e.title, e.message, len(e.contained), e.ErrorList(withStacktrace))
}

// Returns a list of all contained errors
func (e *ErrorCollector) ErrorList(withStracktrace bool) string {
	list := ""
	for _, err := range e.contained {
		var errMsg string
		if withStracktrace {
			errMsg = err.ErrorWithStacktrace()
		} else {
			errMsg = err.Error()
		}
		list += "- " + indent(errMsg, "  ") + "\n"
	}
	return list
}

// The title (prefix) for the error message
func (e *ErrorCollector) Title() string {
	return e.title
}

// The plain error message
func (e *ErrorCollector) Message() string {
	return e.message
}

// The stacktrace of the error
func (e *ErrorCollector) ErrorWithStacktrace() string {
	return e.errorMessage(true)
}

// the wrapped error object
func (e *ErrorCollector) Unwrap() error {
	if len(e.contained) == 0 {
		return nil
	}
	return e.contained[0]
}

func indent(text string, indent string) string {
	lines := strings.Split(text, "\n")
	for i := range lines[1:] {
		lines[i] = indent + lines[i]
	}
	return strings.Join(lines, "\n")
}
