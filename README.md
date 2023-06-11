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
        └── logger_test.go
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

1. **Previous hacky version**: The initial implementation downloaded the file in 4 chunks directly into memory. It was relatively slow with the average download time clocking around 12 minutes per file, depending on the connection speed, server response time, and the host machine's performance.
2. **Previous semi-optimized version**: The next iteration downloaded the file in 3 chunks into temporary files. This version improved the speed, reducing the average download time to approximately 5 minutes per file. However, the actual times may still vary based on the connection, server, and host machine.
3. **Current more-optimized version**: The current version reuses the HTTP client across chunks, enabling TCP connections to be reused. This has significantly boosted the download speed, bringing it on par with the average speed of a typical browser's download capabilities, at around 2 minutes per file. This time, too, can fluctuate depending on the connection speed, server response time, and the host machine's performance.
4. **Future further optimized version**: The next planned iteration will utilize multiple connections, smart chunking, various types of caching, pipelining, and other techniques to approach the maximum download speed of a browser. The aim is to further optimize the download speed, targeting around 1 minute per file. However, it is essential to note that these times will continue to depend on factors such as connection speed, server response times, and host machine performance.

## License

This project is licensed under the terms of the Apache License Version 2.0. For more information, see the `LICENSE` file.