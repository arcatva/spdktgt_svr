package target

import (
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

func (s *target) startProcess() (*exec.Cmd, error) {
	logrus.Infof("Starting nvmf-tgt with args: %v", s.args)
	cmd := exec.Command("/usr/bin/nvmf-tgt", s.args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM, // Send SIGTERM to the process when the parent dies
	}

	if err := cmd.Start(); err != nil {
		logrus.Errorf("Failed to start nvmf-tgt: %v", err)
		return nil, err
	}
	logrus.Infof("nvmf-tgt started with pid: %d", cmd.Process.Pid)

	go func() {
		targetInstance.done <- targetInstance.cmd.Wait()
		logrus.Info("nvmf-tgt process exited")
	}()

	return cmd, nil
}
