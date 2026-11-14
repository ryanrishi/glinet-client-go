package glinettest

import "encoding/json"

// Network serves the check_wan_cable method of the network module. Its result is that of the
// test router until a test sets it. It ignores the params.
type Network struct {
	wanCable readResult
}

func newNetwork() *Network {
	m := &Network{}
	m.wanCable.set(json.RawMessage(networkWANCable))
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

func (m *Network) register(s *Server) {
	s.Handle("network", "check_wan_cable", m.wanCable.handle)
}
