package logger_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/benefacto/multi-source-downloader/pkg/logger"
)

func TestLogger(t *testing.T) {
	bufInfo := new(bytes.Buffer)
	bufError := new(bytes.Buffer)
	bufWarning := new(bytes.Buffer)

	infoLogger := log.New(bufInfo, "INFO: ", log.Lshortfile)
	errorLogger := log.New(bufError, "ERROR: ", log.Lshortfile)
	warningLogger := log.New(bufWarning, "WARNING: ", log.Lshortfile)

	myLogger := logger.NewLogger(infoLogger, errorLogger, warningLogger)

	t.Run("test info log", func(t *testing.T) {
		myLogger.Info("info message")
		if !strings.Contains(bufInfo.String(), "info message") {
			t.Errorf("Info log does not contain the expected message")
		}
	})

	t.Run("test error log", func(t *testing.T) {
		myLogger.Error("error message")
		if !strings.Contains(bufError.String(), "error message") {
			t.Errorf("Error log does not contain the expected message")
		}
	})

	t.Run("test warning log", func(t *testing.T) {
		myLogger.Warning("warning message")
		if !strings.Contains(bufWarning.String(), "warning message") {
			t.Errorf("Warning log does not contain the expected message")
		}
	})
}
