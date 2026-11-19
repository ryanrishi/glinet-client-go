package glinettest

import (
	"encoding/json"
	"sync"
)

// IGMPConfig is the configuration of the igmp module.
type IGMPConfig struct {
	Enabled bool `json:"enable"`
	Version int  `json:"version"`
}

// IGMP serves the igmp module. The configuration is that of the test router until a test or a
// write sets it.
//
// set_config is an assumption, because it was not tested on a router: it sets enable and version
// when they are in the params, and keeps the other value. It does not check the values.
type IGMP struct {
	mu     sync.Mutex
	config IGMPConfig
}

func newIGMP() *IGMP {
	return &IGMP{config: IGMPConfig{Version: 3}}
}

// Config returns the configuration.
func (m *IGMP) Config() IGMPConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *IGMP) SetConfig(c IGMPConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

func (m *IGMP) register(s *Server) {
	s.Handle("igmp", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("igmp", "set_config", m.setConfig)
}

func (m *IGMP) setConfig(params json.RawMessage) (any, error) {
	var p struct {
		Enabled *bool `json:"enable"`
		Version *int  `json:"version"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Enabled != nil {
		m.config.Enabled = *p.Enabled
	}
	if p.Version != nil {
		m.config.Version = *p.Version
	}
	return writeResult, nil
}
