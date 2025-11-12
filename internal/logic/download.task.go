package logic

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka/producer"
	go_idm_v1 "github.com/manhhung2111/go-idm/internal/generated/proto"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateDownloadTaskParams struct {
	Token        string
	DownloadType go_idm_v1.DownloadType
	URL          string
}

type CreateDownloadTaskOutput struct {
	DownloadTask *go_idm_v1.DownloadTask
}

type GetDownloadTaskListParams struct {
	Token  string
	Offset uint64
	Limit  uint64
}

type GetDownloadTaskListOutput struct {
	DownloadTaskList []*go_idm_v1.DownloadTask
	Total            uint64
}

type UpdateDownloadTaskParams struct {
	Token          string
	DownloadTaskId uint64
	URL            string
}

type UpdateDownloadTaskOutput struct {
	DownloadTask *go_idm_v1.DownloadTask
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
	DeleteDownloadTask(ctx context.Context, params DeleteDownloadTaskParams) error
}

type downloadTask struct {
	tokenLogic                  Token
	accountDataAccessor         database.AccountDataAccessor
	downloadTaskDataAccessor    database.DownloadTaskDataAccessor
	goquDatabase                *goqu.Database
	logger                      *zap.Logger
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer
}

func NewDownloadTask(
	tokenLogic Token,
	accountDataAccessor database.AccountDataAccessor,
	downloadTaskDataAccessor database.DownloadTaskDataAccessor,
	goquDatabase *goqu.Database,
	logger *zap.Logger,
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer,
) DownloadTask {
	return &downloadTask{
		tokenLogic:                  tokenLogic,
		accountDataAccessor:         accountDataAccessor,
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

	account, err := d.accountDataAccessor.GetAccountById(ctx, accountId)
	if err != nil {
		return CreateDownloadTaskOutput{}, err
	}

	downloadTask := database.DownloadTask{
		OfAccountID:    account.ID,
		DownloadType:   params.DownloadType,
		URL:            params.URL,
		DownloadStatus: go_idm_v1.DownloadStatus_Pending,
		Metadata:       database.JSON{
			Data: make(map[string]any),
		},
	}

	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTaskId, err := d.downloadTaskDataAccessor.CreateDownloadTask(ctx, downloadTask)
		if err != nil {
			return err
		}

		downloadTask.ID = downloadTaskId
		if err = d.downloadTaskCreatedProducer.Send(ctx, producer.DownloadTaskCreated{
			Id: downloadTaskId,
		}); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return CreateDownloadTaskOutput{}, txErr
	}

	return CreateDownloadTaskOutput{
		DownloadTask: d.databaseDownloadTaskToProtoDownloadTask(downloadTask, account),
	}, nil
}

// DeleteDownloadTask implements DownloadTask.
func (d *downloadTask) DeleteDownloadTask(ctx context.Context, params DeleteDownloadTaskParams) error {
	accountId, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return err
	}

	return d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTask, getDownloadTaskWithXLockErr := d.downloadTaskDataAccessor.WithDatabase(td).
			GetDownloadTaskWithXLock(ctx, params.DownloadTaskId)
		if getDownloadTaskWithXLockErr != nil {
			return getDownloadTaskWithXLockErr
		}

		if downloadTask.OfAccountID != accountId {
			return status.Error(codes.PermissionDenied, "trying to delete a download task the account does not own")
		}

		return d.downloadTaskDataAccessor.WithDatabase(td).DeleteDownloadTask(ctx, params.DownloadTaskId)
	})
}

// GetDownloadTaskList implements DownloadTask.
func (d *downloadTask) GetDownloadTaskList(ctx context.Context, params GetDownloadTaskListParams) (GetDownloadTaskListOutput, error) {
	accountId, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	account, err := d.accountDataAccessor.GetAccountById(ctx, accountId)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	totalDownloadTaskCount, err := d.downloadTaskDataAccessor.GetDownloadTaskCountOfAccount(ctx, accountId)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	downloadTaskList, err := d.downloadTaskDataAccessor.
		GetDownloadTaskListOfAccount(ctx, accountId, params.Offset, params.Limit)
	if err != nil {
		return GetDownloadTaskListOutput{}, err
	}

	return GetDownloadTaskListOutput{
		Total: totalDownloadTaskCount,
		DownloadTaskList: lo.Map(downloadTaskList, func(item database.DownloadTask, index int) *go_idm_v1.DownloadTask {
			return d.databaseDownloadTaskToProtoDownloadTask(item, account)
		}),
	}, nil
}

// UpdateDownloadTask implements DownloadTask.
func (d *downloadTask) UpdateDownloadTask(ctx context.Context, params UpdateDownloadTaskParams) (UpdateDownloadTaskOutput, error) {
	accountId, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return UpdateDownloadTaskOutput{}, err
	}

	account, err := d.accountDataAccessor.GetAccountById(ctx, accountId)
	if err != nil {
		return UpdateDownloadTaskOutput{}, err
	}

	output := UpdateDownloadTaskOutput{}
	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTask, getDownloadTaskWithXLockErr := d.downloadTaskDataAccessor.WithDatabase(td).
			GetDownloadTaskWithXLock(ctx, params.DownloadTaskId)
		if getDownloadTaskWithXLockErr != nil {
			return getDownloadTaskWithXLockErr
		}

		if downloadTask.OfAccountID != accountId {
			return status.Error(codes.PermissionDenied, "trying to update a download task the account does not own")
		}

		downloadTask.URL = params.URL
		output.DownloadTask = d.databaseDownloadTaskToProtoDownloadTask(downloadTask, account)
		return d.downloadTaskDataAccessor.WithDatabase(td).UpdateDownloadTask(ctx, downloadTask)
	})
	if txErr != nil {
		return UpdateDownloadTaskOutput{}, txErr
	}

	return output, nil
}

func (d downloadTask) databaseDownloadTaskToProtoDownloadTask(
	downloadTask database.DownloadTask,
	account database.Account,
) *go_idm_v1.DownloadTask {
	return &go_idm_v1.DownloadTask{
		Id:             downloadTask.ID,
		DownloadType:   downloadTask.DownloadType,
		Url:            downloadTask.URL,
		DownloadStatus: go_idm_v1.DownloadStatus_Pending,
	}
}
