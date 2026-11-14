package glinettest

import (
	"encoding/json"
	"regexp"
	"slices"
	"sync"
)

var dashMAC = regexp.MustCompile(`^[0-9A-Fa-f]{2}(-[0-9A-Fa-f]{2}){5}$`)

// AccessControlConfig is the configuration of the black_white_list module.
type AccessControlConfig struct {
	Mode     string   `json:"mode"`
	BlackMAC []string `json:"black_mac,omitempty"`
}

// AccessControl serves the black_white_list module in mode "black". Like the router:
//
//   - set_single_mac with operate "add" appends mac to the block list unless the list has it with
//     the same letter case, or mac is a MAC with "-" separators. It does not check mac otherwise.
//     Operate "del" removes mac with the same letter case. A write without mac or with another
//     operate changes nothing.
//   - set_config replaces the block list with mac, as written, also with duplicates. Without mac,
//     or with mac that is not an array, it returns {"err_code":-1,"err_msg":"mac required"}.
//   - A null value gives ErrInvalidParams.
//
// The writes do not change the mode. The router was only tested with mode "black".
type AccessControl struct {
	mu     sync.Mutex
	config AccessControlConfig
}

func newAccessControl() *AccessControl {
	return &AccessControl{config: AccessControlConfig{Mode: "black"}}
}

// Config returns the configuration.
func (m *AccessControl) Config() AccessControlConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.config
	c.BlackMAC = slices.Clone(c.BlackMAC)
	return c
}

// SetConfig sets the configuration.
func (m *AccessControl) SetConfig(c AccessControlConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.BlackMAC = slices.Clone(c.BlackMAC)
	m.config = c
}

func (m *AccessControl) register(s *Server) {
	s.Handle("black_white_list", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("black_white_list", "set_single_mac", m.setSingleMAC)
	s.Handle("black_white_list", "set_config", m.setConfig)
}

func (m *AccessControl) setSingleMAC(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}
	var operate, mac string
	_ = json.Unmarshal(p["operate"], &operate)
	if err := json.Unmarshal(p["mac"], &mac); err != nil || mac == "" {
		return writeResult, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	switch operate {
	case "add":
		if !dashMAC.MatchString(mac) && !slices.Contains(m.config.BlackMAC, mac) {
			m.config.BlackMAC = append(m.config.BlackMAC, mac)
		}
	case "del":
		m.config.BlackMAC = slices.DeleteFunc(m.config.BlackMAC, func(v string) bool { return v == mac })
	}
	return writeResult, nil
}

func (m *AccessControl) setConfig(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}
	var macs []string
	if err := json.Unmarshal(p["mac"], &macs); err != nil {
		return map[string]any{"err_code": -1, "err_msg": "mac required"}, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.BlackMAC = macs
	if len(macs) == 0 {
		m.config.BlackMAC = nil
	}
	return writeResult, nil
}
