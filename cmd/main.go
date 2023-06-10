package main

import (
	"log"
	"os"

	"github.com/benefacto/multi-source-downloader/pkg/downloader"
	"github.com/benefacto/multi-source-downloader/pkg/logger"
)

func main() {
	l := logger.NewLogger(
		log.New(os.Stdout, "INFO: ", log.LstdFlags),
		log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		log.New(os.Stderr, "WARNING: ", log.LstdFlags),
	)

	params := downloader.DownloadParams{
		URL:           "https://zenodo.org/record/4435114/files/supplement.csv?download=1",
		ChunkSize:     1024,
		MaxRetries:    3,
		NumberOfChunks: 3,
	}
	_, err := downloader.DownloadFile(params, l)
	if err != nil {
		log.Fatal(err)
	}
}
