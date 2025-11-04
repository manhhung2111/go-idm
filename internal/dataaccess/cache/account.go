package cache

import (
	"context"

	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

const (
	accountKeyName = "go.idm:account.name.set"
)

type AccountNameCache interface {
	Add(ctx context.Context, accountName string) error
	Has(ctx context.Context, accountName string) (bool, error)
}

type accountNameCache struct {
	client CacheClient
	logger *zap.Logger
}

func NewAccountNameCache(
	client CacheClient,
	logger *zap.Logger,
) AccountNameCache {
	return &accountNameCache{
		client: client,
		logger: logger,
	}
}

// Add implements AccountNameCache.
func (a *accountNameCache) Add(ctx context.Context, accountName string) error {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.String("account_name", accountName))

	if err := a.client.AddToSet(ctx, accountKeyName, accountName); err != nil {
		logger.With(zap.Error(err)).Error("failed to add account name to set in cache")
		return err
	}

	return nil
}

// Has implements AccountNameCache.
func (a *accountNameCache) Has(ctx context.Context, accountName string) (bool, error) {
	logger := utils.LoggerWithContext(ctx, a.logger).With(zap.String("account_name", accountName))
	result, err := a.client.IsDataInSet(ctx, accountKeyName, accountName)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check if account name is in set in cache")
		return false, err
	}

	return result, nil
}
