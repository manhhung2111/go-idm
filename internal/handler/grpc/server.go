package grpc

import (
	"context"
	"net"

	go_idm_v1 "github.com/manhhung2111/go-idm/internal/generated/proto"
	"google.golang.org/grpc"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	handler go_idm_v1.GoIDMServiceServer
}

func NewServer(handler go_idm_v1.GoIDMServiceServer) Server {
	return &server {
		handler: handler,
	}
}

func (s *server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		return err
	}

	defer listener.Close()

	server := grpc.NewServer()
	go_idm_v1.RegisterGoIDMServiceServer(server, s.handler)
	return server.Serve(listener)
}
