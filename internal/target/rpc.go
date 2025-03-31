package target

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spdk/spdk/go/rpc/client"
)

func (s *Target) waitForRpcReady() error {
	maxRetries := 10
	retryInterval := 1 * time.Second
	timeout := 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("rpc ready timeout")
		default:
			// try connect rpc
			rpcClient, err := client.CreateClientWithJsonCodec(client.Unix, s.config.RpcSocket)
			if err != nil {
				log.Printf("rpc connection failed (retry: %d/%d): %v", i+1, maxRetries, err)
				time.Sleep(retryInterval)
				continue
			}
			defer rpcClient.Close()

			// send "spdk_get_version"
			_, err = rpcClient.Call("spdk_get_version", nil)
			if err == nil {
				log.Println("rpc ready")
				return nil
			}

			log.Printf("rpc connection failed (retry: %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("RPC not ready after %v", timeout)
}

func (s *Target) configureTarget() error {
	//create client
	rpcClient, err := client.CreateClientWithJsonCodec(client.Unix, s.config.RpcSocket)
	if err != nil {
		return fmt.Errorf("configureNvmfTgt: %s", err.Error())
	}
	defer rpcClient.Close()
	//sends a JSON-RPC 2.0 request with "bdev_get_bdevs" method and provided params
	resp, err := rpcClient.Call("nvmf_create_transport", getTcpParams())
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
