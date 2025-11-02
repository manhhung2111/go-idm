package handler

import (
	"github.com/google/wire"
	"github.com/manhhung2111/go-idm/internal/handler/http"
	"github.com/manhhung2111/go-idm/internal/handler/grpc"
)

var WireSet = wire.NewSet(
	grpc.WireSet,
	http.WireSet,
)
