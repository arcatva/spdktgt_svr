package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/arcatva/spdktgt_svr/internal/target"
	"github.com/arcatva/spdktgt_svr/pkg/api/protos"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// spdkServer implements the pb.SpdkServiceServer interface.
type spdkServer struct {
	protos.UnimplementedSpdkServer
}

func (s *spdkServer) GetSpdkVersion(context.Context, *emptypb.Empty) (*protos.SpdkVersion, error) {

	resp, err := target.Get().CallTargetRpc(target.GetSpdkVersion, nil)

	version := resp.Result.(map[string]interface{})["version"].(string)

	return &protos.SpdkVersion{
		Version: version,
	}, err
}

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
