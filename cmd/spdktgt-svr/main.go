package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/arcatva/spdktgt_svr/internal/grpc"
	"github.com/arcatva/spdktgt_svr/internal/logger"
	"github.com/arcatva/spdktgt_svr/internal/target"
	"github.com/sirupsen/logrus"
)

func main() {
	args := os.Args
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	logger.Init()
	logrus.Info("SPDK Target Server starting...")

	sigMain := make(chan os.Signal, 1)
	signal.Notify(sigMain, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := target.CreateTargetInstance(args).Start(ctx, cancel); err != nil {
			logrus.Fatalf("Failed to start nvmf-tgt: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpc.StartGrpcServer(ctx); err != nil {
			logrus.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	sig := <-sigMain
	logrus.Infof("Signal %v received. Shutting down...", sig)
	cancel()

	logrus.Infof("Waiting for shutting down...")
	wg.Wait()
	logrus.Infof("Terminated")
}
