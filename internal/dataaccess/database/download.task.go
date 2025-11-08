package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	go_idm_v1 "github.com/manhhung2111/go-idm/internal/generated/proto"
	"go.uber.org/zap"
)

type DownloadTask struct {
	ID             uint64                   `db:"id"`
	OfAccountID    uint64                   `db:"of_account_id"`
	DownloadType   go_idm_v1.DownloadType   `db:"download_type"`
	URL            string                   `db:"url"`
	DownloadStatus go_idm_v1.DownloadStatus `db:"download_status"`
	Metadata       string                   `db:"metadata"`
}

type DownloadTaskDataAccessor interface {
	CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error)
	GetDownloadTaskList(ctx context.Context, accountId, offset, limit uint64) ([]DownloadTask, error)
	UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error)
	DeleteDownloadTask(ctx context.Context, id uint64) error
	WithDatabase(database IDatabase) DownloadTaskDataAccessor
}

type downloadTaskDataAccessor struct {
	database IDatabase
	logger *zap.Logger
}

func NewDownloadTaskDataAccessor(
	database *goqu.Database,
	logger *zap.Logger,
) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
		logger: logger,
	}
}

// CreateDownloadTask implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error) {
	return 1, nil
}

// DeleteDownloadTask implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) DeleteDownloadTask(ctx context.Context, id uint64) error {
	panic("unimplemented")
}

// GetDownloadTaskList implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) GetDownloadTaskList(ctx context.Context, accountId uint64, offset uint64, limit uint64) ([]DownloadTask, error) {
	panic("unimplemented")
}

// UpdateDownloadTask implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error) {
	panic("unimplemented")
}

func (d *downloadTaskDataAccessor) WithDatabase(database IDatabase) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
	}
}

