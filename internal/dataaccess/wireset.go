package dataaccess

import (
	"github.com/google/wire"
	"github.com/manhhung2111/go-idm/internal/dataaccess/cache"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
	"github.com/manhhung2111/go-idm/internal/dataaccess/file"
	"github.com/manhhung2111/go-idm/internal/dataaccess/kafka"
)

var WireSet = wire.NewSet(
	database.WireSet,
	cache.WireSet,
	kafka.WireSet,
	file.WireSet,
)