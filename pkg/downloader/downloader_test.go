// Package downloader_test contains tests for the downloader package.
package downloader_test

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/benefacto/multi-source-downloader/pkg/downloader"
	"github.com/benefacto/multi-source-downloader/pkg/logger"
)

// TestDownloadFile tests the DownloadFile function with a valid URL and parameters.
func TestDownloadFile(t *testing.T) {
	l := getTestLogger()
	params := getDownloadParams("https://zenodo.org/record/4435114/files/users_inferred.csv?download=1")
	
	// execute function
	fileName, err := downloader.DownloadFile(params, l)

	// assert no error
	if err != nil {
		t.Fatal(err)
	}
	if fileName == "" {
		t.Fatal("File name not returned")
	}
	// verify that the output file exists
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Fatal("Output file was not created.")
	}
	
	// Open the file for reading
	file, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Read the first row
	firstRow, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read the first row: %v", err)
	}

	// Check if the first row contains the header "id"
	if len(firstRow) != 1 || firstRow[0] != "id" {
		t.Fatalf("Expected header to be 'id', got %v", firstRow)
	}

	// Check all rows to ensure they contain a single column of data that can be parsed as an integer
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read row: %v", err)
		}

		if len(row) != 1 {
			t.Fatalf("Expected single column data, got %v", row)
		}

		_, err = strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			t.Fatalf("Failed to parse id '%s' as integer", row[0])
		}
	}

	os.RemoveAll("./output")
}

// TestDownloadFile_ServerError tests the DownloadFile function with a server that responds with an internal server error.
func TestDownloadFile_ServerError(t *testing.T) {
	// Create a test server that always responds with an internal server error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	l := getTestLogger()
	params := getDownloadParams(ts.URL)
	
	// execute function
	_, err := downloader.DownloadFile(params, l)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
	os.RemoveAll("./output")
}

// TestDownloadFile_MissingEtagHeader tests the DownloadFile function with a server that does not include the Etag header in its response.
func TestDownloadFile_MissingEtagHeader(t *testing.T) {
	// Create a test server that responds without the Etag header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	l := getTestLogger()
	params := getDownloadParams(ts.URL)

	// execute function
	_, err := downloader.DownloadFile(params, l)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
	os.RemoveAll("./output")
}

// TestDownloadFile_MissingContentLengthHeader tests the DownloadFile function with a server that does not include the Content-Length header in its response.
func TestDownloadFile_MissingContentLengthHeader(t *testing.T) {
	// Create a test server that responds without the Content-Length header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Etag", "dummyetag")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	l := getTestLogger()
	params := getDownloadParams(ts.URL)

	// execute function
	_, err := downloader.DownloadFile(params, l)
	if err == nil {
		t.Fatal("Expected error, but got none")
	}
	os.RemoveAll("./output")
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
