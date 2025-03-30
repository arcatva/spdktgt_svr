package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/spdk/spdk/go/rpc/client"
)

const (
	successKeyword = "Reactor started on core 0"
	timeout        = 3 * time.Second
)

type Config struct {
	SpdkBin        string `json:"spdk_bin"`
	RpcSocket      string `json:"rpc_socket"`
	ConfigFile     string `json:"config_file"`
	StatusEndpoint string `json:"status_endpoint"`
}

func loadConfig() Config {

	return Config{ // default config
		SpdkBin:    "/bin/nvmf_tgt",
		RpcSocket:  "/var/tmp/spdk.sock",
		ConfigFile: "/etc/spdk/nvmf.json",
		//		StatusEndpoint: "http://localhost:8080/status/update",
	}

}

func main() {

	// handle system signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// load configurations
	config := loadConfig()

	log.SetOutput(os.Stdout)
	log.Println("spdk target server starting...")

	// start nvmf_tgt
	cmd, err := startNvmfTgt(&config)
	if err != nil {
		log.Fatalln("failed to start nvmf_tgt: %v", err)
	}
	log.Println("nvmf_tgt started at: %v", cmd.ProcessState.Pid())
	log.Println("spdk target server started")

	// watch if nvmf_tgt process exits
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// block main thread for nvmf_tgt
	select {
	case <-sig:
		log.Println("signal received, exiting nvmf_tgt")
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			log.Println("failed to send signal to nvmf_tgt: %v", err)
		}
		// timeout for killing nvmf_tgt process
		forceKillTimer := time.AfterFunc(5*time.Second, func() {
			log.Println("nvmf_tgt not responding, force exiting")
			if err := cmd.Process.Kill(); err != nil {
				log.Println("Failed to force exit nvmf_tgt: %v", err)
			}
		})
		defer forceKillTimer.Stop()
	case err := <-done: // nvmf_tgt exit by itself
		if err != nil {
			log.Println("nvmf_tgt panic: %v", err)
			os.Exit(1)
		} else {
			log.Println("nvmf_tgt gracefully exited")
		}
	}

}

func startNvmfTgt(config *Config) (*exec.Cmd, error) {

	log.Printf("Starting nvmf_tgt with binary: %s", config.SpdkBin)

	// create command，use -r for nvmf_tgt and define RpcSocket location
	cmd := exec.Command(config.SpdkBin, "-r", config.RpcSocket)
	cmd.Stdout = os.Stdout // stdout redirected to parent's stdout
	cmd.Stderr = os.Stderr

	// start nvmf_tgt with config
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	log.Println("nvmf_tgt started")

	// wait for rpc server ready
	if err := waitForRpcReady(config); err != nil {
		return nil, err
	}

	// config nvmf_tgt
	if err := configureNvmfTgt(config); err != nil {
		return nil, err
	}

	return cmd, nil
}

func waitForRpcReady(config *Config) error {
	// retry timeout
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
			rpcClient, err := client.CreateClientWithJsonCodec(client.Unix, config.RpcSocket)
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
	return fmt.Errorf("rpc not ready in %v seconds", timeout)
}

func configureNvmfTgt(config *Config) {
	//create client
	rpcClient, err := client.CreateClientWithJsonCodec(client.Unix, config.RpcSocket)
	if err != nil {
		log.Fatalf("error on client creation, err: %s", err.Error())
	}
	defer rpcClient.Close()
	//sends a JSON-RPC 2.0 request with "bdev_get_bdevs" method and provided params
	resp, err := rpcClient.Call("nvmf_create_transport", getTcpParams())
	if err != nil {
		log.Fatalf("error on JSON-RPC call, method: %s err: %s", "bdev_get_bdevs", err.Error())
	}
	result, err := json.Marshal(resp.Result)
	if err != nil {
		log.Print(fmt.Errorf("error when creating json string representation: %w", err).Error())
	}
	log.Printf("%s\n", string(result))
	log.Println("nvmf_tgt configured")
}

func getTcpParams() map[string]any {
	params := map[string]interface{}{
		"trtype":               "TCP",
		"io_unit_size":         16384,
		"max_io_size":          131072,
		"max_qpairs_per_ctrl":  8,
		"in_capsule_data_size": 8192,
	}
	return params
}
