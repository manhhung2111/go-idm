package handler

import (
	"github.com/google/wire"
	handler_consumer "github.com/manhhung2111/go-idm/internal/handler/consumer"
	"github.com/manhhung2111/go-idm/internal/handler/grpc"
	"github.com/manhhung2111/go-idm/internal/handler/http"
)

var WireSet = wire.NewSet(
	grpc.WireSet,
	http.WireSet,
	handler_consumer.WireSet,
)
