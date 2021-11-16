package log

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/errors"
)

const LoggerErrorTitle = "Logger Error"

// The layers which are used from different modules/for different purposes and can be toggled on/off
// in the configuration.
type Layer string

const (
	Information    Layer = "information"
	Critical       Layer = "critical"
	Communicator   Layer = "communicator"
	LanguageServer Layer = "languageServer"
	Messager       Layer = "messager"
)

// Implements different methods to log messages using upd, stdout or a logfile.
type logger struct {
	conn               *net.UDPConn
	logFile            io.WriteCloser
	addTimestamp       bool
	logToStdoutEnabled bool
	silentErrorLogging bool
	mutex              sync.Mutex
}

var _logger *logger
var problems []string

// Setups the logger for a specific port.
func (l *logger) SetupFileLogging() errors.Error {
	if logFile, err := os.Create(filepath.Join(configuration.GoProjectDir(), "logfile.log")); err != nil {
		return errors.Wrap(err, LoggerErrorTitle, "Could not create logfile")
	} else {
		l.logFile = logFile
		return nil
	}
}

// Setups the remote logger connection.
func (l *logger) SetupRemoteLogging(port int) errors.Error {
	if !configuration.LoggerActivateRemoteLogging() {
		return nil
	}
	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return errors.Wrap(err, LoggerErrorTitle, "Could not connect to remote logger")
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return errors.Wrap(err, LoggerErrorTitle, "Could not connect to remote logger")
	}
	l.conn = conn
	return nil
}

// Formats the message and prints it to the configured log outputs
func (l *logger) Print(layer Layer, format string, args ...interface{}) errors.Error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if !l.isLoggingAllowed(layer) {
		return nil
	}

	logText := l.formatLogMessage(format, args...)
	if layer == Critical && !configuration.LoggerErrorsInConsoleOutput() {
		lines := strings.Split(logText, "\n")
		l.logToStdout(lines[0] + "\n")
	} else {
		l.logToStdout(logText)
	}

	errRemote := l.logToRemote(logText)
	errLogfile := l.logToFile(logText)
	if errRemote != nil && errLogfile != nil {
		return errors.Wrap(errLogfile, LoggerErrorTitle, "Could not log logtext")
	} else if errRemote != nil {
		return errRemote
	} else {
		return errLogfile
	}
}

// Returns true if logging to the layer is allowed.
func (l *logger) isLoggingAllowed(layer Layer) bool {
	if configuration.LoggerLayers() == nil {
		// Default: log criticals/informations but ignore specific stuff like language server logs/communicator logs etc.
		return layer == Critical || layer == Information
	}
	for _, allowed := range configuration.LoggerLayers() {
		if Layer(allowed) == layer {
			return true
		}
	}
	return false
}

// Formats the log message.
func (l *logger) formatLogMessage(format string, args ...interface{}) string {
	if l.addTimestamp {
		format = fmt.Sprintf("%s %s", l.currentTimestamp(), format)
	}
	return fmt.Sprintf(format, args...)
}

// Logs to the remote debug logger connection.
func (l *logger) logToRemote(content string) errors.Error {
	if l.conn != nil {
		if _, err := fmt.Fprintf(l.conn, content); err != nil {
			l.CloseConnection()
			return errors.Wrap(err, LoggerErrorTitle, "Could not send log to remote logger")
		}
	}
	return nil
}

// Writes the logtext to the logfile.
func (l *logger) logToFile(content string) errors.Error {
	if l.logFile != nil {
		if _, err := fmt.Fprintf(l.logFile, content); err != nil {
			l.CloseFile()
			return errors.Wrap(err, LoggerErrorTitle, "Could not write log to logfile")
		}
	}
	return nil
}

// Writes the logtext to the standard output stream.
func (l *logger) logToStdout(content string) {
	if l.logToStdoutEnabled {
		fmt.Print(content)
	}
}

// If true, the logger will not post error messages into the stdout log if logged by Error.
func (l *logger) SetSilentErrorLogging(state bool) {
	l.silentErrorLogging = state
}

// Enables/Disables logging to the standard output stream.
func (l *logger) SetLoggingToStdout(state bool) {
	l.logToStdoutEnabled = state
}

// Returns the current time as a timestamp in hh:mm:ss.000 format.
func (l *logger) currentTimestamp() string {
	now := time.Now()
	return now.Format("[15:04:05.000]")
}

// Closes the connection to the remote debug logger.
func (l *logger) CloseConnection() {
	if l.conn != nil {
		l.conn.Close()
		l.conn = nil
	}
}

// Closes the logfile.
func (l *logger) CloseFile() {
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}
}

// Logs an error and exits the program with exit code 1.
// Used for fatal errors which will heavily impair the output and cannot be handled on program side, so it is better to stop the program.
func FatalError(err errors.Error) {
	Error(err)
	os.Exit(1)
}

// Logs an error.
func Error(err errors.Error) errors.Error {
	return Print(Critical, "%s\n%s", err.Error(), err.Stacktrace())
}

// Logs a simple message informing the user.
func Info(format string, args ...interface{}) errors.Error {
	return Print(Information, format, args...)
}

// Prints a log message on the given layer.
func Print(layer Layer, format string, args ...interface{}) errors.Error {
	if _logger != nil {
		return _logger.Print(layer, format, args...)
	}
	return nil
}

// Enables/Disables logging to the standard output stream.
func SetLoggingToStdout(state bool) {
	createLoggerIfNotExist()
	_logger.SetLoggingToStdout(state)
}

// Sets the port the logger should send log messages to. (Using UDP)
func SetupRemoteLogging(port int) errors.Error {
	createLoggerIfNotExist()
	return _logger.SetupRemoteLogging(port)
}

// If true, the logger will not post error messages into the stdout log if logged by LogError
func SetSilentErrorLogging(state bool) {
	createLoggerIfNotExist()
	_logger.SetSilentErrorLogging(state)
}

// Sets up file logging.
func SetupFileLogging() errors.Error {
	createLoggerIfNotExist()
	return _logger.SetupFileLogging()
}

// Creates a new logger if not exist.
func createLoggerIfNotExist() {
	if _logger == nil {
		_logger = &logger{
			addTimestamp: true,
		}
	}
}

// Closes the logger and the used resources for logging.
func Close() {
	_logger.CloseConnection()
	_logger.CloseFile()
}

// Reports a problem which may have negative influence on some parts of the data generation task
// but are not that critical to stop the whole prorgam from working. Will handled as a fatal error if the
// program is running using strict mode.
func ReportProblemWithError(err errors.Error, problemMessage string, args ...interface{}) {
	if configuration.StrictMode() {
		FatalError(err)
	}
	ReportProblem(problemMessage, args...)
}

// Reports a problem which may have negative influence on some parts of the data generation task
// but are not that critical to stop the whole prorgam from working.
func ReportProblem(format string, args ...interface{}) {
	if problems == nil {
		problems = make([]string, 0, 1)
	}
	problems = append(problems, fmt.Sprintf(format, args...))
}

// Returns a list of the problems reported
func GetProblems() []string {
	return problems
}
