package glinettest

import (
	"encoding/json"
	"sync"
)

// AdGuardHomeConfig is the configuration of the adguardhome module.
type AdGuardHomeConfig struct {
	Enabled    bool `json:"enabled"`
	DNSEnabled bool `json:"dns_enabled"`
}

// AdGuardHome serves the adguardhome module.
type AdGuardHome struct {
	mu     sync.Mutex
	config AdGuardHomeConfig
}

func newAdGuardHome() *AdGuardHome {
	return &AdGuardHome{config: AdGuardHomeConfig{Enabled: false, DNSEnabled: true}}
}

// Config returns the configuration.
func (m *AdGuardHome) Config() AdGuardHomeConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *AdGuardHome) SetConfig(c AdGuardHomeConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

func (m *AdGuardHome) register(s *Server) {
	s.Handle("adguardhome", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("adguardhome", "set_config", m.setConfig)
}

func (m *AdGuardHome) setConfig(params json.RawMessage) (any, error) {
	var p struct {
		Enabled    *bool `json:"enabled"`
		DNSEnabled *bool `json:"dns_enabled"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Enabled != nil {
		m.config.Enabled = *p.Enabled
	}
	if p.DNSEnabled != nil {
		m.config.DNSEnabled = *p.DNSEnabled
	}

	return struct{}{}, nil
}
