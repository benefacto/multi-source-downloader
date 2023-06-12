// Package main contains the entrypoint of the application.
package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/benefacto/multi-source-downloader/pkg/downloader"
	"github.com/benefacto/multi-source-downloader/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	l := logger.NewLogger(
		log.New(os.Stdout, "INFO: ", log.LstdFlags),
		log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		log.New(os.Stderr, "WARNING: ", log.LstdFlags),
	)

	chunkSize, err := strconv.Atoi(os.Getenv("CHUNK_SIZE"))
	if err != nil {
		log.Fatal("Failed to parse CHUNK_SIZE environment variable: ", err)
	}

	maxRetries, err := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	if err != nil {
		log.Fatal("Failed to parse MAX_RETRIES environment variable: ", err)
	}

	numOfChunks, err := strconv.Atoi(os.Getenv("NUM_OF_CHUNKS"))
	if err != nil {
		log.Fatal("Failed to parse NUM_OF_CHUNKS environment variable: ", err)
	}

	params := downloader.DownloadParams{
		URL:            os.Getenv("URL"),
		FileExtension:  os.Getenv("FILE_EXTENSION"),
		ChunkSize:      chunkSize,
		MaxRetries:     maxRetries,
		NumberOfChunks: numOfChunks,
	}

	// Creating a context with a timeout of 20 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()  // It's important to cancel when we are finished, to free resources

	_, err = downloader.DownloadFile(ctx, params, l)
	if err != nil {
		log.Fatal(err)
	}
}
