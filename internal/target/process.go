package target

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

func (s *target) startProcess() (*exec.Cmd, error) {
	logrus.Infof("Starting nvmf_tgt with args: %v", s.args)
	cmd := exec.Command("/bin/nvmf_tgt", s.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM, // Send SIGTERM to the process when the parent dies
	}
	logrus.Infof("nvmf_tgt pid: %d", cmd.Process.Pid)
	if err := cmd.Start(); err != nil {
		logrus.Errorf("Failed to start nvmf_tgt: %v", err)
		return nil, err
	}
	logrus.Infof("nvmf_tgt started with pid: %d", cmd.Process.Pid)
	return cmd, nil
}
