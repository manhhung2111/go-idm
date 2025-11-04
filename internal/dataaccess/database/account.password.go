package database

import (
	"context"

	"database/sql"
	"github.com/doug-martin/goqu/v9"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	tableNameAccountPasswords            = "account_passwords"
	colNameAccountPasswordsOfAccountID = "of_account_id"
	colNameAccountPasswordsHash        = "hashed_password"
)

type AccountPassword struct {
	OfAccountId    uint64 `sql:"of_account_id"`
	HashedPassword string `sql:"hashed_password"`
}

type AccountPasswordDataAccessor interface {
	CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	GetAccountPassword(ctx context.Context, ofAccountId uint64) (AccountPassword, error)
	UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	WithDatabase(database IDatabase) AccountPasswordDataAccessor
}

type accountPasswordDataAccessor struct {
	database IDatabase
	logger   *zap.Logger
}

func NewAccountPasswordDataAccessor(database *goqu.Database, logger *zap.Logger) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
		logger:   logger,
	}
}

// CreateAccountPassword implements AccountPasswordDataAccessor.
func (a *accountPasswordDataAccessor) CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	logger := utils.LoggerWithContext(ctx, a.logger)

	_, err := a.database.
		Insert(tableNameAccountPasswords).
		Rows(goqu.Record{
			colNameAccountPasswordsHash: accountPassword.HashedPassword,
			colNameAccountPasswordsOfAccountID: accountPassword.OfAccountId,
		}).Executor().ExecContext(ctx)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create account password")
		return status.Errorf(codes.Internal, "failed to create account password: %+v", err)
	}

	return nil
}

// UpdateAccountPassword implements AccountPasswordDataAccessor.
func (a *accountPasswordDataAccessor) UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	logger := utils.LoggerWithContext(ctx, a.logger)

	_, err := a.database.
		Update(tableNameAccountPasswords).
		Set(goqu.Record{colNameAccountPasswordsHash: accountPassword.HashedPassword}).
		Where(goqu.Ex{colNameAccountPasswordsOfAccountID: accountPassword.OfAccountId}).
		Executor().
		ExecContext(ctx)
	
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update account password")
		return status.Errorf(codes.Internal, "failed to update account password: %+v", err)
	}

	return nil
}

func (a *accountPasswordDataAccessor) GetAccountPassword(ctx context.Context, ofAccountId uint64) (AccountPassword, error) {
	logger := utils.LoggerWithContext(ctx, a.logger)

	accountPassword := AccountPassword{}
	found, err := a.database.
		From(tableNameAccountPasswords).
		Where(goqu.Ex{colNameAccountPasswordsOfAccountID: ofAccountId}).
		ScanStructContext(ctx, &accountPassword)
	
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account password by id")
		return AccountPassword{}, status.Errorf(codes.Internal, "failed to get account password by id: %+v", err)
	}

	if !found {
		logger.Warn("cannot find account by id")
		return AccountPassword{}, sql.ErrNoRows
	}

	return accountPassword, nil
}

// WithDatabase implements AccountPasswordDataAccessor.
func (a *accountPasswordDataAccessor) WithDatabase(database IDatabase) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
	}
}
