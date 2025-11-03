package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
	"database/sql"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

const (
	tableNameAccounts          = "accounts"
	colNameAccountsID          = "id"
	colNameAccountsAccountName = "account_name"
)

type Account struct {
	ID          uint64 `sql:"id"`
	AccountName string `sql:"account_name"`
}

type AccountDataAccessor interface {
	CreateAccount(ctx context.Context, account Account) (uint64, error)
	GetAccountById(ctx context.Context, id uint64) (Account, error)
	GetAccountByAccountName(ctx context.Context, accountName string) (Account, error)
	WithDatabase(database IDatabase) AccountDataAccessor
}

type accountDataAccessor struct {
	database IDatabase
	logger   *zap.Logger
}

func NewAccountDataAccessor(database *goqu.Database, logger *zap.Logger) AccountDataAccessor {
	return &accountDataAccessor{
		database: database,
		logger:   logger,
	}
}

// CreateAccount implements AccountDataAccessor.
func (a *accountDataAccessor) CreateAccount(ctx context.Context, account Account) (uint64, error) {
	logger := utils.LoggerWithContext(ctx, a.logger)

	result, err := a.database.Insert(tableNameAccounts).Rows(goqu.Record{
		colNameAccountsAccountName: account.AccountName,
	}).Executor().ExecContext(ctx)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create account")
		return 0, err
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get last inserted id")
		return 0, err
	}
	return uint64(lastInsertedId), nil
}

// WithDatabase implements AccountDataAccessor.
func (a *accountDataAccessor) WithDatabase(database IDatabase) AccountDataAccessor {
	return &accountDataAccessor{
		database: database,
	}
}

// GetAccountById implements AccountDataAccessor.
func (a *accountDataAccessor) GetAccountById(ctx context.Context, id uint64) (Account, error) {
	logger := utils.LoggerWithContext(ctx, a.logger)

	account := Account{}
	found, err := a.database.
		From(tableNameAccounts).
		Where(goqu.Ex{colNameAccountsID: id}).
		ScanStructContext(ctx, &account)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account by id")
		return Account{}, err
	}

	if !found {
		logger.Warn("cannot find account by id")
		return Account{}, sql.ErrNoRows
	}

	return account, nil
}

// GetAccountByAccountName implements AccountDataAccessor.
func (a *accountDataAccessor) GetAccountByAccountName(ctx context.Context, accountName string) (Account, error) {
	logger := utils.LoggerWithContext(ctx, a.logger)

	account := Account{}
	found, err := a.database.
		From(tableNameAccounts).
		Where(goqu.Ex{colNameAccountsAccountName: accountName}).
		ScanStructContext(ctx, &account)

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get account by account_name")
		return Account{}, err
	}

	if !found {
		logger.Warn("cannot find account by account_name")
		return Account{}, sql.ErrNoRows
	}

	return account, nil
}
