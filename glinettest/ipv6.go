package glinettest

import (
	"encoding/json"
	"maps"
)

// IPv6 serves the get_ipv6 and set_ipv6 methods of the ipv6 module. The configuration is that of
// the test router until a test or a write sets it.
//
// The write behaviour is an assumption, because set_ipv6 was not verified on a router: set_ipv6
// sets each field that it gets and keeps the others. Not known: which fields the router keeps when
// a write does not send them. It does not check the values.
type IPv6 struct {
	config readResult
}

func newIPv6() *IPv6 {
	m := &IPv6{}
	m.config.set(json.RawMessage(ipv6Config))
	return m
}

// Config returns the result of get_ipv6.
func (m *IPv6) Config() json.RawMessage {
	return m.config.get()
}

// SetConfig sets the result of get_ipv6.
func (m *IPv6) SetConfig(result json.RawMessage) {
	m.config.set(result)
}

func (m *IPv6) register(s *Server) {
	s.Handle("ipv6", "get_ipv6", m.config.handle)
	s.Handle("ipv6", "set_ipv6", m.setIPv6)
}

func (m *IPv6) setIPv6(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}

	err = m.config.update(func(v any) (any, error) {
		config, _ := v.(map[string]any)
		if config == nil {
			config = map[string]any{}
		}
		maps.Copy(config, p)
		return config, nil
	})
	if err != nil {
		return nil, err
	}
	return writeResult, nil
}
