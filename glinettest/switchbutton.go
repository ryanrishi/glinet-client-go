package glinettest

import (
	"encoding/json"
	"sync"
)

// SwitchButtonConfig is the configuration of the switch-button module.
type SwitchButtonConfig struct {
	Func    string `json:"func"`
	SubFunc string `json:"sub_func,omitempty"`
}

// SwitchButton serves the switch-button module. The functions of get_funcs have the shape of
// firmware 4.8.3, with the VPN tunnel ID "1"; a test can set them. set_config sets func and
// sub_func as written (no sub_func clears it) and ignores sync. It does not check the values
// against the functions. A func or sub_func that is not a string gives ErrInvalidParams.
// The write was not tested on a router, so this behaviour is an assumption.
type SwitchButton struct {
	mu     sync.Mutex
	config SwitchButtonConfig

	funcs readResult
}

func newSwitchButton() *SwitchButton {
	m := &SwitchButton{config: SwitchButtonConfig{Func: "vpn", SubFunc: "1"}}
	m.funcs.set(json.RawMessage(`{"funcs":[{"func":"adguardhome"},{"func":"led"},{"func":"repeater"},{"func":"tor"},` +
		`{"func":"wifi","sub_func":[{"func":"main_wifi"},{"func":"guest_wifi"}]},{"func":"vpn","sub_func":[{"func":"1"}]}]}`))
	return m
}

// Config returns the configuration.
func (m *SwitchButton) Config() SwitchButtonConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config
}

// SetConfig sets the configuration.
func (m *SwitchButton) SetConfig(c SwitchButtonConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c
}

// Funcs returns the result of get_funcs.
func (m *SwitchButton) Funcs() json.RawMessage {
	return m.funcs.get()
}

// SetFuncs sets the result of get_funcs.
func (m *SwitchButton) SetFuncs(result json.RawMessage) {
	m.funcs.set(result)
}

func (m *SwitchButton) register(s *Server) {
	s.Handle("switch-button", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("switch-button", "get_funcs", m.funcs.handle)
	s.Handle("switch-button", "set_config", m.setConfig)
}

func (m *SwitchButton) setConfig(params json.RawMessage) (any, error) {
	var c SwitchButtonConfig
	if err := json.Unmarshal(params, &c); err != nil {
		return nil, ErrInvalidParams
	}

	m.SetConfig(c)
	return writeResult, nil
}
