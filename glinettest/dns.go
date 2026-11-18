package glinettest

import (
	"encoding/json"
	"slices"
	"strings"
	"sync"
)

// DNSConfig is the configuration of the dns module.
type DNSConfig struct {
	Mode             string   `json:"mode"`
	ForceDNS         bool     `json:"force_dns"`
	RebindProtection bool     `json:"rebind_protection"`
	OverrideVPN      bool     `json:"override_vpn"`
	Server           []string `json:"server"`
	DoTProvider      string   `json:"dot_provider"`
	NextDNSID        string   `json:"nextdns_id"`
	ControlDID       string   `json:"controld_id"`
	FreeControlDID   string   `json:"free_controld_id"`
}

// DNS serves the config and hosts file methods of the dns module.
//
// Like the router, set_config sets force_dns, rebind_protection and override_vpn as flags:
// no value and false are false, and each other value is true. It sets mode when the params have it
// and ignores the other fields. Server is the list that get_config returns; set_config does not change it.
//
// Like the router, set_host keeps the content as it is and adds "\n" when it does not end with one.
// It returns HTTP status 500 when content is missing or is not a string.
type DNS struct {
	mu     sync.Mutex
	config DNSConfig
	hosts  string
}

const defaultHosts = "127.0.0.1 localhost\n\n::1     localhost ip6-localhost ip6-loopback\nff02::1 ip6-allnodes\nff02::2 ip6-allrouters\n"

func newDNS() *DNS {
	return &DNS{
		config: DNSConfig{
			Mode:        "auto",
			OverrideVPN: true,
			Server:      []string{"wwan 192.168.100.53", "wg 10.66.0.1"},
			DoTProvider: "0",
		},
		hosts: defaultHosts,
	}
}

// Config returns the configuration.
func (m *DNS) Config() DNSConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.config
	c.Server = slices.Clone(c.Server)
	return c
}

// SetConfig sets the configuration.
func (m *DNS) SetConfig(c DNSConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.Server = slices.Clone(c.Server)
	m.config = c
}

// Hosts returns the hosts file.
func (m *DNS) Hosts() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hosts
}

// SetHosts sets the hosts file.
func (m *DNS) SetHosts(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hosts = content
}

func (m *DNS) register(s *Server) {
	s.Handle("dns", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("dns", "set_config", m.setConfig)
	s.Handle("dns", "get_host", func(json.RawMessage) (any, error) {
		return map[string]string{"content": m.Hosts()}, nil
	})
	s.Handle("dns", "set_host", m.setHost)
}

func (m *DNS) setConfig(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}
	var mode *string
	if raw, ok := p["mode"]; ok {
		if err := json.Unmarshal(raw, &mode); err != nil {
			return nil, ErrInvalidParams
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.ForceDNS = truthy(p["force_dns"])
	m.config.RebindProtection = truthy(p["rebind_protection"])
	m.config.OverrideVPN = truthy(p["override_vpn"])
	if mode != nil {
		m.config.Mode = *mode
	}
	return writeResult, nil
}

func (m *DNS) setHost(params json.RawMessage) (any, error) {
	p, err := objectParams(params)
	if err != nil {
		return nil, err
	}
	var content string
	if err := json.Unmarshal(p["content"], &content); err != nil {
		return nil, ErrInternalServerError
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	m.SetHosts(content)
	return writeResult, nil
}
