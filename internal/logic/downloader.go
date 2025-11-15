package logic

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

const (
	HTTPResponseHeaderContentType = "Content-Type"
	HTTPMetadataKeyContentType    = "content-type"
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, http.NoBody)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create http request")
		return nil, err
	}

	start := time.Now() // ← start measuring

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to make http request")
		return nil, err
	}
	defer resp.Body.Close()

	// measure from first byte → last byte copied
	n, err := io.Copy(writer, resp.Body)

	elapsed := time.Since(start) // ← end measurement

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to copy response body")
		return nil, err
	}

	logger.Info("download completed",
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
