package target

import (
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

func (s *target) startProcess() (*exec.Cmd, error) {
	logrus.Infof("Starting nvmf_tgt with args: %v", s.args)
	cmd := exec.Command("/usr/bin/nvmf_tgt", s.args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM, // Send SIGTERM to the process when the parent dies
	}

	if err := cmd.Start(); err != nil {
		logrus.Errorf("Failed to start nvmf_tgt: %v", err)
		return nil, err
	}
	logrus.Infof("nvmf_tgt started with pid: %d", cmd.Process.Pid)

	go func() {
		t.done <- t.cmd.Wait()
		logrus.Info("nvmf_tgt process exited")
	}()

	return cmd, nil
}
