package glinettest

import (
	"encoding/json"
	"slices"
	"sync"
)

// MultiWANInterface is one WAN interface in the kmwan get_config result.
type MultiWANInterface struct {
	Interface   string   `json:"interface"`
	EnableCheck bool     `json:"enable_check"`
	EnableSSL   bool     `json:"enable_ssl"`
	Metric      int      `json:"metric"`
	TrackIPv4   []string `json:"track_ipv4"`
	TrackIPv6   []string `json:"track_ipv6"`
	TrackMethod int      `json:"track_method"`
	TrackMode   int      `json:"track_mode"`
	TrackProto  int      `json:"track_proto"`
	Weight      int      `json:"weight"`
}

// MultiWANConfig is the kmwan get_config result.
type MultiWANConfig struct {
	Interfaces []MultiWANInterface `json:"interfaces"`
	Mode       int                 `json:"mode"`
}

// MultiWANSensitivity is the sensitivity in the kmwan get_sensitivity result.
type MultiWANSensitivity struct {
	Level    string  `json:"level"`
	Interval float64 `json:"val"`
}

// MultiWAN serves the kmwan module. The default configuration has the interfaces, metrics, weights
// and track modes of the test router (firmware 4.7.0) with made-up track hosts. The get_status
// result is that of the test router until a test sets it.
//
// The write behaviour is an assumption from the web UI code of firmware 4.8.3, because no kmwan
// write was tested on a router, and get_sensitivity was not called on a router:
//
//   - set_config sets mode, and the metric and the weight that each item of interfaces has, in the
//     interface with the same name. It ignores an unknown interface.
//   - set_interface sets the fields that it gets in the interface with the same name and keeps the
//     other fields. It ignores an unknown interface.
//   - set_sensitivity sets the fields of sensitivity that it gets. get_sensitivity returns
//     {"sensitivity": {...}}, at first level "medium" and val 5, the values that the web UI shows
//     when the result has no sensitivity.
//   - A value of a wrong type gives ErrInvalidParams. No write checks the values.
//
// It does not change get_status after a write.
type MultiWAN struct {
	mu          sync.Mutex
	config      MultiWANConfig
	sensitivity MultiWANSensitivity
	status      readResult
}

func newMultiWAN() *MultiWAN {
	iface := func(name string, metric, trackMode int) MultiWANInterface {
		return MultiWANInterface{
			Interface:   name,
			EnableCheck: true,
			Metric:      metric,
			TrackIPv4:   []string{"192.0.2.1", "192.0.2.2", "198.51.100.1", "198.51.100.2"},
			TrackIPv6:   []string{"2001:db8::1", "2001:db8::2", "2001:db8::3", "2001:db8::4"},
			TrackMode:   trackMode,
			Weight:      1,
		}
	}
	m := &MultiWAN{
		config: MultiWANConfig{Interfaces: []MultiWANInterface{
			iface("wan", 10, 1),
			iface("wwan", 20, 1),
			iface("tethering", 30, 1),
		}},
		sensitivity: MultiWANSensitivity{Level: "medium", Interval: 5},
	}
	m.status.set(json.RawMessage(multiWANStatus))
	return m
}

// Config returns the configuration.
func (m *MultiWAN) Config() MultiWANConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneMultiWANConfig(m.config)
}

// SetConfig sets the configuration.
func (m *MultiWAN) SetConfig(c MultiWANConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cloneMultiWANConfig(c)
}

// Sensitivity returns the sensitivity.
func (m *MultiWAN) Sensitivity() MultiWANSensitivity {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sensitivity
}

// SetSensitivity sets the sensitivity.
func (m *MultiWAN) SetSensitivity(s MultiWANSensitivity) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sensitivity = s
}

// Status returns the result of get_status.
func (m *MultiWAN) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *MultiWAN) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

func cloneMultiWANConfig(c MultiWANConfig) MultiWANConfig {
	c.Interfaces = slices.Clone(c.Interfaces)
	for i := range c.Interfaces {
		c.Interfaces[i].TrackIPv4 = slices.Clone(c.Interfaces[i].TrackIPv4)
		c.Interfaces[i].TrackIPv6 = slices.Clone(c.Interfaces[i].TrackIPv6)
	}
	return c
}

