package logic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

const (
	HTTPResponseHeaderContentType = "Content-Type"
	HTTPMetadataKeyContentType    = "content-type"
    maxParallel = 8          // number of goroutines
    chunkSize   = 5 * 1024 * 1024 // 5MB per chunk (IDM default 5–8MB)
)

type Downloader interface {
	Download(ctx context.Context, writer io.Writer) (map[string]any, error)
}

type HTTPDownloader struct {
	url    string
	logger *zap.Logger
}

func NewHTTPDownloader(
	url string,
	logger *zap.Logger,
) Downloader {
	return &HTTPDownloader{
		url:    url,
		logger: logger,
	}
}

func (h HTTPDownloader) Download(ctx context.Context, writer io.Writer) (map[string]any, error) {
	logger := utils.LoggerWithContext(ctx, h.logger)

	supportsRange, totalSize, err := utils.DetectRangeAndSize(ctx, http.DefaultClient, h.url)
	if err != nil {
		logger.With(zap.Error(err)).Warn("range detection failed, falling back to sequential")
		return h.sequentialDownload(ctx, writer, logger)
	}

	logger.Info("range detection",
		zap.Bool("supports_range", supportsRange),
		zap.Int64("total_size", totalSize),
	)

	if !supportsRange || totalSize <= 0 {
		logger.Info("range not supported or unknown size, falling back to sequential")
		return h.sequentialDownload(ctx, writer, logger)
	}

	if totalSize < 5*1024*1024 {
		logger.Info("file too small for parallel download, falling back to sequential")
		return h.sequentialDownload(ctx, writer, logger)
	}

	logger.Info("starting parallel range download")

	return h.parallelDownload(ctx, writer, totalSize, logger)
}


func (h HTTPDownloader) sequentialDownload(
	ctx context.Context,
	writer io.Writer,
	logger *zap.Logger,
) (map[string]any, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, http.NoBody)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create http request")
		return nil, err
	}

	start := time.Now()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to make http request")
		return nil, err
	}
	defer resp.Body.Close()

	// Large buffer = fewer syscalls, better performance
	buf := make([]byte, 512*1024)

	n, err := io.CopyBuffer(writer, resp.Body, buf)
	elapsed := time.Since(start)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to copy response body")
		return nil, err
	}

	logger.Info("sequential download completed",
		zap.Int64("bytes", n),
		zap.Duration("duration", elapsed),
		zap.Float64("speed_mb_s", float64(n)/elapsed.Seconds()/1024/1024),
	)

	metadata := map[string]any{
		HTTPMetadataKeyContentType: resp.Header.Get(HTTPResponseHeaderContentType),
		"download-duration-ms":     elapsed.Milliseconds(),
		"download-size-bytes":      n,
	}

	return metadata, nil
}

func (h HTTPDownloader) parallelDownload(
    ctx context.Context,
    writer io.Writer,
    totalSize int64,
    logger *zap.Logger,
) (map[string]any, error) {

    start := time.Now()

    chunks := int((totalSize + chunkSize - 1) / chunkSize)
    logger.Info("parallel streaming download",
        zap.Int("chunks", chunks),
        zap.Int64("total_size", totalSize),
    )

    type chunkResult struct {
        index int
        data  []byte
        err   error
    }

    // Workers push results here
    results := make(chan chunkResult, maxParallel)

    // 1. Start workers
    var wg sync.WaitGroup
    wg.Add(maxParallel)

    jobs := make(chan int, chunks)
    for i := 0; i < chunks; i++ {
        jobs <- i
    }
    close(jobs)

    for w := 0; w < maxParallel; w++ {
        go func() {
            defer wg.Done()
            httpClient := http.DefaultClient

            for idx := range jobs {

                startByte := int64(idx) * chunkSize
                endByte := min(startByte+chunkSize-1, totalSize-1)

                req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
                if err != nil {
                    results <- chunkResult{idx, nil, err}
                    return
                }
                req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, endByte))

                resp, err := httpClient.Do(req)
                if err != nil {
                    results <- chunkResult{idx, nil, err}
                    return
                }

                if resp.StatusCode != http.StatusPartialContent {
                    resp.Body.Close()
                    results <- chunkResult{idx, nil, fmt.Errorf("expected 206, got %d", resp.StatusCode)}
                    return
                }

                data, err := io.ReadAll(resp.Body)
                resp.Body.Close()
                if err != nil {
                    results <- chunkResult{idx, nil, err}
                    return
                }

                results <- chunkResult{index: idx, data: data}
            }
        }()
    }

    // Close results when workers end
    go func() {
        wg.Wait()
        close(results)
    }()

    // 2. Streaming ordered writer
    pending := make(map[int][]byte)
    next := 0
    var downloaded int64

    for res := range results {
        if res.err != nil {
            return nil, res.err
        }

        pending[res.index] = res.data

        // Write all ready chunks in order
        for {
            chunk, ok := pending[next]
            if !ok {
                break
            }

            n, err := writer.Write(chunk)
            if err != nil {
                return nil, err
            }
            downloaded += int64(n)

            delete(pending, next)
            next++
        }
    }

    elapsed := time.Since(start)

    logger.Info("parallel streaming download completed",
        zap.Int64("bytes", downloaded),
        zap.Duration("duration", elapsed),
        zap.Float64("speed_mb_s", float64(downloaded)/elapsed.Seconds()/1024/1024),
    )

    return map[string]any{
        "total_size":        totalSize,
        "downloaded_bytes":  downloaded,
        "duration_ms":       elapsed.Milliseconds(),
        "content_type":      "application/octet-stream",
    }, nil
}


