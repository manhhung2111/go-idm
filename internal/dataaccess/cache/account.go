package cache

import (
	"context"

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
	if err := a.client.AddToSet(ctx, accountKeyName, accountName); err != nil {
		return err
	}

	return nil
}

// Has implements AccountNameCache.
func (a *accountNameCache) Has(ctx context.Context, accountName string) (bool, error) {
	return a.client.IsDataInSet(ctx, accountKeyName, accountName)
}
