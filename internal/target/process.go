package target

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func (s *target) startProcess() (*exec.Cmd, error) {
	logrus.Printf("starting nvmf_tgt with binary: %s\n", s.config.SpdkBin)

	cmd := exec.Command(s.config.SpdkBin, "-r", s.config.RpcSocket)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
