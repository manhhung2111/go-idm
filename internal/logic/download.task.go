package logic

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/producer"
	go_idm_v1 "github.com/manhhung2111/go-idm/internal/generated/proto"
	"go.uber.org/zap"
)

type CreateDownloadTaskParams struct {
	Token        string
	DownloadType go_idm_v1.DownloadType
	URL          string
}

type CreateDownloadTaskOutput struct {
	DownloadTask go_idm_v1.DownloadTask
}

type GetDownloadTaskListParams struct {
	Token  string
	Offset uint64
	Limit  uint64
}

type GetDownloadTaskListOutput struct {
	DownloadTaskList []go_idm_v1.DownloadTask
	Total            uint64
}

type UpdateDownloadTaskParams struct {
	Token          string
	DownloadTaskId uint64
	URL            string
}

type UpdateDownloadTaskOutput struct {
	DownloadTask go_idm_v1.DownloadTask
}

type DeleteDownloadTaskParams struct {
	Token          string
	DownloadTaskId uint64
}

type DeleteDownloadTaskOutput struct{}

type DownloadTask interface {
	CreateDownloadTask(ctx context.Context, params CreateDownloadTaskParams) (CreateDownloadTaskOutput, error)
	GetDownloadTaskList(ctx context.Context, params GetDownloadTaskListParams) (GetDownloadTaskListOutput, error)
	UpdateDownloadTask(ctx context.Context, params UpdateDownloadTaskParams) (UpdateDownloadTaskOutput, error)
	DeleteDownloadTask(ctx context.Context, params DeleteDownloadTaskParams) (DeleteDownloadTaskOutput, error)
}

type downloadTask struct {
	tokenLogic                  Token
	downloadTaskDataAccessor    database.DownloadTaskDataAccessor
	goquDatabase                *goqu.Database
	logger                      *zap.Logger
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer
}

func NewDownloadTask(
	tokenLogic Token,
	downloadTaskDataAccessor database.DownloadTaskDataAccessor,
	goquDatabase *goqu.Database,
	logger *zap.Logger,
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer,
) DownloadTask {
	return &downloadTask{
		tokenLogic:                  tokenLogic,
		downloadTaskDataAccessor:    downloadTaskDataAccessor,
		goquDatabase:                goquDatabase,
		logger:                      logger,
		downloadTaskCreatedProducer: downloadTaskCreatedProducer,
	}
}

// CreateDownloadTask implements DownloadTask.
func (d *downloadTask) CreateDownloadTask(ctx context.Context, params CreateDownloadTaskParams) (CreateDownloadTaskOutput, error) {
	accountId, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return CreateDownloadTaskOutput{}, err
	}

	downloadTask := database.DownloadTask{
		OfAccountID:    accountId,
		DownloadType:   params.DownloadType,
		URL:            params.URL,
		DownloadStatus: go_idm_v1.DownloadStatus_Pending,
		Metadata:       "{}",
	}

	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTaskId, err := d.downloadTaskDataAccessor.CreateDownloadTask(ctx, downloadTask)
		if err != nil {
			return err
		}

		downloadTask.ID = downloadTaskId
		if err = d.downloadTaskCreatedProducer.Send(ctx, producer.DownloadTaskCreated{
			DownloadTask: downloadTask,
		}); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return CreateDownloadTaskOutput{}, txErr
	}

	return CreateDownloadTaskOutput{
		DownloadTask: go_idm_v1.DownloadTask{
			Id:             downloadTask.ID,
			DownloadType:   downloadTask.DownloadType,
			Url:            downloadTask.URL,
			DownloadStatus: downloadTask.DownloadStatus,
		},
	}, nil
}

// DeleteDownloadTask implements DownloadTask.
func (d *downloadTask) DeleteDownloadTask(ctx context.Context, params DeleteDownloadTaskParams) (DeleteDownloadTaskOutput, error) {
	panic("unimplemented")
}

// GetDownloadTaskList implements DownloadTask.
func (d *downloadTask) GetDownloadTaskList(ctx context.Context, params GetDownloadTaskListParams) (GetDownloadTaskListOutput, error) {
	panic("unimplemented")
}

// UpdateDownloadTask implements DownloadTask.
func (d *downloadTask) UpdateDownloadTask(ctx context.Context, params UpdateDownloadTaskParams) (UpdateDownloadTaskOutput, error) {
	panic("unimplemented")
}
