package target

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spdk/spdk/go/rpc/client"
)

type target struct {
	RpcClient *client.Client
	args      []string
	cmd       *exec.Cmd
	done      chan error
}

var t *target

func New(args []string) *target {
	t = &target{
		args: args,
		done: make(chan error, 1),
	}
	return t
}

func Get() *target {
	return t
}

func (t *target) Start(ctx context.Context) error {
	logrus.Info("nvmf_tgt process starting")

	cmd, err := t.startProcess()
	if err != nil {
		return err
	}
	t.cmd = cmd

	go func() {
		t.done <- t.cmd.Wait()
	}()

	if err := t.waitForRpcReady(); err != nil {
		return err
	}

	logrus.Info("nvmf_tgt configured successfully")

	<-ctx.Done()
	logrus.Info("received context cancel, shutting down nvmf_tgt")

	return t.Stop()
}

func (t *target) Stop() error {

	if t.RpcClient != nil {
		t.RpcClient.Close()
		logrus.Info("json_rpc client stopped")
	}

	if t.cmd == nil || t.cmd.Process == nil {
		return fmt.Errorf("no nvmf_tgt process found")
	}

	logrus.Info("sending SIGTERM")
	if err := t.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case err := <-t.done:
		if err != nil {
			logrus.Errorf("nvmf_tgt exited with error: %v", err)
		}
		logrus.Info("nvmf_tgt was successfully terminated")
		return nil
	case <-ctx.Done():
		logrus.Error("force killing nvmf_tgt process")
		return t.cmd.Process.Kill()
	}
}