func (m *MultiWAN) register(s *Server) {
	s.Handle("kmwan", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("kmwan", "get_status", m.status.handle)
	s.Handle("kmwan", "get_sensitivity", func(json.RawMessage) (any, error) {
		return map[string]any{"sensitivity": m.Sensitivity()}, nil
	})
	s.Handle("kmwan", "set_config", m.setConfig)
	s.Handle("kmwan", "set_interface", m.setInterface)
	s.Handle("kmwan", "set_sensitivity", m.setSensitivity)
}

// decodeMultiWANParams decodes params, which must be a JSON object, into v.
func decodeMultiWANParams(params json.RawMessage, v any) error {
	if _, err := paramsObject(params); err != nil {
		return err
	}
	if err := json.Unmarshal(params, v); err != nil {
		return ErrInvalidParams
	}
	return nil
}

// find returns the interface with the name, or nil.
func (m *MultiWAN) find(name string) *MultiWANInterface {
	for i := range m.config.Interfaces {
		if m.config.Interfaces[i].Interface == name {
			return &m.config.Interfaces[i]
		}
	}
	return nil
}

func (m *MultiWAN) setConfig(params json.RawMessage) (any, error) {
	var p struct {
		Mode       *int `json:"mode"`
		Interfaces []struct {
			Interface string `json:"interface"`
			Metric    *int   `json:"metric"`
			Weight    *int   `json:"weight"`
		} `json:"interfaces"`
	}
	if err := decodeMultiWANParams(params, &p); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Mode != nil {
		m.config.Mode = *p.Mode
	}
	for _, item := range p.Interfaces {
		iface := m.find(item.Interface)
		if iface == nil {
			continue
		}
		if item.Metric != nil {
			iface.Metric = *item.Metric
		}
		if item.Weight != nil {
			iface.Weight = *item.Weight
		}
	}
	return writeResult, nil
}

func (m *MultiWAN) setInterface(params json.RawMessage) (any, error) {
	var p struct {
		Interface   string    `json:"interface"`
		EnableCheck *bool     `json:"enable_check"`
		EnableSSL   *bool     `json:"enable_ssl"`
		TrackIPv4   *[]string `json:"track_ipv4"`
		TrackIPv6   *[]string `json:"track_ipv6"`
		TrackMethod *int      `json:"track_method"`
		TrackMode   *int      `json:"track_mode"`
		TrackProto  *int      `json:"track_proto"`
	}
	if err := decodeMultiWANParams(params, &p); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	iface := m.find(p.Interface)
	if iface == nil {
		return writeResult, nil
	}
	if p.EnableCheck != nil {
		iface.EnableCheck = *p.EnableCheck
	}
	if p.EnableSSL != nil {
		iface.EnableSSL = *p.EnableSSL
	}
	if p.TrackIPv4 != nil {
		iface.TrackIPv4 = slices.Clone(*p.TrackIPv4)
	}
	if p.TrackIPv6 != nil {
		iface.TrackIPv6 = slices.Clone(*p.TrackIPv6)
	}
	if p.TrackMethod != nil {
		iface.TrackMethod = *p.TrackMethod
	}
	if p.TrackMode != nil {
		iface.TrackMode = *p.TrackMode
	}
	if p.TrackProto != nil {
		iface.TrackProto = *p.TrackProto
	}
	return writeResult, nil
}

func (m *MultiWAN) setSensitivity(params json.RawMessage) (any, error) {
	var p struct {
		Sensitivity struct {
			Level    *string  `json:"level"`
			Interval *float64 `json:"val"`
		} `json:"sensitivity"`
	}
	if err := decodeMultiWANParams(params, &p); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Sensitivity.Level != nil {
		m.sensitivity.Level = *p.Sensitivity.Level
	}
	if p.Sensitivity.Interval != nil {
		m.sensitivity.Interval = *p.Sensitivity.Interval
	}
	return writeResult, nil
}

// multiWANStatus is the kmwan get_status result of the test router (firmware 4.7.0).
const multiWANStatus = `{
	"interfaces": [
		{
			"interface": "wan",
			"status_v4": 1,
			"status_v6": 1
		},
		{
			"interface": "wwan",
			"status_v4": 0,
			"status_v6": 1
		},
		{
			"interface": "tethering",
			"status_v4": 1,
			"status_v6": 1
		}
	]
}`
