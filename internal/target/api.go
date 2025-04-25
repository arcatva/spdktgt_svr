package target

import (
	"github.com/spdk/spdk/go/rpc/client"
)

type Api string

const (
	SpdkGetVersion      Api = "spdk_get_version"
	FramworkGetReactors     = "framework_get_reactors"
	NvmfGetSubsystems       = "nvmf_get_subsystems"
)

func (t *target) CallTargetRpc(api Api, param any) (*client.Response, error) {
	return t.RpcClient.Call(string(api), param)
}
