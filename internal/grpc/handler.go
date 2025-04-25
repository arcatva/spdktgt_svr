package grpc

import (
	"context"
	"fmt"

	"github.com/arcatva/spdktgt_svr/internal/target"
	"github.com/arcatva/spdktgt_svr/pkg/api/protos"
	"google.golang.org/protobuf/types/known/emptypb"
)

// spdkServer implements the pb.SpdkServiceServer interface.
type spdkServer struct {
	protos.UnimplementedSpdkServer
}

func (s *spdkServer) SpdkGetVersion(context.Context, *emptypb.Empty) (*protos.SpdkVersion, error) {

	resp, err := target.Get().CallTargetRpc(target.SpdkGetVersion, nil)

	version := resp.Result.(map[string]interface{})["version"].(string)

	return &protos.SpdkVersion{
		Version: version,
	}, err
}

func (s *spdkServer) FrameworkGetReactors(context.Context, *emptypb.Empty) (*protos.FrameworkReactors, error) {
	resp, err := target.Get().CallTargetRpc(target.FramworkGetReactors, nil)
	if err != nil {
		return nil, err
	}
	m := resp.Result.(map[string]interface{})
	tickRate := uint64(m["tick_rate"].(float64))
	pid := uint32(m["pid"].(float64))

	rawRs := m["reactors"].([]interface{})
	reactors := make([]*protos.Reactor, len(rawRs))

	for i, r := range rawRs {
		rm := r.(map[string]interface{})
		reactor := &protos.Reactor{
			Lcore:       uint32(rm["lcore"].(float64)),
			Tid:         uint32(rm["tid"].(float64)),
			Busy:        uint64(rm["busy"].(float64)),
			Idle:        uint64(rm["idle"].(float64)),
			InInterrupt: rm["in_interrupt"].(bool),
			Irq:         uint32(rm["irq"].(float64)),
			Sys:         uint64(rm["sys"].(float64)),
			Usr:         uint64(rm["usr"].(float64)),
		}
		rawLT := rm["lw_threads"].([]interface{})
		lwts := make([]*protos.LwThread, len(rawLT))
		for j, lt := range rawLT {
			ltm := lt.(map[string]interface{})
			lwts[j] = &protos.LwThread{
				Name:    ltm["name"].(string),
				Id:      uint32(ltm["id"].(float64)),
				Cpumask: ltm["cpumask"].(string),
				Elapsed: uint64(ltm["elapsed"].(float64)),
			}
		}
		reactor.LwThreads = lwts
		reactors[i] = reactor
	}

	return &protos.FrameworkReactors{
		TickRate: tickRate,
		Pid:      pid,
		Reactors: reactors,
	}, nil
}

func (s *spdkServer) NvmfGetSubsystems(ctx context.Context, _ *emptypb.Empty) (*protos.NvmfSubsystems, error) {
	resp, err := target.Get().CallTargetRpc(target.NvmfGetSubsystems, nil)
	if err != nil {
		return nil, err
	}

	rawSubsystems := resp.Result.([]interface{})

	subsystems := make([]*protos.NvmfSubsystem, len(rawSubsystems))
	for i, item := range rawSubsystems {
		mItem, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid subsystem format")
		}

		// Process ListenAddresses field
		rawListenAddresses, ok := mItem["listen_addresses"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid listen_addresses data")
		}
		listenAddresses := make([]*protos.ListenAddress, len(rawListenAddresses))
		for j, addr := range rawListenAddresses {
			mAddr := addr.(map[string]interface{})
			listenAddresses[j] = &protos.ListenAddress{
				Trtype:  mAddr["trtype"].(string),
				Adrfam:  mAddr["adrfam"].(string),
				Traddr:  mAddr["traddr"].(string),
				Trsvcid: mAddr["trsvcid"].(string),
			}
		}

		// Process Namespaces field; note that namespaces may not be present for Discovery subsystem.
		var namespaces []*protos.Namespace
		if rawNS, exists := mItem["namespaces"]; exists {
			rawNamespaces, ok := rawNS.([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid namespaces data")
			}
			namespaces = make([]*protos.Namespace, len(rawNamespaces))
			for j, ns := range rawNamespaces {
				mNS := ns.(map[string]interface{})
				namespaces[j] = &protos.Namespace{
					Nsid:     uint32(mNS["nsid"].(float64)),
					BdevName: mNS["bdev_name"].(string),
					Name:     mNS["name"].(string),
					Nguid:    mNS["nguid"].(string),
					Uuid:     mNS["uuid"].(string),
				}
			}
		}

		// Process Hosts field.
		var hosts []string
		if rawHosts, exists := mItem["hosts"]; exists {
			rawHostsArray, ok := rawHosts.([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid hosts data")
			}
			hosts = make([]string, len(rawHostsArray))
			for j, h := range rawHostsArray {
				hosts[j] = h.(string)
			}
		}

		subsystems[i] = &protos.NvmfSubsystem{
			Nqn:             mItem["nqn"].(string),
			Subtype:         mItem["subtype"].(string),
			ListenAddresses: listenAddresses,
			AllowAnyHost:    mItem["allow_any_host"].(bool),
			Hosts:           hosts,
			SerialNumber:    mItem["serial_number"].(string),
			ModelNumber:     mItem["model_number"].(string),
			MaxNamespaces:   uint32(mItem["max_namespaces"].(float64)),
			MinCntlid:       uint32(mItem["min_cntlid"].(float64)),
			MaxCntlid:       uint32(mItem["max_cntlid"].(float64)),
			Namespaces:      namespaces,
		}
	}

	return &protos.NvmfSubsystems{
		Subsystems: subsystems,
	}, nil
}
