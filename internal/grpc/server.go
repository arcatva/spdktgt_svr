package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/arcatva/spdktgt_svr/pkg/api/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func StartGrpcServer(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	grpcServer := grpc.NewServer()
	protos.RegisterSpdkServer(grpcServer, &spdkServer{})

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- grpcServer.Serve(lis)
	}()

	logrus.Info("gRPC server started on port 50051")

	select {
	case <-ctx.Done():
		logrus.Info("received context cancel, shutting down gRPC server")
		grpcServer.GracefulStop()
		return nil
	case err := <-serveErr:
		logrus.Errorf("failed to serve: %v", err)
		grpcServer.Stop()
		return err
	}
}
