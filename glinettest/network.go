package glinettest

import "encoding/json"

// Network serves the check_wan_cable and get_arp_list methods of the network module. Their
// results are those of the test router until a test sets them. check_wan_cable ignores the params.
type Network struct {
	wanCable, arpList readResult
}

func newNetwork() *Network {
	m := &Network{}
	m.wanCable.set(json.RawMessage(networkWANCable))
	m.arpList.set(json.RawMessage(networkARPList))
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

func (m *Network) register(s *Server) {
	s.Handle("network", "check_wan_cable", m.wanCable.handle)
	s.Handle("network", "get_arp_list", m.arpList.handle)
}
