package glinettest

import (
	"encoding/json"
	"sync"
)

// NetworkModeConfig is the configuration of the netmode module. Mask and MAC are for the mode
// "passthrough".
type NetworkModeConfig struct {
	Mode string `json:"mode"`
	Mask string `json:"mask,omitempty"`
	MAC  string `json:"mac,omitempty"`
}

// NetworkModeLANIP is the lanip in the set_mode result of a change to the mode "router".
const NetworkModeLANIP = "192.168.8.1"

// NetworkMode serves the netmode module. The mode is "router", as on the test router, until a test
// or a write sets it.
//
// set_mode is an assumption, because it was not tested on a router: it sets mode, and for the mode
// "passthrough" also mask and mac. It returns {"lanip": NetworkModeLANIP} for the mode "router"
// and [] for the other modes. It does not check the values, does not connect to a Wi-Fi network
// and does not change the other modules.
type NetworkMode struct {
	mu     sync.Mutex
	config NetworkModeConfig
}

func newNetworkMode() *NetworkMode {
	return &NetworkMode{config: NetworkModeConfig{Mode: "router"}}
}

// Config returns the configuration.
func (m *NetworkMode) Config() NetworkModeConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *NetworkMode) SetConfig(c NetworkModeConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

func (m *NetworkMode) register(s *Server) {
	s.Handle("netmode", "get_mode", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("netmode", "set_mode", m.setMode)
}

func (m *NetworkMode) setMode(params json.RawMessage) (any, error) {
	var p NetworkModeConfig
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}
	config := NetworkModeConfig{Mode: p.Mode}
	if p.Mode == "passthrough" {
		config.Mask, config.MAC = p.Mask, p.MAC
	}
	m.SetConfig(config)

	if p.Mode == "router" {
		return map[string]string{"lanip": NetworkModeLANIP}, nil
	}
	return writeResult, nil
}
