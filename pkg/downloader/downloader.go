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

type DownloadParams struct {
	URL            string
	FileExtension  string
	ChunkSize      int
	MaxRetries     int
	NumberOfChunks int
}

func DownloadFile(params DownloadParams, logger logger.Logger) (string, error) {
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	tempFiles := make([]string, params.NumberOfChunks) // to store names of temporary files
	var merr *multierror.Error
	merrMux := &sync.Mutex{}

	client := &http.Client{}

	for currentChunkIndex := 0; currentChunkIndex < params.NumberOfChunks; currentChunkIndex++ {
		wg.Add(1)
		go func(currentChunkIndex int, client *http.Client) {
			defer wg.Done()
			err := downloadChunk(currentChunkIndex, chunkSize, remainingBytes, params, ctx, client, merr, merrMux, cancel, tempFiles, logger)
			if err != nil {
				merrMux.Lock()
				merr = multierror.Append(merr, err)
				merrMux.Unlock()
				logger.Error("Error downloading chunk", currentChunkIndex, err)
			}
		}(currentChunkIndex, client)
	}

	wg.Wait()
	if merr != nil {
		return "", merr.ErrorOrNil()
	}

	// Ensure output directory exists
	if err := os.MkdirAll("./output", 0755); err != nil {
		logger.Error("Error creating output directory", err)
		return "", err
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

	return fileName, nil
}

func downloadChunk(currentChunkIndex, chunkSize, remainingBytes int, params DownloadParams, ctx context.Context, client *http.Client, merr *multierror.Error, merrMux *sync.Mutex, cancel context.CancelFunc, tempFiles []string, logger logger.Logger) error {
	var lastErr error
	for retries := 0; retries < params.MaxRetries; retries++ {
		start := currentChunkIndex * chunkSize
		end := start + chunkSize - 1
		if currentChunkIndex == params.NumberOfChunks-1 {
			end += remainingBytes
		}

		req, err := http.NewRequest("GET", params.URL, nil)
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

		tmpFileName := fmt.Sprintf("tmpfile_%d", currentChunkIndex)
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

		break
	}
	return lastErr
}

func mergeFiles(tempFiles []string, etag string, outputFile *os.File, logger logger.Logger) error {
	md5HashTempFiles := md5.New()
	for _, tmpFileName := range tempFiles {
		tmpFile, err := os.Open(tmpFileName)
		if err != nil {
			logger.Error("Error opening temporary file", tmpFileName, err)
			return err
		}
		defer tmpFile.Close()
		logger.Info("Successfully opened temporary file", tmpFileName)

		// Write contents of tmpFile to outputFile and md5HashTempFiles
		if _, err := io.Copy(outputFile, io.TeeReader(tmpFile, md5HashTempFiles)); err != nil {
			logger.Error("Error writing content of temporary file to output file", tmpFileName, err)
			return err
		} else {
			logger.Info("Successfully wrote content of temporary file to output file", tmpFileName)
		}

		os.Remove(tmpFileName)
	}

	checksumTempFiles := fmt.Sprintf(`"md5:%x"`, md5HashTempFiles.Sum(nil))
	if checksumTempFiles != etag {
		logger.Warning("MD5 checksum of temp files", checksumTempFiles, "does not match ETag", etag)
	} else {
		logger.Info("MD5 checksum of temp files", checksumTempFiles, "matches ETag", etag)
	}

	// Reopen the output file and calculate its MD5 checksum
	outputFile.Seek(0, 0)
	md5HashFinalFile := md5.New()
	if _, err := io.Copy(md5HashFinalFile, outputFile); err != nil {
		logger.Error("Error calculating MD5 checksum of output file", err)
		return err
	}

	checksumFinalFile := fmt.Sprintf(`"md5:%x"`, md5HashFinalFile.Sum(nil))
	if checksumFinalFile != etag {
		logger.Warning("MD5 checksum of output file", checksumFinalFile, "does not match ETag", etag)
	} else {
		logger.Info("MD5 checksum of output file", checksumFinalFile, "matches ETag", etag)
	}

	return nil
}
