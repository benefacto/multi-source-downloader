package downloader

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/benefacto/multi-source-downloader/logger"
)

type DownloadParams struct {
	URL           string
	ChunkSize     int
	MaxRetries    int
	NumberOfChunks int
}

func DownloadFile(params DownloadParams, logger logger.Logger) error {
	resp, err := http.Head(params.URL)
	if err != nil {
		logger.Error("Error making HTTP HEAD request to", params.URL, err)
		return err
	}
	logger.Info("Successfully made HTTP HEAD request to", params.URL)

	etag := resp.Header.Get("Etag")
	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		logger.Error("Error getting length of ETag header", etag, "for", params.URL)
		return err
	}
	logger.Info("File is", fileSize, "bytes with ETag:", etag)

	chunkSize := fileSize / params.NumberOfChunks
	remainingBytes := fileSize % params.NumberOfChunks

	var wg sync.WaitGroup
	md5Hash := md5.New()
	tempFiles := make([]string, params.NumberOfChunks) // to store names of temporary files
	client := &http.Client{}

	for currentChunkIndex := 0; currentChunkIndex < params.NumberOfChunks; currentChunkIndex++ {
		wg.Add(1)
		go func(currentChunkIndex int, client *http.Client) {
			defer wg.Done()

			var err error
			for retries := 0; retries < params.MaxRetries; retries++ {
				start := currentChunkIndex * chunkSize
				end := start + chunkSize - 1
				if currentChunkIndex == params.NumberOfChunks - 1 {
					end += remainingBytes
				}

				req, err := http.NewRequest("GET", params.URL, nil)
				if err != nil {
					logger.Error("Chunk", currentChunkIndex, "had an error making HTTP GET request to", params.URL, err)
					return
				} else {
					logger.Info("Chunk", currentChunkIndex, "successfully made HTTP GET request to", params.URL)
				}

				rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
				req.Header.Add("Range", rangeHeader)

				resp, err := client.Do(req)
				if err != nil {
					logger.Error("Chunk", currentChunkIndex, "had an error making HTTP Range request to", params.URL, err)
					if retries < params.MaxRetries - 1 {
						logger.Info("Retrying chunk", currentChunkIndex, "...")
						time.Sleep(time.Second * time.Duration(retries+1)) // exponential back-off
						continue
					}
					return
				}

				tmpFileName := fmt.Sprintf("tmpfile_%d", currentChunkIndex)
				tempFiles[currentChunkIndex] = tmpFileName // save name of temp file
				tmpFile, err := os.Create(tmpFileName)
				if err != nil {
					logger.Error("Chunk", currentChunkIndex, "had an error creating temporary file", tmpFileName, err)
					return
				} else {
					logger.Info("Chunk", currentChunkIndex, "successfully created temporary file", tmpFileName)
				}

				_, err = io.Copy(tmpFile, resp.Body)
				if err != nil {
					logger.Error("Chunk", currentChunkIndex, "had an error writing to temporary file", tmpFileName, err)
					return
				} else {
					logger.Info("Chunk", currentChunkIndex, "successfully wrote to temporary file", tmpFileName)
				}

				resp.Body.Close()
				tmpFile.Close()
				break
			}

			if err != nil {
				logger.Error("Chunk", currentChunkIndex, "had an error after retries", err)
			}
		}(currentChunkIndex, client)
	}

	wg.Wait()

	for _, tmpFileName := range tempFiles {
		tmpFile, err := os.Open(tmpFileName)
		if err != nil {
			logger.Error("Error opening temporary file", tmpFileName, err)
			return err
		} else {
			logger.Info("Successfully opened temporary file", tmpFileName)
		}

		_, err = io.Copy(md5Hash, tmpFile)
		if err != nil {
			logger.Error("Error hashing content of temporary file", tmpFileName, err)
			return err
		} else {
			logger.Info("Successfully hashed content of temporary file", tmpFileName)
		}

		tmpFile.Close()
		os.Remove(tmpFileName)
	}

	checksum := fmt.Sprintf(`"md5:%x"`, md5Hash.Sum(nil))
	if checksum != etag {
		logger.Warning("MD5 checksum", checksum, "does not match ETag", etag)
	} else {
		logger.Info("MD5 checksum", checksum, "matches ETag", etag)
	}
	
	return nil
}