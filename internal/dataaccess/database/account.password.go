package database

import (
	"context"

	"github.com/doug-martin/goqu/v9"
)

type AccountPassword struct {
	OfAccountId    uint64 `sql:"of_account_id"`
	HashedPassword string `sql:"hashed_password"`
}

type AccountPasswordDataAccessor interface {
	CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error
	WithDatabase(database IDatabase) AccountPasswordDataAccessor
}

type accountPasswordDataAccessor struct {
	database IDatabase
}

func NewAccountPasswordDataAccessor(database *goqu.Database) AccountPasswordDataAccessor {
	return accountPasswordDataAccessor{
		database: database,
	}
}

// CreateAccountPassword implements AccountPasswordDataAccessor.
func (a accountPasswordDataAccessor) CreateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	panic("unimplemented")
}

// UpdateAccountPassword implements AccountPasswordDataAccessor.
func (a accountPasswordDataAccessor) UpdateAccountPassword(ctx context.Context, accountPassword AccountPassword) error {
	panic("unimplemented")
}

// WithDatabase implements AccountPasswordDataAccessor.
func (a accountPasswordDataAccessor) WithDatabase(database IDatabase) AccountPasswordDataAccessor {
	return &accountPasswordDataAccessor{
		database: database,
	}
}
