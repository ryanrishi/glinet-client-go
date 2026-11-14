package glinettest

import (
	"encoding/json"
	"sync"
)

// LEDConfig is the configuration of the led module.
type LEDConfig struct {
	Enabled bool `json:"led_enable"`
}

// LED serves the led module. Like the router, set_config without led_enable disables the LEDs,
// and each led_enable value other than false enables them, also 0 and "false".
type LED struct {
	mu     sync.Mutex
	config LEDConfig
}

func newLED() *LED {
	return &LED{config: LEDConfig{Enabled: true}}
}

// Config returns the configuration.
func (m *LED) Config() LEDConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *LED) SetConfig(c LEDConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

func (m *LED) register(s *Server) {
	s.Handle("led", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("led", "set_config", m.setConfig)
}

func (m *LED) setConfig(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}

	m.SetConfig(LEDConfig{Enabled: truthy(p["led_enable"])})
	return writeResult, nil
}
