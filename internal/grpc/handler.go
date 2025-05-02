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

	resp, err := target.GetTargetInstance().CallTargetRpcGet(target.SpdkGetVersion, nil)

	version := resp.Result.(map[string]interface{})["version"].(string)

	return &protos.SpdkVersion{
		Version: version,
	}, err
}

func (s *spdkServer) FrameworkGetReactors(context.Context, *emptypb.Empty) (*protos.FrameworkReactors, error) {
	resp, err := target.GetTargetInstance().CallTargetRpcGet(target.FramworkGetReactors, nil)
	if err != nil {
		return nil, err
	}
	m := resp.Result.(map[string]interface{})
	tickRate := uint64(m["tick_rate"].(float64))
	pid := uint32(m["pid"].(float64))

	rawRs := m["reactors"].([]interface{})
	reactors := make([]*protos.FrameworkReactor, len(rawRs))

	for i, r := range rawRs {
		rm := r.(map[string]interface{})
		reactor := &protos.FrameworkReactor{
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
		lwts := make([]*protos.FrameworkReactorLwThread, len(rawLT))
		for j, lt := range rawLT {
			ltm := lt.(map[string]interface{})
			lwts[j] = &protos.FrameworkReactorLwThread{
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
	resp, err := target.GetTargetInstance().CallTargetRpcGet(target.NvmfGetSubsystems, nil)
	if err != nil {
		return nil, err
	}

	rawSubsystems, ok := resp.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid subsystems data format")
	}

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
		listenAddresses := make([]*protos.NvmfListenAddress, len(rawListenAddresses))
		for j, addr := range rawListenAddresses {
			mAddr, ok := addr.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid listen address format")
			}
			// Using safe assertion for fields in listen address.
			trtype, ok := mAddr["trtype"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'trtype' field")
			}
			adrfam, ok := mAddr["adrfam"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'adrfam' field")
			}
			traddr, ok := mAddr["traddr"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'traddr' field")
			}
			trsvcid, ok := mAddr["trsvcid"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'trsvcid' field")
			}
			listenAddresses[j] = &protos.NvmfListenAddress{
				Trtype:  trtype,
				Adrfam:  adrfam,
				Traddr:  traddr,
				Trsvcid: trsvcid,
			}
		}

		// Common required fields.
		nqn, ok := mItem["nqn"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid or missing 'nqn' field")
		}
		subtype, ok := mItem["subtype"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid or missing 'subtype' field")
		}

		// For non-discovery subsystems, these fields must be present.
		var serialNumber, modelNumber string
		var maxNamespaces float64
		var minCntlid float64
		var maxCntlid float64

		if subtype != "Discovery" {
			sn, ok := mItem["serial_number"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'serial_number' field")
			}
			mn, ok := mItem["model_number"].(string)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'model_number' field")
			}
			mxNS, ok := mItem["max_namespaces"].(float64)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'max_namespaces' field")
			}
			minCt, ok := mItem["min_cntlid"].(float64)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'min_cntlid' field")
			}
			maxCt, ok := mItem["max_cntlid"].(float64)
			if !ok {
				return nil, fmt.Errorf("invalid or missing 'max_cntlid' field")
			}
			serialNumber = sn
			modelNumber = mn
			maxNamespaces = mxNS
			minCntlid = minCt
			maxCntlid = maxCt
		}

		// Process Namespaces field; may not be present
		var namespaces []*protos.NvmfNamespace
		if rawNS, exists := mItem["namespaces"]; exists {
			rawNamespaces, ok := rawNS.([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid namespaces data")
			}
			namespaces = make([]*protos.NvmfNamespace, len(rawNamespaces))
			for j, ns := range rawNamespaces {
				mNS, ok := ns.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("invalid namespace format")
				}
				nsid, ok := mNS["nsid"].(float64)
				if !ok {
					return nil, fmt.Errorf("invalid or missing 'nsid' field")
				}
				bdevName, ok := mNS["bdev_name"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid or missing 'bdev_name' field")
				}
				name, ok := mNS["name"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid or missing 'name' field")
				}
				nguid, ok := mNS["nguid"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid or missing 'nguid' field")
				}
				uuid, ok := mNS["uuid"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid or missing 'uuid' field")
				}
				namespaces[j] = &protos.NvmfNamespace{
					Nsid:     uint32(nsid),
					BdevName: bdevName,
					Name:     name,
					Nguid:    nguid,
					Uuid:     uuid,
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
				host, ok := h.(string)
				if !ok {
					return nil, fmt.Errorf("invalid host format")
				}
				hosts[j] = host
			}
		}

		subsystems[i] = &protos.NvmfSubsystem{
			Nqn:             nqn,
			Subtype:         subtype,
			ListenAddresses: listenAddresses,
			AllowAnyHost:    mItem["allow_any_host"].(bool), // Adjust check as needed.
			Hosts:           hosts,
			SerialNumber:    serialNumber,
			ModelNumber:     modelNumber,
			MaxNamespaces:   uint32(maxNamespaces),
			MinCntlid:       uint32(minCntlid),
			MaxCntlid:       uint32(maxCntlid),
			Namespaces:      namespaces,
		}
	}

	return &protos.NvmfSubsystems{
		Subsystems: subsystems,
	}, nil
}
