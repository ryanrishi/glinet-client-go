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

// ZeroTier serves the zerotier module like the router:
//   - A set_config that does not enable ZeroTier (no enabled, or enabled false) only disables it
//     and keeps the other values, also id.
//   - A set_config that enables ZeroTier sets id, lan_enabled and wan_enabled to the sent values.
//     With id "" it returns {"err_code":1,"err_msg":"parameter error"} and changes nothing.
//     Without id it returns ErrInternalServerError.
//   - A null value gives ErrInvalidParams.
//   - get_status returns [] while ZeroTier is disabled and {"status":1} ("connecting") while it is
//     enabled. The router returns [] for about 1 s after the enable.
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
	s.Handle("zerotier", "get_status", m.status)
	s.Handle("zerotier", "set_config", m.setConfig)
}

func (m *ZeroTier) status(json.RawMessage) (any, error) {
	if !m.Config().Enabled {
		return json.RawMessage(`[]`), nil
	}
	return map[string]any{"status": 1}, nil
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
	raw, ok := p["id"]
	if !ok {
		return nil, ErrInternalServerError
	}
	var id string
	if err := json.Unmarshal(raw, &id); err != nil {
		return nil, ErrInvalidParams
	}
	if id == "" {
		return map[string]any{"err_code": 1, "err_msg": "parameter error"}, nil
	}
	m.config = ZeroTierConfig{Enabled: true, ID: id, LANEnabled: truthy(p["lan_enabled"]), WANEnabled: truthy(p["wan_enabled"])}
	return writeResult, nil
}
