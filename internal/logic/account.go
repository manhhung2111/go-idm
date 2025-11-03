package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
)

type CreateAccountParams struct {
	AccountName string
	Password    string
}

type CreateSessionParams struct {
	AccountName string
	Password    string
}

type CreateAccountOutput struct {
	ID          uint64
	AccountName string
}

type Account interface {
	CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error)
	CreateSession(ctx context.Context, params CreateSessionParams) (token string, err error)
}

type account struct {
	goquDatabase                *goqu.Database
	accountDataAccessor         database.AccountDataAccessor
	accountPasswordDataAccessor database.AccountPasswordDataAccessor
	hashLogic                   Hash
	tokenLogic                  Token
}

func NewAccount(
	goquDatabase *goqu.Database,
	accountDataAccessor database.AccountDataAccessor,
	accountPasswordDataAccessor database.AccountPasswordDataAccessor,
	hashLogic Hash,
	tokenLogic Token,
) Account {
	return &account{
		goquDatabase:                goquDatabase,
		accountDataAccessor:         accountDataAccessor,
		accountPasswordDataAccessor: accountPasswordDataAccessor,
		hashLogic:                   hashLogic,
		tokenLogic:                  tokenLogic,
	}
}

// CreateAccount implements Account.
func (a *account) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
	var accountId uint64

	txError := a.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		accountNameTaken, err := a.isAccountNameTaken(ctx, params.AccountName)
		if err != nil {
			return nil
		}

		if accountNameTaken {
			return errors.New("account_name is already taken")
		}

		accountId, err = a.accountDataAccessor.WithDatabase(td).CreateAccount(ctx, database.Account{
			AccountName: params.AccountName,
		})

		if err != nil {
			return err
		}

		hashedPassword, err := a.hashLogic.Hash(ctx, params.Password)
		if err != nil {
			return err
		}

		if err := a.accountPasswordDataAccessor.WithDatabase(td).CreateAccountPassword(ctx, database.AccountPassword{
			OfAccountId:    accountId,
			HashedPassword: hashedPassword,
		}); err != nil {
			return nil
		}

		return nil
	})

	if txError != nil {
		return CreateAccountOutput{}, txError
	}

	return CreateAccountOutput{
		ID:          accountId,
		AccountName: params.AccountName,
	}, nil
}

// CreateSession implements Account.
func (a *account) CreateSession(ctx context.Context, params CreateSessionParams) (token string, err error) {
	existingAccount, err := a.accountDataAccessor.GetAccountByAccountName(ctx, params.AccountName)
	if err != nil {
		return "", err
	}

	existingAccountPassword, err := a.accountPasswordDataAccessor.GetAccountPassword(ctx, existingAccount.ID)
	if err != nil {
		return "", err
	}

	isHashEqual, err := a.hashLogic.IsHashEqual(ctx, params.Password, existingAccountPassword.HashedPassword)
	if err != nil {
		return "", err
	}

	if !isHashEqual {
		return "", errors.New("incorrect password")
	}

	return "", nil
}

func (a *account) isAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	if _, err := a.accountDataAccessor.GetAccountByAccountName(ctx, accountName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
