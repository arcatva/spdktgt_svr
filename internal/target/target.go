package target

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/arcatva/spdktgt_svr/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spdk/spdk/go/rpc/client"
)

type target struct {
	RpcClient *client.Client
	config    *config.Config
	cmd       *exec.Cmd
	done      chan error
}

var Target *target

func New(config *config.Config) error {

	Target = &target{
		config: config,
		done:   make(chan error, 1),
	}

	return nil
}

func (s *target) Start() error {
	logrus.Info("nvmf_tgt process starting")
	var err error
	// start nvmf_tgt process
	s.cmd, err = s.startProcess()
	if err != nil {
		return err
	}

	logrus.Info("nvmf_tgt process started")

	// start daemon co-routine
	go func() {
		s.done <- s.cmd.Wait()
	}()

	if err := s.waitForRpcReady(); err != nil {
		return err
	}

	if err := s.configureTarget(); err != nil {
		return err
	}

	logrus.Info("nvmf_tgt configured successfully")
	return nil
}

func (s *target) Stop() error {

	logrus.Info("stopping nvmf_tgt")

	s.RpcClient.Close()

	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("no nvmf_tgt process found")
	}

	logrus.Info("sending SIGTERM")

	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-s.done:
		logrus.Info("nvmf_tgt was successfully terminated")
		return nil
	case <-ctx.Done():
		logrus.Error("force killing nvmf_tgt process")
		return s.cmd.Process.Kill()
	}
}
