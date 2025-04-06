package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arcatva/spdktgt_svr/grpc"
	"github.com/arcatva/spdktgt_svr/internal/config"
	"github.com/arcatva/spdktgt_svr/internal/logger"
	"github.com/arcatva/spdktgt_svr/internal/target"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.Init()
	logrus.Info("SPDK Target Server starting...")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// load configuration
	cfg := config.Load()

	if err := target.New(&cfg); err != nil {
		logrus.Fatalf("failed to create target: %v", err)
	}

	if err := target.Target.Start(); err != nil {
		logrus.Fatalf("failed to start target: %v", err)
	}

	grpc.StartGrpcServer()

	<-sig
	log.Println("shutting down target...")
	if err := target.Target.Stop(); err != nil {
		logrus.Fatalf("error shutting down target: %v", err)
	}
	logrus.Info("SPDK Target Server stopped...")
}
