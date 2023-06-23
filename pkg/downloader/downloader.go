// Package downloader implements methods for downloading a file in parallel from a single source.
package downloader

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/benefacto/multi-source-downloader/pkg/logger"
	"github.com/hashicorp/go-multierror"
)

// DownloadParams defines the parameters needed for the download operation.
type DownloadParams struct {
	URL            string
	FileExtension  string
	ChunkSize      int
	MaxRetries     int
	NumberOfChunks int
}

// DownloadFile downloads a file using the specified parameters and logs events with the given logger.
// It returns the file name and any error encountered.
func DownloadFile(ctx context.Context, params DownloadParams, logger logger.Logger) (string, error) {
	resp, err := http.Head(params.URL)
	if err != nil {
		logger.Error("Error making HTTP HEAD request to", params.URL, err)
		return "", err
	}
	logger.Info("Successfully made HTTP HEAD request to", params.URL)

	etag := resp.Header.Get("Etag")
	fileSize, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		logger.Error("Error getting Content-Length header for", params.URL, err)
		return "", err
	}
	logger.Info("File is", fileSize, "bytes with ETag:", etag)

	chunkSize := int(fileSize) / params.NumberOfChunks
	remainingBytes := int(fileSize) % params.NumberOfChunks

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	tempFiles := make([]string, params.NumberOfChunks) // to store names of temporary files
	var merr *multierror.Error
	merrMux := &sync.Mutex{}

	// Use a single http.Client for all requests
	client := &http.Client{}

	// Ensure output directory exists
	if err := os.MkdirAll("./output", 0755); err != nil {
		logger.Error("Error creating output directory", err)
		return "", err
	}

	// Create a channel to feed worker routines.
	chunkIndexChannel := make(chan int, params.NumberOfChunks)
	for i := 0; i < params.NumberOfChunks; i++ {
		chunkIndexChannel <- i
	}
	close(chunkIndexChannel)

	// Create worker routines
	for i := 0; i < params.NumberOfChunks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for currentChunkIndex := range chunkIndexChannel {
				err := downloadChunk(currentChunkIndex, chunkSize, remainingBytes, params, ctx, client, merr, merrMux, cancel, tempFiles, logger)
				if err != nil {
					merrMux.Lock()
					merr = multierror.Append(merr, err)
					merrMux.Unlock()
					logger.Error("Error downloading chunk", currentChunkIndex, err)
					cancel() // cancel all operations when an error is encountered
					return
				}
			}
		}()
	}

	wg.Wait()
	if merr != nil {
		return "", merr.ErrorOrNil()
	}

	t := time.Now()
	timestamp := t.Format("20060102_150405")
	fileName := fmt.Sprintf("./output/output_%s.%s", timestamp, params.FileExtension)
	file, err := os.Create(fileName)
	if err != nil {
		logger.Error("Error creating file", fileName, err)
		return "", err
	}
	defer file.Close()

	err = mergeFiles(tempFiles, etag, file, logger)
	if err != nil {
		logger.Error("Error merging files", err)
		return "", err
	}

	// delete temporary files after merging them into the final file
	for _, tmpFile := range tempFiles {
		os.Remove(tmpFile)
	}

	return fileName, nil
}

// downloadChunk is a helper function for downloading a file chunk.
// It also handles retries if a download fails.
func downloadChunk(currentChunkIndex, chunkSize, remainingBytes int, params DownloadParams, ctx context.Context, client *http.Client, merr *multierror.Error, merrMux *sync.Mutex, cancel context.CancelFunc, tempFiles []string, logger logger.Logger) error {
	var lastErr error
	for retries := 0; retries < params.MaxRetries; retries++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			start := currentChunkIndex * chunkSize
			end := start + chunkSize - 1
			if currentChunkIndex == params.NumberOfChunks-1 {
				end += remainingBytes
			}

			req, err := http.NewRequestWithContext(ctx, "GET", params.URL, nil) // context provided to network call
			if err != nil {
				logger.Error("Chunk", currentChunkIndex, "had an error making HTTP GET request to", params.URL, err)
				return err
			}
			logger.Info("Chunk", currentChunkIndex, "successfully made HTTP GET request to", params.URL)

			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
			req.Header.Add("Range", rangeHeader)

			resp, err := client.Do(req)
			if err != nil {
				logger.Error("Chunk", currentChunkIndex, "had an error making HTTP Range request to", params.URL, err)
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() && retries < params.MaxRetries-1 {
					logger.Info("Retrying chunk", currentChunkIndex, "...")
					lastErr = err
					time.Sleep(time.Second * time.Duration(retries+1)) // exponential back-off
					continue
				}
				return err
			}
			defer resp.Body.Close()

			tmpFileName := fmt.Sprintf("./output/tmpfile_%d", currentChunkIndex)
			tempFiles[currentChunkIndex] = tmpFileName // save name of temp file
			tmpFile, err := os.Create(tmpFileName)
			if err != nil {
				logger.Error("Chunk", currentChunkIndex, "had an error creating temporary file", tmpFileName, err)
				return err
			}
			defer tmpFile.Close()

			_, err = io.Copy(tmpFile, resp.Body)
			if err != nil {
				logger.Error("Chunk", currentChunkIndex, "had an error writing to temporary file", tmpFileName, err)
				return err
			}
			logger.Info("Chunk", currentChunkIndex, "successfully wrote to temporary file", tmpFileName)

			return nil
		}
	}
	return lastErr
}

// mergeFiles merges temporary files into a single output file.
// It also verifies the integrity of the download by comparing the MD5 hash of the output file with the ETag received from the server.
func mergeFiles(tempFiles []string, etag string, outputFile *os.File, logger logger.Logger) error {
	md5Hash := md5.New()
	for _, tmpFileName := range tempFiles {
		tmpFile, err := os.Open(tmpFileName)
		if err != nil {
			logger.Error("Error opening temporary file", tmpFileName, err)
			return err
		}
		defer tmpFile.Close()
		logger.Info("Successfully opened temporary file", tmpFileName)

		// Write contents of tmpFile to outputFile and md5Hash
		if _, err := io.Copy(outputFile, io.TeeReader(tmpFile, md5Hash)); err != nil {
			logger.Error("Error writing temporary file", tmpFileName, "to output file", err)
			return err
		}
		logger.Info("Successfully wrote temporary file", tmpFileName, "to output file")
	}

	// Check if the downloaded file's md5 hash matches the server's ETag
	etagHash := fmt.Sprintf(`"md5:%x"`, md5Hash.Sum(nil))
	if etagHash != etag {
		return fmt.Errorf("file integrity check failed: MD5 hash %v does not match ETag %v", etagHash, etag)
	}
	logger.Info("File integrity check passed: MD5 hash matches ETag")
	return nil
}
