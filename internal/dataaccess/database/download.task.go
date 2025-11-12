package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	go_idm_v1 "github.com/manhhung2111/go-idm/internal/generated/proto"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	tableNameDownloadTasks = goqu.T("download_tasks")
)

const (
	ColNameDownloadTaskId             = "id"
	ColNameDownloadTaskOfAccountId    = "of_account_id"
	ColNameDownloadTaskDownloadType   = "download_type"
	ColNameDownloadTaskURL            = "url"
	ColNameDownloadTaskDownloadStatus = "download_status"
	ColNameDownloadTaskMetadata       = "metadata"
)

type DownloadTask struct {
	ID             uint64                   `db:"id" goqu:"skipinsert,skipupdate"`
	OfAccountID    uint64                   `db:"of_account_id" goqu:"skipinsert,skipupdate"`
	DownloadType   go_idm_v1.DownloadType   `db:"download_type"`
	URL            string                   `db:"url"`
	DownloadStatus go_idm_v1.DownloadStatus `db:"download_status"`
	Metadata       JSON                     `db:"metadata"`
}

type DownloadTaskDataAccessor interface {
	CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error)
	GetDownloadTaskListOfAccount(ctx context.Context, accountId, offset, limit uint64) ([]DownloadTask, error)
	GetDownloadTaskCountOfAccount(ctx context.Context, accountId uint64) (uint64, error)
	GetDownloadTask(ctx context.Context, id uint64) (DownloadTask, error)
	GetDownloadTaskWithXLock(ctx context.Context, id uint64) (DownloadTask, error)
	UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask) error
	DeleteDownloadTask(ctx context.Context, id uint64) error
	WithDatabase(database IDatabase) DownloadTaskDataAccessor
}

type downloadTaskDataAccessor struct {
	database IDatabase
	logger   *zap.Logger
}

func NewDownloadTaskDataAccessor(
	database *goqu.Database,
	logger *zap.Logger,
) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
		logger:   logger,
	}
}

// CreateDownloadTask implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) CreateDownloadTask(ctx context.Context, downloadTask DownloadTask) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("downloadTask", downloadTask))

	result, err := d.database.
		Insert(tableNameDownloadTasks).
		Rows(downloadTask).
		Executor().
		ExecContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create download task")
		return 0, status.Errorf(codes.Internal, "failed to create download task")
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get last inserted id")
		return 0, status.Errorf(codes.Internal, "failed to get last inserted id")
	}

	return uint64(lastInsertedId), nil
}

// DeleteDownloadTask implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) DeleteDownloadTask(ctx context.Context, id uint64) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	if _, err := d.database.
		Delete(tableNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskId: id}).
		Executor().
		ExecContext(ctx); err != nil {
		logger.With(zap.Error(err)).Error("failed to delete download task")
		return status.Errorf(codes.Internal, "failed to delete download task")
	}

	return nil
}

func (d downloadTaskDataAccessor) GetDownloadTaskCountOfAccount(ctx context.Context, accountId uint64) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("account_id", accountId))

	count, err := d.database.
		From(tableNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskOfAccountId: accountId}).
		CountContext(ctx)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to count download task of user")
		return 0, status.Errorf(codes.Internal, "failed to count download task of user")
	}

	return uint64(count), nil
}

// GetDownloadTaskList implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) GetDownloadTaskListOfAccount(ctx context.Context, accountId uint64, offset uint64, limit uint64) ([]DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).
		With(zap.Uint64("account_id", accountId)).
		With(zap.Uint64("offset", offset)).
		With(zap.Uint64("limit", limit))

	downloadTaskList := make([]DownloadTask, 0)
	if err := d.database.
		Select().
		From(tableNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskOfAccountId: accountId}).
		Offset(uint(offset)).
		Limit(uint(limit)).
		Executor().
		ScanStructsContext(ctx, &downloadTaskList); err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task list of account")
		return nil, status.Errorf(codes.Internal, "failed to get download task list of account")
	}

	return downloadTaskList, nil
}

func (d downloadTaskDataAccessor) GetDownloadTask(ctx context.Context, id uint64) (DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	downloadTask := DownloadTask{}
	found, err := d.database.
		Select().
		From(tableNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskId: id}).
		ScanStructContext(ctx, &downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task")
		return DownloadTask{}, status.Errorf(codes.Internal, "failed to get download task list of account")
	}

	if !found {
		logger.Error("download task not found")
		return DownloadTask{}, status.Error(codes.NotFound, "download task not found")
	}

	return downloadTask, nil
}

func (d downloadTaskDataAccessor) GetDownloadTaskWithXLock(ctx context.Context, id uint64) (DownloadTask, error) {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Uint64("id", id))

	downloadTask := DownloadTask{}
	found, err := d.database.
		Select().
		From(tableNameDownloadTasks).
		Where(goqu.Ex{ColNameDownloadTaskId: id}).
		ForUpdate(goqu.Wait).
		ScanStructContext(ctx, &downloadTask)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get download task")
		return DownloadTask{}, status.Errorf(codes.Internal, "failed to get download task list of account")
	}

	if !found {
		logger.Error("download task not found")
		return DownloadTask{}, status.Error(codes.NotFound, "download task not found")
	}

	return downloadTask, nil
}

// UpdateDownloadTask implements DownloadTaskDataAccessor.
func (d *downloadTaskDataAccessor) UpdateDownloadTask(ctx context.Context, downloadTask DownloadTask)  error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("task", downloadTask))

	if _, err := d.database.
		Update(tableNameDownloadTasks).
		Set(downloadTask).
		Where(goqu.Ex{ColNameDownloadTaskId: downloadTask.ID}).
		Executor().
		ExecContext(ctx); err != nil {
		logger.With(zap.Error(err)).Error("failed to update download task")
		return status.Errorf(codes.Internal, "failed to update download task")
	}

	return nil
}

func (d *downloadTaskDataAccessor) WithDatabase(database IDatabase) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
		logger: d.logger,
	}
}
