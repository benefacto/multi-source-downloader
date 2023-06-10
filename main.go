package main

import (
	"crypto/md5"
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
	maxRetries = 3
	numberOfChunks = 3
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
		infoLogger.Println("Successfully made HTTP HEAD request to", fileURL)
	}

	etag := resp.Header.Get("Etag")
	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		errorLogger.Println("Error getting length of ETag header", etag, "for", fileURL)
	} else {
		infoLogger.Println("File is", fileSize, "bytes with ETag:", etag)
	}

	chunkSize := fileSize / numberOfChunks
	remainingBytes := fileSize % numberOfChunks

	var wg sync.WaitGroup
	md5Hash := md5.New()
	tempFiles := make([]string, numberOfChunks) // to store names of temporary files
	client := &http.Client{}

	for currentChunkIndex := 0; currentChunkIndex < numberOfChunks; currentChunkIndex++ {
		wg.Add(1)
		go func(currentChunkIndex int, client *http.Client) {
			defer wg.Done()

			var err error
			for retries := 0; retries < maxRetries; retries++ {
				start := currentChunkIndex * chunkSize
				end := start + chunkSize - 1
				if currentChunkIndex == numberOfChunks - 1 {
					end += remainingBytes
				}

				req, err := http.NewRequest("GET", fileURL, nil)
				if err != nil {
					errorLogger.Println("Chunk", currentChunkIndex, "had an error making HTTP GET request to", fileURL, err)
					return
				} else {
					infoLogger.Println("Chunk", currentChunkIndex, "successfully made HTTP GET request to", fileURL)
				}

				rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
				req.Header.Add("Range", rangeHeader)

				resp, err := client.Do(req)
				if err != nil {
					errorLogger.Println("Chunk", currentChunkIndex, "had an error making HTTP Range request to", fileURL, err)
					if retries < maxRetries - 1 {
						infoLogger.Println("Retrying chunk", currentChunkIndex, "...")
						time.Sleep(time.Second * time.Duration(retries+1)) // exponential back-off
						continue
					}
					return
				}

				tmpFileName := fmt.Sprintf("tmpfile_%d", currentChunkIndex)
				tempFiles[currentChunkIndex] = tmpFileName // save name of temp file
				tmpFile, err := os.Create(tmpFileName)
				if err != nil {
					errorLogger.Println("Chunk", currentChunkIndex, "had an error creating temporary file", tmpFileName, err)
					return
				} else {
					infoLogger.Println("Chunk", currentChunkIndex, "successfully created temporary file", tmpFileName)
				}

				_, err = io.Copy(tmpFile, resp.Body)
				if err != nil {
					errorLogger.Println("Chunk", currentChunkIndex, "had an error writing to temporary file", tmpFileName, err)
					return
				} else {
					infoLogger.Println("Chunk", currentChunkIndex, "successfully wrote to temporary file", tmpFileName)
				}

				resp.Body.Close()
				tmpFile.Close()
				break
			}

			if err != nil {
				errorLogger.Println("Chunk", currentChunkIndex, "had an error after retries", err)
			}
		}(currentChunkIndex, client)
	}

	wg.Wait()

	for _, tmpFileName := range tempFiles {
		tmpFile, err := os.Open(tmpFileName)
		if err != nil {
			errorLogger.Println("Error opening temporary file", tmpFileName, err)
			return
		} else {
			infoLogger.Println("Successfully opened temporary file", tmpFileName)
		}

		_, err = io.Copy(md5Hash, tmpFile)
		if err != nil {
			errorLogger.Println("Error hashing content of temporary file", tmpFileName, err)
			return
		} else {
			infoLogger.Println("Successfully hashed content of temporary file", tmpFileName)
		}

		tmpFile.Close()
		os.Remove(tmpFileName)
	}

	checksum := fmt.Sprintf(`"md5:%x"`, md5Hash.Sum(nil))
	if checksum != etag {
		warningLogger.Println("MD5 checksum", checksum, "does not match ETag", etag)
	} else {
		infoLogger.Println("MD5 checksum", checksum, "matches ETag", etag)
	}
}