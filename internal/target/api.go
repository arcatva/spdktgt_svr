package target

import (
	"github.com/spdk/spdk/go/rpc/client"
)

type GetApi string
type SetApi string

const (
	SpdkGetVersion      GetApi = "spdk_get_version"
	FramworkGetReactors GetApi = "framework_get_reactors"
	NvmfGetSubsystems   GetApi = "nvmf_get_subsystems"
)

func (t *target) CallTargetRpcGet(getApi GetApi, param any) (*client.Response, error) {
	t.rwMutex.RLock()
	defer t.rwMutex.RUnlock()
	return t.rpcClient.Call(string(getApi), param)
}

func (t *target) CallTargetRpcSet(setApi SetApi, param any) (*client.Response, error) {
	t.rwMutex.Lock()
	defer t.rwMutex.Unlock()
	return t.rpcClient.Call(string(setApi), param)
}
