module github.com/benefacto/multi-source-downloader

replace (
	github.com/benefacto/multi-source-downloader/downloader => ./pkg/downloader
	github.com/benefacto/multi-source-downloader/logger => ./pkg/logger
)

go 1.20

require github.com/benefacto/multi-source-downloader/downloader v0.0.0-00010101000000-000000000000

require github.com/benefacto/multi-source-downloader/logger v0.0.0-00010101000000-000000000000
