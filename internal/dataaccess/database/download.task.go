package database

import "github.com/doug-martin/goqu/v9"

type DownloadTaskDataAccessor interface{}

type downloadTaskDataAccessor struct {
	database IDatabase
}

func NewDownloadTaskDataAccessor(database *goqu.Database) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
	}
}
