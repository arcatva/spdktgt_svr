package target

import (
	"github.com/spdk/spdk/go/rpc/client"
)

type Api string

const (
	GetSpdkVersion Api = "spdk_get_version"
	
)

func (t *target) CallTargetRpc(api Api, param any) (*client.Response, error) {
	return t.RpcClient.Call(string(api), param)
}
