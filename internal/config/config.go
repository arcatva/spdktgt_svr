package config

type Config struct {
	SpdkBin        string `json:"spdk_bin"`
	RpcSocket      string `json:"rpc_socket"`
	ConfigFile     string `json:"config_file"`
}

func Load() *Config {
	return &Config{ // default config
		SpdkBin:    "/bin/nvmf_tgt",
		RpcSocket:  "/var/tmp/spdk.sock",
		ConfigFile: "/etc/spdk/nvmf.json",
	}
}
