package logger

import (
	"log"
)

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Warning(args ...interface{})
}

type logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	warningLogger *log.Logger
}

func NewLogger(infoLogger, errorLogger, warningLogger *log.Logger) Logger {
	return &logger{
		infoLogger:    infoLogger,
		errorLogger:   errorLogger,
		warningLogger: warningLogger,
	}
}

func (l *logger) Info(args ...interface{}) {
	l.infoLogger.Println(args...)
}

func (l *logger) Error(args ...interface{}) {
	l.errorLogger.Println(args...)
}

func (l *logger) Warning(args ...interface{}) {
	l.warningLogger.Println(args...)
}
