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
	numberOfChunks = 4
)

func main() {
	infoLogger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLogger := log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	warningLogger := log.New(os.Stderr, "WARNING: ", log.LstdFlags)

	resp, err := http.Head(fileURL)
	if err != nil {
		errorLogger.Println("Error making HTTP HEAD request to", fileURL, err)
		return
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

	for currentChunkIndex := 0; currentChunkIndex < numberOfChunks; currentChunkIndex++ {
		wg.Add(1)
		go func(currentChunkIndex int) {
			defer wg.Done()

			var err error
			for retries := 0; retries < maxRetries; retries++ {
				start := currentChunkIndex * chunkSize
				end := start + chunkSize - 1
				if currentChunkIndex == numberOfChunks - 1 {
					end += remainingBytes
				}

				client := &http.Client{}
				req, err := http.NewRequest("GET", fileURL, nil)
				if err != nil {
					errorLogger.Println("Error making HTTP GET request to", fileURL, err)
					return
				} else {
					infoLogger.Println("Successfuly HTTP GET request made to", fileURL)
				}

				rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
				req.Header.Add("Range", rangeHeader)

				resp, err := client.Do(req)
				if err != nil {
					errorLogger.Println("Error making HTTP Range request to", fileURL, err)
					if retries < maxRetries - 1 {
						infoLogger.Println("Retrying chunk", currentChunkIndex, "...")
						time.Sleep(time.Second * time.Duration(retries+1)) // exponential back-off
						continue
					}
					return
				}

				tmpFileName := fmt.Sprintf("tmpfile_%d", currentChunkIndex)
				tmpFile, err := os.Create(tmpFileName)
				if err != nil {
					errorLogger.Println("Error creating temporary file", tmpFileName, err)
					return
				} else {
					infoLogger.Println("Created temporary file", tmpFileName)
				}

				_, err = io.Copy(tmpFile, resp.Body)
				if err != nil {
					errorLogger.Println("Error copying response body to temporary file", tmpFileName, err)
					tmpFile.Close()
					os.Remove(tmpFileName) // clean up
					if retries < maxRetries - 1 {
						infoLogger.Println("Retrying chunk", currentChunkIndex, "...")
						time.Sleep(time.Second * time.Duration(retries+1)) // exponential back-off
						continue
					}
					return
				} else {
					infoLogger.Println("Copied response body to temporary file", tmpFileName)
				}
				resp.Body.Close()
				tmpFile.Seek(0, 0) // reset file pointer to beginning

				partBytes, err := io.ReadAll(tmpFile)
				if err != nil {
					errorLogger.Println("Error reading from temporary file", tmpFileName, err)
					return
				} else {
					infoLogger.Println("Read temporary file to write hash", tmpFileName)
				}

				md5Hash.Write(partBytes) // update hash with file part

				tmpFile.Close()
				os.Remove(tmpFileName) // remove temp file
				break
			}
			if err != nil {
				errorLogger.Printf("Error downloading chunk %d: %v\n", currentChunkIndex, err)
			} else {
				infoLogger.Printf("Successfully downloaded chunk %d", currentChunkIndex)
			}
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
		infoLogger.Println("Created file", fileName)
	}
	defer file.Close()

	// TODO: Get this working
	// File hash: "db330cb3a5fbddee1f36f77f9f83670e" does not match the ETag: "md5:aab5e2178af3844b1ab801112ab748f6"
	calculatedMd5 := fmt.Sprintf(`"%x"`, md5Hash.Sum(nil))
	if calculatedMd5 == etag {
		infoLogger.Println("File hash:", calculatedMd5 ,"matches the ETag:", etag)
	} else {
		warningLogger.Println("File hash:", calculatedMd5 ,"does not match the ETag:", etag)
	}
}