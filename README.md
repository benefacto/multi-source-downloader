# multi-source-downloader
A file download manager that can download a single file in parallel

## Optimization History & Future Enhancement

1. **Previous hacky version**: download file in 4 chunks directly into memory; this was very slow on my machine & network (~12 minutes)
2. **Previous semi-optimized version**: download file in 3 chunks into temp files; this was somewhat faster on my machine & network (~5 minutes)
3. **Current more-optimized version**: reuse HTTP client across chunks so that TCP connections can be reused; this was much faster on my machine & network and is largely equivalent to the average speed of my browser (my browser was occassionally significantly faster so my connection may be getting throttled by repeated downloads) download (~2 minutes)
4. **Future further optimized version**: utilize multiple connections, smart chunking, various types of caching, pipelining, and other techniques to approach the maximum download speed of a browser; this would be even faster on my machine & network (~1 minute)