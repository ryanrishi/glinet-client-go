package glinettest

import (
	"encoding/json"
	"sync"
)

// FanConfig is the configuration of the fan module.
type FanConfig struct {
	Temperature     int `json:"temperature"`
	WarnTemperature int `json:"warn_temperature"`
}

// Fan serves the fan module. The configuration and the status are those of the test router
// until a test sets them. set_config sets Temperature when the params have temperature, and
// keeps WarnTemperature. A temperature that is not an integer gives ErrInvalidParams. The router
// accepted 60 and 62, which are out of the range of the web UI, and kept warn_temperature; that it
// accepts each integer is an assumption. set_config does not change the status.
type Fan struct {
	mu     sync.Mutex
	config FanConfig

	status readResult
}

func newFan() *Fan {
	m := &Fan{config: FanConfig{Temperature: 75, WarnTemperature: 75}}
	m.status.set(json.RawMessage(`{"speed":0,"status":false}`))
	return m
}

// Config returns the configuration.
func (m *Fan) Config() FanConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *Fan) SetConfig(c FanConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

// Status returns the result of get_status.
func (m *Fan) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *Fan) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

func (m *Fan) register(s *Server) {
	s.Handle("fan", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("fan", "get_status", m.status.handle)
	s.Handle("fan", "set_config", m.setConfig)
}

func (m *Fan) setConfig(params json.RawMessage) (any, error) {
	var p struct {
		Temperature *int `json:"temperature"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Temperature != nil {
		m.config.Temperature = *p.Temperature
	}
	return writeResult, nil
}
