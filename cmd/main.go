package main

import (
	"log"
	"os"

	"github.com/benefacto/multi-source-downloader/downloader"
)

type logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	warningLogger *log.Logger
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

func main() {
	l := &logger{
		infoLogger:    log.New(os.Stdout, "INFO: ", log.LstdFlags),
		errorLogger:   log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		warningLogger: log.New(os.Stderr, "WARNING: ", log.LstdFlags),
	}

	params := downloader.DownloadParams{
		URL:           "https://zenodo.org/record/4435114/files/supplement.csv?download=1",
		ChunkSize:     1024,
		MaxRetries:    3,
		NumberOfChunks: 3,
	}
	err := downloader.DownloadFile(params, l)
	if err != nil {
		log.Fatal(err)
	}
}
