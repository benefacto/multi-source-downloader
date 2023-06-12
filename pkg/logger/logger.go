// Package logger provides a logging interface and an implementation for logging
// informational messages, error messages, and warnings.
package logger

import (
	"log"
)

// Logger is an interface that provides methods for logging informational messages,
// error messages, and warnings.
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Warning(args ...interface{})
}

// logger is a struct that implements the Logger interface.
type logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	warningLogger *log.Logger
}

// NewLogger creates a new logger with the provided infoLogger, errorLogger, and warningLogger.
func NewLogger(infoLogger, errorLogger, warningLogger *log.Logger) Logger {
	return &logger{
		infoLogger:    infoLogger,
		errorLogger:   errorLogger,
		warningLogger: warningLogger,
	}
}

// Info logs an informational message.
func (l *logger) Info(args ...interface{}) {
	l.infoLogger.Println(args...)
}

// Error logs an error message.
func (l *logger) Error(args ...interface{}) {
	l.errorLogger.Println(args...)
}

// Warning logs a warning message.
func (l *logger) Warning(args ...interface{}) {
	l.warningLogger.Println(args...)
}
