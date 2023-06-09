package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	// TODO: Move this to config or have this be a default for a command line argument
	fileURL = "https://zenodo.org/record/4435114/files/supplement.csv?download=1"
)

func main() {
	infoLogger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLogger := log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	// warningLogger := log.New(os.Stderr, "WARNING: ", log.LstdFlags)

	resp, err := http.Head(fileURL)
	if err != nil {
		errorLogger.Println("Error making HTTP HEAD request to", fileURL, err)
		return
	} else {
		infoLogger.Println("HTTP HEAD request successfully made to", fileURL)
	}

	etag := resp.Header.Get("Etag")
	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		errorLogger.Println("Error getting length of Etag header", etag, "for", fileURL)
	} else {
		infoLogger.Println("File is", fileSize, "bytes with Etag:", etag)
	}

	// TODO: Move to config or otherwise determine dynamically somehow
	const numberOfChunks = 4
	const maxRetries = 5

	infoLogger.Println("Downloading file in", numberOfChunks, "chunks...")

	chunkSize := fileSize / numberOfChunks
	remainingBytes := fileSize % numberOfChunks

	var wg sync.WaitGroup

	for currentChunkIndex := 0; currentChunkIndex < numberOfChunks; currentChunkIndex++ {
		wg.Add(1)
		go func(currentChunkIndex int) {
			defer wg.Done()

			for i := 0; i < maxRetries; i++ {
				start := currentChunkIndex * chunkSize
				end := start + chunkSize - 1
				if currentChunkIndex == numberOfChunks - 1 {
					end += remainingBytes
				}

				client := &http.Client{}
				req, err := http.NewRequest("GET", fileURL, nil)
				if err != nil {
					errorLogger.Println("Chunk", currentChunkIndex, ": Error making HTTP GET request to", fileURL, err)
				} else {
					infoLogger.Println("Chunk", currentChunkIndex, ": Downloading file...")
				}

				rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
				req.Header.Add("Range", rangeHeader)
				resp, err := client.Do(req)
				if err != nil {
					errorLogger.Println("Chunk", currentChunkIndex, ": Error making HTTP Range request to", fileURL, err)
				} else {
					infoLogger.Println("Chunk", currentChunkIndex, ": Downloading file range for", rangeHeader, "...")
				}

				tmpFileName := fmt.Sprintf("tmpfile_%d", currentChunkIndex)
				tmpFile, err := os.OpenFile(tmpFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					errorLogger.Println("Error creating temporary file", tmpFileName, err)
				} else {
					infoLogger.Println("Successfully created temporary file", tmpFileName)
				}
				defer tmpFile.Close()

				_, err = io.Copy(tmpFile, resp.Body)
				if err != nil {
					errorLogger.Println("Error copying response body to temporary file", tmpFileName, err)
				}
				if err != nil {
					if i < maxRetries-1 {
						errorLogger.Println("Error downloading chunk", currentChunkIndex, ". Retrying...")
						continue
					} else {
						errorLogger.Println("Error downloading chunk", currentChunkIndex, ". No more retries.")
						return
					}
				} else {
					infoLogger.Println("Successfully wrote chunk", currentChunkIndex, "to temporary file", tmpFileName)
					break
				}
			}

		}(currentChunkIndex)
	}

	wg.Wait()

	t := time.Now()
	timestamp := t.Format("20060102_150405")
	fileName := fmt.Sprintf("output_%s.csv", timestamp)

	finalFile, err := os.Create(fileName)
	if err != nil {
		errorLogger.Println("Error creating final file", fileName, err)
		return
	}
	defer finalFile.Close()

	for currentChunkIndex := 0; currentChunkIndex < numberOfChunks; currentChunkIndex++ {
		tmpFileName := fmt.Sprintf("tmpfile_%d", currentChunkIndex)
		tmpFile, err := os.Open(tmpFileName)
		if err != nil {
			errorLogger.Println("Error opening temporary file", tmpFileName, err)
			return
		}

		_, err = io.Copy(finalFile, tmpFile)
		tmpFile.Close()
		if err != nil {
			errorLogger.Println("Error copying temporary file", tmpFileName, "to final file", fileName, err)
			return
		}

		os.Remove(tmpFileName)
	}

	infoLogger.Println("Successfully merged all chunks into final file", fileName)

	// TODO: Compare hash of final file with etag from server to verify download integrity
}
