// Package logger_test contains tests for the logger package.
package logger_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/benefacto/multi-source-downloader/pkg/logger"
)

// testMessage is the message used in the logger tests.
const testMessage = "test message"

// TestLogger tests the logger's Info, Error, and Warning methods.
func TestLogger(t *testing.T) {
	bufInfo := new(bytes.Buffer)
	bufError := new(bytes.Buffer)
	bufWarning := new(bytes.Buffer)

	infoLogger := log.New(bufInfo, "INFO: ", log.Lshortfile)
	errorLogger := log.New(bufError, "ERROR: ", log.Lshortfile)
	warningLogger := log.New(bufWarning, "WARNING: ", log.Lshortfile)

	myLogger := logger.NewLogger(infoLogger, errorLogger, warningLogger)

	t.Run("Info logs the expected message", func(t *testing.T) {
		myLogger.Info(testMessage)
		assertLogContains(t, bufInfo, testMessage)
	})

	t.Run("Error logs the expected message", func(t *testing.T) {
		myLogger.Error(testMessage)
		assertLogContains(t, bufError, testMessage)
	})

	t.Run("Warning logs the expected message", func(t *testing.T) {
		myLogger.Warning(testMessage)
		assertLogContains(t, bufWarning, testMessage)
	})
}

// assertLogContains checks that the log buffer contains the expected message and fails the test if it does not.
func assertLogContains(t *testing.T, buf *bytes.Buffer, expected string) {
	t.Helper()
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("Log does not contain the expected message: %s", expected)
	}
}
