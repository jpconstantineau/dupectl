package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging to stderr
type Logger struct {
	level  LogLevel
	output io.Writer
}

var defaultLogger = &Logger{
	level:  INFO,
	output: os.Stderr,
}

// SetLevel sets the global log level
func SetLevel(level LogLevel) {
	defaultLogger.level = level
}

// SetOutput sets the output writer (for testing)
func SetOutput(w io.Writer) {
	defaultLogger.output = w
}

// log writes a log message with timestamp and level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	levelStr := map[LogLevel]string{
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
	}[level]

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.output, "[%s] %s: %s\n", timestamp, levelStr, message)
}

// Debug logs debug messages
func Debug(format string, args ...interface{}) {
	defaultLogger.log(DEBUG, format, args...)
}

// Info logs informational messages
func Info(format string, args ...interface{}) {
	defaultLogger.log(INFO, format, args...)
}

// Warn logs warning messages
func Warn(format string, args ...interface{}) {
	defaultLogger.log(WARN, format, args...)
}

// Error logs error messages
func Error(format string, args ...interface{}) {
	defaultLogger.log(ERROR, format, args...)
}
