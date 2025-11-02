package dataaccess

import (
	"github.com/google/wire"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
)

var WireSet = wire.NewSet(
	database.WireSet,
)