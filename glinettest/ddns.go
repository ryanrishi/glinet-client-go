package glinettest

import (
	"encoding/json"
	"sync"
)

// DDNSConfig is the configuration of the ddns module.
type DDNSConfig struct {
	Enabled  bool   `json:"enable_ddns"`
	DeviceID string `json:"device_id"`
}

// DDNSAddress is the address in the ddns get_status result while DDNS is enabled.
const DDNSAddress = "203.0.113.1"

// DDNS serves the ddns module. Like the router, set_config sets enable_ddns as a flag (false is
// false, and each other value is true), and a set_config without enable_ddns returns
// {"err_code":-1,"err_msg":"missing parameters"} and changes nothing.
//
// get_status has the WAN addresses of the test router and status 2. While DDNS is enabled, it
// also has DDNSAddress at once; the router adds it some seconds after the write.
type DDNS struct {
	mu     sync.Mutex
	config DDNSConfig
}

func newDDNS() *DDNS {
	return &DDNS{config: DDNSConfig{DeviceID: "redacted-ddns-1"}}
}

// Config returns the configuration.
func (m *DDNS) Config() DDNSConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *DDNS) SetConfig(c DDNSConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

func (m *DDNS) register(s *Server) {
	s.Handle("ddns", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("ddns", "get_status", m.status)
	s.Handle("ddns", "set_config", m.setConfig)
}

func (m *DDNS) status(json.RawMessage) (any, error) {
	status := map[string]any{
		"ips": []map[string]any{
			{"interface": "wan6", "ip": []string{}},
			{"interface": "wwan", "ip": []string{"192.168.100.20"}},
			{"interface": "wwan6", "ip": []string{}},
		},
		"status": 2,
	}
	if m.Config().Enabled {
		status["ddns"] = DDNSAddress
	}
	return status, nil
}

func (m *DDNS) setConfig(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}
	enabled, ok := p["enable_ddns"]
	if !ok {
		return map[string]any{"err_code": -1, "err_msg": "missing parameters"}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.Enabled = truthy(enabled)
	return writeResult, nil
}
