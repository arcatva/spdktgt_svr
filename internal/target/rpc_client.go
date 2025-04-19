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
			return fmt.Errorf("nvmf_tgt rpc ready timeout")
		default:
			// try connect rpc
			rpcClient, err := client.CreateClientWithJsonCodec(client.Unix, s.config.RpcSocket)
			if err != nil {
				time.Sleep(retryInterval)
				continue
			}

			// send "spdk_get_version"
			_, err = rpcClient.Call("spdk_get_version", nil)
			if err != nil {
				time.Sleep(retryInterval)
			}
			logrus.Info("nvmf_tgt rpc ready")
			s.RpcClient = rpcClient
			return nil
		}
	}

}

/* func (s *target) configureTarget() error {

	//sends a JSON-RPC 2.0 request with "bdev_get_bdevs" method and provided params
	resp, err := s.RpcClient.Call("nvmf_create_transport", getTcpParams())
	if err != nil {
		return fmt.Errorf("configureNvmfTgt: error on JSON-RPC call, method: %s err: %s", "nvmf_create_transport", err.Error())
	}
	result, err := json.Marshal(resp.Result)
	if err != nil {
		return fmt.Errorf("configureNvmfTgt: %w", err)
	}
	logrus.Println("nvmf_tgt configuration result: ", result)
	return nil
}

func getTcpParams() map[string]interface{} {
	return map[string]interface{}{
		"trtype":               "TCP",
		"io_unit_size":         16384,
		"max_io_size":          131072,
		"max_qpairs_per_ctrl":  8,
		"in_capsule_data_size": 8192,
	}
}
*/
