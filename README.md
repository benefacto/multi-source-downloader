# Multi-source Downloader

The multi-source downloader is a Golang application designed to download [a single file](https://zenodo.org/record/4435114/files/supplement.csv?download=1) in parallel chunks from a public web server. The purpose is to emulate the functionality of protocols like BitTorrent where a file is downloaded in portions from multiple sources. However, this application downloads different parts of the same file from a single source. It is capable of handling large files (60MB or more) and doesn't need to know the file size beforehand.

The application also has the added functionality of verifying the downloaded file if the server returns an `etag` in a known hash format (assumed to be MD5). This isn't a requirement, but it's a good-to-have feature to ensure the integrity of the downloaded file.

## Directory Structure

```
.
├── cmd
│   └── main.go
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── pkg
│   ├── downloader
│   │   ├── downloader.go
│   │   └── downloader_test.go
│   └── logger
│       └── logger.go
└── README.md
```

## How to Run

You can easily run the downloader using Golang or Docker:

- **Using Golang**:

    1. Clone the repository to your local machine.
    2. Navigate to the root directory of the project.
    3. Run the command: `go run cmd/main.go`

- **Using Docker**:

    1. Build the Docker image with: `docker-compose build`
    2. Start the service with: `docker-compose up`

To run tests, navigate to the root directory of the project and run `go test ./...`

## Optimization History & Future Enhancement

The application has undergone several optimization phases and future enhancements are planned:

1. **Previous hacky version**: The initial implementation downloaded the file in 4 chunks directly into memory, which proved to be quite slow (around 12 minutes per file).
2. **Previous semi-optimized version**: The next iteration downloaded the file in 3 chunks into temporary files. This version was somewhat faster (approximately 5 minutes per file).
3. **Current more-optimized version**: The current version reuses the HTTP client across chunks, enabling TCP connections to be reused. This implementation is much faster and is largely equivalent to the average speed of my browser's download capabilities (around 2 minutes per file).
4. **Future further optimized version**: The next planned iteration will utilize multiple connections, smart chunking, various types of caching, pipelining, and other techniques to approach the maximum download speed of a browser. It is expected that this version will further improve the download speed (targeting around 1 minute per file).

## License

This project is licensed under the terms of the Apache License Version 2.0. For more information, see the `LICENSE` file.