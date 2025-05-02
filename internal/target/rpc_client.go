package target

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spdk/spdk/go/rpc/client"
)

func (s *target) waitForRpcReady() error {
	retryInterval := 100 * time.Millisecond
	timeout := 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("nvmf-tgt rpc ready timeout")
		default:
			// try connect rpc
			rpcClient, err := client.CreateClientWithJsonCodec(client.Unix, "/var/tmp/spdk.sock")
			if err != nil {
				time.Sleep(retryInterval)
				continue
			}

			// send "spdk_get_version"
			_, err = rpcClient.Call("spdk_get_version", nil)
			if err != nil {
				time.Sleep(retryInterval)
			}
			logrus.Info("nvmf-tgt rpc ready")
			s.rpcClient = rpcClient
			return nil
		}
	}

}
