package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/arcatva/spdktgt_svr/internal/config"
	"github.com/arcatva/spdktgt_svr/internal/grpc"
	"github.com/arcatva/spdktgt_svr/internal/logger"
	"github.com/arcatva/spdktgt_svr/internal/target"
	"github.com/sirupsen/logrus"
)

func main() {
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
		target.New(config.Load()).Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		grpc.StartGrpcServer(ctx)
	}()

	sig := <-sigMain
	logrus.Infof("Signal %v received. Shutting down...", sig)
	cancel()

	logrus.Infof("Waiting for shutting down...")
	wg.Wait()
	logrus.Infof("Terminated")
}
