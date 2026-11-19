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
	WhiteMAC []string `json:"white_mac,omitempty"`
}

// AccessControl serves the black_white_list module. The params mode "white" selects the allow
// list (white_mac), and each other mode the block list (black_mac). Like the router in mode
// "black":
//
//   - set_single_mac with operate "add" appends mac to the list unless the list has it with
//     the same letter case, or mac is a MAC with "-" separators. It does not check mac otherwise.
//     Operate "del" removes mac with the same letter case. A write without mac or with another
//     operate changes nothing. It does not change the mode.
//   - set_config replaces the list with mac, as written, also with duplicates. Without mac,
//     or with mac that is not an array, it returns {"err_code":-1,"err_msg":"mac required"}.
//   - A null value gives ErrInvalidParams.
//
// set_config also sets the mode when the params mode is "black" or "white", as the web UI of
// firmware 4.8.3 uses it to change the mode, and keeps the list of the other mode. On the router
// (firmware 4.8.3), set_single_mac with mode "white" in mode "black" changed the allow list and
// not the mode, as here. The other behaviour of mode "white" and of the mode switch is an
// assumption.
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
	c.WhiteMAC = slices.Clone(c.WhiteMAC)
	return c
}

// SetConfig sets the configuration.
func (m *AccessControl) SetConfig(c AccessControlConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.BlackMAC = slices.Clone(c.BlackMAC)
	c.WhiteMAC = slices.Clone(c.WhiteMAC)
	m.config = c
}

func (m *AccessControl) register(s *Server) {
	s.Handle("black_white_list", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("black_white_list", "set_single_mac", m.setSingleMAC)
	s.Handle("black_white_list", "set_config", m.setConfig)
}

// list returns the list of mode.
func (m *AccessControl) list(mode string) *[]string {
	if mode == "white" {
		return &m.config.WhiteMAC
	}
	return &m.config.BlackMAC
}

func (m *AccessControl) setSingleMAC(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}
	var mode, operate, mac string
	_ = json.Unmarshal(p["mode"], &mode)
	_ = json.Unmarshal(p["operate"], &operate)
	if err := json.Unmarshal(p["mac"], &mac); err != nil || mac == "" {
		return writeResult, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.list(mode)
	switch operate {
	case "add":
		if !dashMAC.MatchString(mac) && !slices.Contains(*list, mac) {
			*list = append(*list, mac)
		}
	case "del":
		*list = slices.DeleteFunc(*list, func(v string) bool { return v == mac })
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
	var mode string
	_ = json.Unmarshal(p["mode"], &mode)

	m.mu.Lock()
	defer m.mu.Unlock()
	if mode == "black" || mode == "white" {
		m.config.Mode = mode
	}
	list := m.list(mode)
	*list = macs
	if len(macs) == 0 {
		*list = nil
	}
	return writeResult, nil
}
