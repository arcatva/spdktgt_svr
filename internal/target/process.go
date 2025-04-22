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
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
