package file

import (
	"context"
	"fmt"
	"io"

	"github.com/manhhung2111/go-idm/internal/config"
	"go.uber.org/zap"
)

type Client interface {
	Write(ctx context.Context, filePath string) (io.WriteCloser, error)
	Read(ctx context.Context, filePath string) (io.ReadCloser, error)
}

func NewClient(
	downloadConfig config.Download,
	logger *zap.Logger,
) (Client, error) {
	switch downloadConfig.Mode {
	case config.DownloadModeLocal:
		return NewLocalClient(downloadConfig, logger)
	case config.DownloadModeS3:
		return NewS3Client(downloadConfig, logger)
	default:
		return nil, fmt.Errorf("unsupported download mode: %s", downloadConfig.Mode)
	}
}