//go:build wireinject
// +build wireinject

//
// go:generate go run github.com/google/wire/cmd/wire

package wiring

import (
	"github.com/google/wire"
	"github.com/manhhung2111/go-idm/internal/config"
	"github.com/manhhung2111/go-idm/internal/dataaccess"
	"github.com/manhhung2111/go-idm/internal/handler"
	"github.com/manhhung2111/go-idm/internal/logic"
	"github.com/manhhung2111/go-idm/internal/handler/grpc"
)

var WireSet = wire.NewSet(
	config.WireSet,
	dataaccess.WireSet,
	handler.WireSet,
	logic.WireSet,
)

func InitializeGrpcServer(configFilePath config.ConfigFilePath) (grpc.Server, func(), error) {
	wire.Build(WireSet)

	return nil, nil, nil
}