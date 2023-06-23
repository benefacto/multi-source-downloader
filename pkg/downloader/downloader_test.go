// Package downloader_test contains tests for the downloader package.
package downloader_test

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/benefacto/multi-source-downloader/pkg/downloader"
	"github.com/benefacto/multi-source-downloader/pkg/logger"
)

// TestDownloadFile_ServerError tests the DownloadFile function with a server that responds with an internal server error.
func TestDownloadFile_ServerError(t *testing.T) {
	defer os.RemoveAll("./output")

	// Create a test server that always responds with an internal server error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	l := getTestLogger()
	params := getDownloadParams(ts.URL)
	ctx, cancel := getContext()
	defer cancel() // Cancel when we are finished, to free resources

	// execute function
	_, err := downloader.DownloadFile(ctx, params, l)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
}

// TestDownloadFile_MissingEtagHeader tests the DownloadFile function with a server that does not include the Etag header in its response.
func TestDownloadFile_MissingEtagHeader(t *testing.T) {
	defer os.RemoveAll("./output")

	responseBody := "id\n" + strings.Repeat("123\n456\n789\n", 120)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer ts.Close()

	l := getTestLogger()
	params := getDownloadParams(ts.URL)
	ctx, cancel := getContext()
	defer cancel() // Cancel when we are finished, to free resources

	// execute function
	_, err := downloader.DownloadFile(ctx, params, l)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
}

// TestDownloadFile_MissingContentLengthHeader tests the DownloadFile function with a server that does not include the Content-Length header in its response.
func TestDownloadFile_MissingContentLengthHeader(t *testing.T) {
	defer os.RemoveAll("./output")

	responseBody := "id\n" + strings.Repeat("123\n456\n789\n", 120)
	hash := md5.Sum([]byte(responseBody))
	etag := fmt.Sprintf("\"md5:%x\"", hash[:])
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Etag", etag)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer ts.Close()

	l := getTestLogger()
	params := getDownloadParams(ts.URL)
	ctx, cancel := getContext()
	defer cancel() // Cancel when we are finished, to free resources

	// execute function
	_, err := downloader.DownloadFile(ctx, params, l)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
}

// getTestLogger creates and returns a test logger with predefined log outputs.
func getTestLogger() logger.Logger {
	return logger.NewLogger(
		log.New(os.Stdout, "INFO: ", log.LstdFlags),
		log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		log.New(os.Stderr, "WARNING: ", log.LstdFlags),
	)
}

// getDownloadParams creates and returns DownloadParams with the provided URL.
func getDownloadParams(url string) downloader.DownloadParams {
	return downloader.DownloadParams{
		URL:            url,
		FileExtension:  "csv",
		ChunkSize:      1024,
		MaxRetries:     3,
		NumberOfChunks: 3,
	}
}

// getContext creates a context with a timeout of 20 minutes
func getContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	return ctx, cancel
}
