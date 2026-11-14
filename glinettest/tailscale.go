package glinettest

import (
	"encoding/json"
	"sync"
)

// TailscaleConfig is the configuration of the tailscale module.
type TailscaleConfig struct {
	Enabled    bool   `json:"enabled"`
	LANEnabled bool   `json:"lan_enabled"`
	WANEnabled bool   `json:"wan_enabled"`
	ExitNodeIP string `json:"exit_node_ip,omitempty"`
	LANIP      string `json:"lan_ip"`
}

// Tailscale serves the config methods of the tailscale module. Like the router, set_config sets
// enabled, lan_enabled and wan_enabled as flags (no value and false are false, and each other
// value is true) and sets exit_node_ip to the sent string, or to "" without it. It keeps lan_ip.
// A null value gives ErrInvalidParams. The router was verified with enabled false only.
type Tailscale struct {
	mu     sync.Mutex
	config TailscaleConfig
}

func newTailscale() *Tailscale {
	return &Tailscale{config: TailscaleConfig{LANIP: "192.168.8.0/24"}}
}

// Config returns the configuration.
func (m *Tailscale) Config() TailscaleConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *Tailscale) SetConfig(c TailscaleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

func (m *Tailscale) register(s *Server) {
	s.Handle("tailscale", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("tailscale", "set_config", m.setConfig)
}

func (m *Tailscale) setConfig(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}
	var exitNodeIP string
	if raw, ok := p["exit_node_ip"]; ok {
		if err := json.Unmarshal(raw, &exitNodeIP); err != nil {
			return nil, ErrInvalidParams
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.Enabled = truthy(p["enabled"])
	m.config.LANEnabled = truthy(p["lan_enabled"])
	m.config.WANEnabled = truthy(p["wan_enabled"])
	m.config.ExitNodeIP = exitNodeIP
	return writeResult, nil
}
