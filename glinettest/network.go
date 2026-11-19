package glinettest

import (
	"encoding/json"
	"maps"
)

// Network serves the check_wan_cable, get_arp_list, get_advance_config and set_advance_config
// methods of the network module. The results of check_wan_cable and get_arp_list are those of the
// test router until a test sets them. check_wan_cable ignores the params.
//
// The result of get_advance_config is {"nat_enable":0,"sip_enable":0} until a test or a write sets
// it. This is an assumption: the test router did not give this result, and the values are the
// defaults of the web UI. set_advance_config is also an assumption, because it was not tested on a
// router: it sets the fields of its params in the result and keeps the others. It does not check
// the values.
type Network struct {
	wanCable, arpList, advanceConfig readResult
}

func newNetwork() *Network {
	m := &Network{}
	m.wanCable.set(json.RawMessage(networkWANCable))
	m.arpList.set(json.RawMessage(networkARPList))
	m.advanceConfig.set(json.RawMessage(`{"nat_enable":0,"sip_enable":0}`))
	return m
}

// WANCable returns the result of check_wan_cable.
func (m *Network) WANCable() json.RawMessage {
	return m.wanCable.get()
}

// SetWANCable sets the result of check_wan_cable.
func (m *Network) SetWANCable(result json.RawMessage) {
	m.wanCable.set(result)
}

// ARPList returns the result of get_arp_list.
func (m *Network) ARPList() json.RawMessage {
	return m.arpList.get()
}

// SetARPList sets the result of get_arp_list.
func (m *Network) SetARPList(result json.RawMessage) {
	m.arpList.set(result)
}

// AdvanceConfig returns the result of get_advance_config.
func (m *Network) AdvanceConfig() json.RawMessage {
	return m.advanceConfig.get()
}

// SetAdvanceConfig sets the result of get_advance_config.
func (m *Network) SetAdvanceConfig(result json.RawMessage) {
	m.advanceConfig.set(result)
}

func (m *Network) register(s *Server) {
	s.Handle("network", "check_wan_cable", m.wanCable.handle)
	s.Handle("network", "get_arp_list", m.arpList.handle)
	s.Handle("network", "get_advance_config", m.advanceConfig.handle)
	s.Handle("network", "set_advance_config", m.setAdvanceConfig)
}

func (m *Network) setAdvanceConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	err = m.advanceConfig.update(func(v any) (any, error) {
		result, _ := v.(map[string]any)
		if result == nil {
			result = map[string]any{}
		}
		maps.Copy(result, p)
		return result, nil
	})
	if err != nil {
		return nil, err
	}
	return writeResult, nil
}
