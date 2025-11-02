package database

import (
	"context"
	"log"

	"github.com/doug-martin/goqu/v9"
)

type Account struct {
	AccountId   uint64 `sql:"account_id"`
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
}

func NewAccountDataAccessor(database *goqu.Database) AccountDataAccessor {
	return &accountDataAccessor{
		database: database,
	}
}

// CreateAccount implements AccountDataAccessor.
func (a *accountDataAccessor) CreateAccount(ctx context.Context, account Account) (uint64, error) {
	result, err := a.database.Insert("accounts").Rows(goqu.Record{
		"account_name": account.AccountName,
	}).Executor().ExecContext(ctx)

	if err != nil {
		log.Printf("failed to create account, err=%+v", err)
		return 0, err
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last inserted id, err=%+v", err)
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
	panic("unimplemented")
}

// GetAccountByAccountName implements AccountDataAccessor.
func (a *accountDataAccessor) GetAccountByAccountName(ctx context.Context, accountName string) (Account, error) {
	panic("unimplemented")
}
