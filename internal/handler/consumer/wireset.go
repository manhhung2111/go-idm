package handler_consumer

import "github.com/google/wire"

var WireSet = wire.NewSet(
	NewRoot,
	NewDownloadTaskCreatedHandler,
)