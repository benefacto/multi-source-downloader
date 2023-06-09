package main

import (
	"crypto/sha256"
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
	fileURL = "https://zenodo.org/record/4435114/files/supplement.csv?download=1"
)

func main() {
	infoLogger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLogger := log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	warningLogger := log.New(os.Stderr, "WARNING: ", log.LstdFlags)

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

	numberOfChunks := 4

	infoLogger.Println("Downloading file in", numberOfChunks, "chunks")

	chunkSize := fileSize / numberOfChunks
	remainingBytes := fileSize % numberOfChunks

	chunks := make([][]byte, numberOfChunks)
	var wg sync.WaitGroup

	for currentChunkIndex := 0; currentChunkIndex < numberOfChunks; currentChunkIndex++ {
		wg.Add(1)
		go func(currentChunkIndex int) {
			defer wg.Done()

			start := currentChunkIndex * chunkSize
			end := start + chunkSize - 1
			if currentChunkIndex == numberOfChunks - 1 {
				end += remainingBytes
			}

			client := &http.Client{}
			req, err := http.NewRequest("GET", fileURL, nil)
			if err != nil {
				errorLogger.Println("Chunk", currentChunkIndex, ": Error making HTTP GET request to", fileURL, err)
				return
			} else {
				infoLogger.Println("Chunk", currentChunkIndex, ": Downloading file in", numberOfChunks, "chunks")
			}

			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
			req.Header.Add("Range", rangeHeader)
			resp, err := client.Do(req)
			if err != nil {
				errorLogger.Println("Chunk", currentChunkIndex, ": Error making HTTP Range request to", fileURL, err)
				return
			} else {
				infoLogger.Println("Chunk", currentChunkIndex, ": Downloading file range for", rangeHeader)
			}

			body := make([]byte, end-start+1)
			_, err = io.ReadFull(resp.Body, body)
			if err != nil {
				errorLogger.Println("Chunk", currentChunkIndex, ": Error reading response body", err)
				return
			} else {
				infoLogger.Println("Chunk", currentChunkIndex, ": Body successfully read")
			}

			chunks[currentChunkIndex] = body
		}(currentChunkIndex)
	}

	wg.Wait()

	t := time.Now()
	timestamp := t.Format("20060102_150405")
	fileName := fmt.Sprintf("output_%s.csv", timestamp)
	file, err := os.Create(fileName)
	if err != nil {
		errorLogger.Println("Error creating file", fileName, err)
		return
	} else {
		infoLogger.Println("File", fileName, "created successfully.")
	}
	defer file.Close()

	hash := sha256.New()
	for currentChunkIndex := 0; currentChunkIndex < numberOfChunks; currentChunkIndex++ {
		_, err = file.Write(chunks[currentChunkIndex])
		if err != nil {
			errorLogger.Println("Error writing chunk", currentChunkIndex, "to file", fileName , err)
			return
		} else {
			infoLogger.Println("Successfully wrote chunk", currentChunkIndex, "to file", fileName)
		}

		_, err = hash.Write(chunks[currentChunkIndex])
		if err != nil {
			errorLogger.Println("Error writing chunk", currentChunkIndex, "hash", err)
		} else {
			infoLogger.Println("Successfully wrote chunk", currentChunkIndex, "hash")
		}
	}

	downloadedFileHash := fmt.Sprintf(`"%x"`, hash.Sum(nil))
	if downloadedFileHash == etag {
		infoLogger.Println("Download completed successfully. File hash matches the ETag.")
	} else {
		warningLogger.Println("File hash:", downloadedFileHash ,"does not match the ETag:", etag)
	}
}
