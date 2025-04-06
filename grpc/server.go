package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/arcatva/spdktgt_svr/grpc/protos"
	"github.com/arcatva/spdktgt_svr/internal/target"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// spdkServer implements the pb.SpdkServiceServer interface.
type spdkServer struct {
	protos.UnimplementedSpdkServer
}

func (s *spdkServer) GetSpdkVersion(context.Context, *emptypb.Empty) (*protos.SpdkVersion, error) {

	resp, err := target.Target.CallTargetRpc(target.SpdkGetVersion, nil)

	return &protos.SpdkVersion{
		Version: string(resp.Result.(string)),
	}, err
}

func StartGrpcServer() {

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 50051))
	if err != nil {
		logrus.Fatalf("failed create gRPC server to listen at: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	protos.RegisterSpdkServer(grpcServer, &spdkServer{})
	grpcServer.Serve(lis)
}
