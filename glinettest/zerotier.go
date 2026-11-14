package glinettest

import (
	"encoding/json"
	"sync"
)

// ZeroTierConfig is the configuration of the zerotier module.
type ZeroTierConfig struct {
	Enabled    bool   `json:"enabled"`
	ID         string `json:"id,omitempty"`
	LANEnabled bool   `json:"lan_enabled"`
	WANEnabled bool   `json:"wan_enabled"`
}

// ZeroTier serves the config methods of the zerotier module. Like the router, a set_config that
// does not enable ZeroTier (no enabled, or enabled false) only disables it and keeps the other
// values. A null value gives ErrInvalidParams.
//
// A set_config that enables ZeroTier sets id, lan_enabled and wan_enabled to the sent values, as
// the web UI expects. This was not verified on the router.
type ZeroTier struct {
	mu     sync.Mutex
	config ZeroTierConfig
}

func newZeroTier() *ZeroTier {
	return &ZeroTier{}
}

// Config returns the configuration.
func (m *ZeroTier) Config() ZeroTierConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *ZeroTier) SetConfig(c ZeroTierConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

func (m *ZeroTier) register(s *Server) {
	s.Handle("zerotier", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("zerotier", "set_config", m.setConfig)
}

func (m *ZeroTier) setConfig(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if !truthy(p["enabled"]) {
		m.config.Enabled = false
		return writeResult, nil
	}
	var id string
	if raw, ok := p["id"]; ok {
		if err := json.Unmarshal(raw, &id); err != nil {
			return nil, ErrInvalidParams
		}
	}
	m.config = ZeroTierConfig{Enabled: true, ID: id, LANEnabled: truthy(p["lan_enabled"]), WANEnabled: truthy(p["wan_enabled"])}
	return writeResult, nil
}
