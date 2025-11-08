package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/manhhung2111/go-idm/internal/dataaccess/cache"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	accountNameCache            cache.AccountNameCache
	logger                      *zap.Logger
}

func NewAccount(
	goquDatabase *goqu.Database,
	accountDataAccessor database.AccountDataAccessor,
	accountPasswordDataAccessor database.AccountPasswordDataAccessor,
	hashLogic Hash,
	tokenLogic Token,
	accountNameCache cache.AccountNameCache,
	logger *zap.Logger,
) Account {
	return &account{
		goquDatabase:                goquDatabase,
		accountDataAccessor:         accountDataAccessor,
		accountPasswordDataAccessor: accountPasswordDataAccessor,
		hashLogic:                   hashLogic,
		tokenLogic:                  tokenLogic,
		accountNameCache:            accountNameCache,
		logger:                      logger,
	}
}

// CreateAccount implements Account.
func (a *account) CreateAccount(ctx context.Context, params CreateAccountParams) (CreateAccountOutput, error) {
	logger := utils.LoggerWithContext(ctx, a.logger)

	accountNameTaken, err := a.isAccountNameTaken(ctx, params.AccountName)
	if err != nil {
		return CreateAccountOutput{}, status.Errorf(codes.Internal, "failed to check if account name is taken")
	}

	if accountNameTaken {
		return CreateAccountOutput{}, status.Error(codes.AlreadyExists, "account name is already taken")
	}
	
	var accountId uint64
	txError := a.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		accountNameTaken, err := a.isAccountNameTaken(ctx, params.AccountName)
		if err != nil {
			return nil
		}

		if accountNameTaken {
			return errors.New("account_name is already taken")
		}

		accountId, err = a.accountDataAccessor.WithDatabase(td, logger).CreateAccount(ctx, database.Account{
			AccountName: params.AccountName,
		})

		if err != nil {
			return err
		}

		hashedPassword, hashErr := a.hashLogic.Hash(ctx, params.Password)
		if hashErr != nil {
			return hashErr
		}

		if err := a.accountPasswordDataAccessor.WithDatabase(td, logger).CreateAccountPassword(ctx, database.AccountPassword{
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

		
	if err := a.accountNameCache.Add(ctx, params.AccountName); err != nil {
		logger.With(zap.Error(err)).Warn("failed to set account name into taken set in cache")
	}

	return CreateAccountOutput{
		ID:          accountId,
		AccountName: params.AccountName,
	}, nil
}

// CreateSession implements Account.
func (a *account) CreateSession(ctx context.Context, params CreateSessionParams) (string, error) {
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
		return "", status.Error(codes.Unauthenticated, "incorrect password")
	}

	token, _, err := a.tokenLogic.GetToken(ctx, existingAccount.ID)
	if err != nil {
		return "", err
	}
	
	return token, nil
}

func (a *account) isAccountNameTaken(ctx context.Context, accountName string) (bool, error) {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.String("account_name", accountName))
	
	accountNameTaken, err := a.accountNameCache.Has(ctx, accountName)
	if err != nil {
		logger.With(zap.Error(err)).Warn("failed to get account name from taken set in cache, will fall back to database")
	} else {
		return accountNameTaken, nil
	}
	
	_, err = a.accountDataAccessor.GetAccountByAccountName(ctx, accountName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
