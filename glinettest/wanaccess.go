package glinettest

import (
	"encoding/json"
	"slices"
)

// WANAccess is the access to the router from the WAN of the firewall module
// (get_wan_access and set_wan_access).
type WANAccess struct {
	EnableHTTPS     bool              `json:"enable_https"`
	EnableSSH       bool              `json:"enable_ssh"`
	EnablePing      bool              `json:"enable_ping"`
	EnableWhitelist bool              `json:"enable_whitelist"`
	Whitelist       []WANAccessSource `json:"whitelist"`
}

// WANAccessSource is an item of WANAccess.Whitelist.
type WANAccessSource struct {
	IPAddr string `json:"ipaddr"`
	Name   string `json:"name"`
}

// WANAccess returns the access from the WAN.
func (m *Firewall) WANAccess() WANAccess {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.wanAccess
	a.Whitelist = append([]WANAccessSource{}, a.Whitelist...)
	return a
}

// SetWANAccess sets the access from the WAN.
func (m *Firewall) SetWANAccess(a WANAccess) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a.Whitelist = slices.Clone(a.Whitelist)
	m.wanAccess = a
}

func (m *Firewall) setWANAccess(params json.RawMessage) (any, error) {
	if _, err := objectParams(params); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.wanAccess
	a.Whitelist = slices.Clone(a.Whitelist)
	if err := json.Unmarshal(params, &a); err != nil {
		return nil, ErrInvalidParams
	}
	m.wanAccess = a
	return writeResult, nil
}
