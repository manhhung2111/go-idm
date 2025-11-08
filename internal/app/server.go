package app

import (
	"context"
	"syscall"

	handler_consumer "github.com/manhhung2111/go-idm/internal/handler/consumer"
	"github.com/manhhung2111/go-idm/internal/handler/grpc"
	"github.com/manhhung2111/go-idm/internal/handler/http"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

type Server struct {
	grpcServer grpc.Server
	httpServer http.Server
	rootConsumer handler_consumer.Root
	logger *zap.Logger
}

func NewServer(
	grpcServer grpc.Server,
	httpServer http.Server,
	rootConsumer handler_consumer.Root,
	logger *zap.Logger,
) *Server {
	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
		rootConsumer: rootConsumer,
		logger: logger,
	}
}

func (s *Server) Start() error {
	go func() {
		err := s.grpcServer.Start(context.Background())
		s.logger.With(zap.Error(err)).Info("grpc server stopped")
	}()

	go func() {
		err := s.httpServer.Start(context.Background())
		s.logger.With(zap.Error(err)).Info("http server stopped")
	}()

	go func() {
		err := s.rootConsumer.Start(context.Background())
		s.logger.With(zap.Error(err)).Info("message queue consumer stopped")
	}()

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)
	return nil
}