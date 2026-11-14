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

// DDNS serves the config methods of the ddns module. set_config sets enable_ddns as a flag, like
// the other modules of the router: no value and false are false, and each other value is true.
// The write was not verified on the router.
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
	s.Handle("ddns", "set_config", m.setConfig)
}

func (m *DDNS) setConfig(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.Enabled = truthy(p["enable_ddns"])
	return writeResult, nil
}
