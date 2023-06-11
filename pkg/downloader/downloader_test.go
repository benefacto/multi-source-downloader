package downloader_test

import (
	"github.com/benefacto/multi-source-downloader/pkg/downloader"
	"github.com/benefacto/multi-source-downloader/pkg/logger"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	// create a logger
	l := logger.NewLogger(
		log.New(os.Stdout, "INFO: ", log.LstdFlags),
		log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		log.New(os.Stderr, "WARNING: ", log.LstdFlags),
	)
	// setup parameters
	params := downloader.DownloadParams{
		// Using a small test file for a quick download
		URL:            "https://zenodo.org/record/4435114/files/users_inferred.csv?download=1",
		FileExtension:  "csv",
		ChunkSize:      1024,
		MaxRetries:     3,
		NumberOfChunks: 3,
	}
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
	// TODO: Add more assertions based on what you know about the specific output file,
	// such as its expected size, contents, etc.
	defer cleanupOutputDirectory(t, "./output")
	os.Remove(fileName)
}

func cleanupOutputDirectory(t *testing.T, dir string) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.RemoveAll(path)
	})

	if err != nil {
		t.Fatalf("Failed to clean up test directory %s: %v", dir, err)
	}
}
