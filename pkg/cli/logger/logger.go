package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Level represents the logging level
type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

// Formatter is the interface for formatting log messages
type Formatter interface {
	Format(level Level, timestamp time.Time, message string, fields ...string) string
}

// DefaultFormatter is the default formatter
type DefaultFormatter struct{}

// Format formats the log message
func (f *DefaultFormatter) Format(level Level, timestamp time.Time, message string, fields ...string) string {
	var levelStr string
	switch level {
	case Debug:
		levelStr = "DEBUG"
	case Info:
		levelStr = "INFO"
	case Warn:
		levelStr = "WARN"
	case Error:
		levelStr = "ERROR"
	default:
		levelStr = "UNKNOWN"
	}

	timestampStr := timestamp.Format("2006-01-02T15:04:05")

	if len(fields) > 0 {
		fieldsStr := strings.Join(fields, " ")
		return fmt.Sprintf("%s [%s] %s: %s", timestampStr, levelStr, message, fieldsStr)
	}

	return fmt.Sprintf("%s [%s] %s", timestampStr, levelStr, message)
}

// Logger is the main logger struct
type Logger struct {
	level     Level
	formatter Formatter
	output    *os.File
}

// New creates a new logger
func New(level Level) *Logger {
	return &Logger{
		level:     level,
		formatter: &DefaultFormatter{},
		output:    os.Stderr,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...string) {
	if l.level <= Debug {
		fmt.Fprintln(l.output, l.formatter.Format(Debug, time.Now(), message, fields...))
	}
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...string) {
	if l.level <= Info {
		fmt.Fprintln(l.output, l.formatter.Format(Info, time.Now(), message, fields...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...string) {
	if l.level <= Warn {
		fmt.Fprintln(l.output, l.formatter.Format(Warn, time.Now(), message, fields...))
	}
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...string) {
	fmt.Fprintln(l.output, l.formatter.Format(Error, time.Now(), message, fields...))
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}
