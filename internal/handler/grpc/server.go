package grpc

import (
	"context"
	"net"

	"github.com/manhhung2111/go-idm/internal/config"
	go_idm_v1 "github.com/manhhung2111/go-idm/internal/generated/proto"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	handler    go_idm_v1.GoIDMServiceServer
	grpcConfig config.GRPC
	logger     *zap.Logger
}

func NewServer(
	handler go_idm_v1.GoIDMServiceServer, 
	grpcConfig config.GRPC,
	logger *zap.Logger,
) Server {
	return &server{
		handler: handler,
		grpcConfig: grpcConfig,
		logger: logger,
	}
}

func (s *server) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to open tcp listener")
		return err
	}

	defer listener.Close()

	server := grpc.NewServer()
	go_idm_v1.RegisterGoIDMServiceServer(server, s.handler)

	logger.With(zap.String("address", s.grpcConfig.Address)).Info("starting grpc server")
	return server.Serve(listener)
}
