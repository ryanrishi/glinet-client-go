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
	Proto            string   `json:"proto,omitempty"`
	DoTProvider      string   `json:"dot_provider"`
	NextDNSID        string   `json:"nextdns_id"`
	ControlDID       string   `json:"controld_id"`
	FreeControlDID   string   `json:"free_controld_id"`
}

// DNS serves the methods of the dns module.
//
// Like the router, set_config sets force_dns, rebind_protection and override_vpn as flags:
// no value and false are false, and each other value is true. It sets mode when the params have it.
// Only writes in mode "auto" were tested on a router. For the other modes, the fake router
// does what the web UI code expects:
//   - In mode "auto", get_config returns the upstream servers that the router learned in server.
//   - In the other modes, set_config sets server when the params have it, and get_config returns it.
//     A write without server keeps the previous list. The router value of server in mode
//     "secure" with proto "DoT" is not known.
//   - get_config returns proto only in mode "secure".
//   - set_config sets dot_provider, nextdns_id, controld_id and free_controld_id when the params
//     have them, and keeps them otherwise.
//
// set_config returns -32602 when a field has the wrong type. It does not check the values.
//
// get_info returns a small made-up server list.
//
// Like the router, set_host keeps the content as it is and adds "\n" when it does not end with one.
// It returns HTTP status 500 when content is missing or is not a string.
type DNS struct {
	mu       sync.Mutex
	config   DNSConfig
	upstream []string
	hosts    string
}

const defaultHosts = "127.0.0.1 localhost\n\n::1     localhost ip6-localhost ip6-loopback\nff02::1 ip6-allnodes\nff02::2 ip6-allrouters\n"

var dnsInfo = json.RawMessage(`{
	"dnscrypt_version": 2,
	"protos": ["DoT", "DNSCrypt", "DoH", "oDoH"],
	"serverlist": [
		{"name": "example-doh", "proto": "DoH", "dnssec": true, "ipv6": false, "nofilter": true, "nolog": true},
		{"name": "example-doh-ipv6", "proto": "DoH", "dnssec": true, "ipv6": true, "nofilter": false, "nolog": true},
		{"name": "example-dnscrypt", "proto": "DNSCrypt", "dnssec": false, "ipv6": false, "nofilter": true, "nolog": false},
		{"name": "example-odoh", "proto": "oDoH", "dnssec": true, "ipv6": false, "nofilter": true, "nolog": true}
	]
}`)

func newDNS() *DNS {
	upstream := []string{"wwan 192.168.100.53", "wg 10.66.0.1"}
	return &DNS{
		config: DNSConfig{
			Mode:        "auto",
			OverrideVPN: true,
			Server:      slices.Clone(upstream),
			DoTProvider: "0",
		},
		upstream: upstream,
		hosts:    defaultHosts,
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

// SetConfig sets the configuration. In mode "auto", c.Server is also the upstream servers that
// get_config returns after a later write of mode "auto".
func (m *DNS) SetConfig(c DNSConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.Server = slices.Clone(c.Server)
	m.config = c
	if c.Mode == "auto" {
		m.upstream = slices.Clone(c.Server)
	}
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
	s.Handle("dns", "get_info", func(json.RawMessage) (any, error) {
		return dnsInfo, nil
	})
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
	var w struct {
		Mode           *string   `json:"mode"`
		Server         *[]string `json:"server"`
		Proto          *string   `json:"proto"`
		DoTProvider    *string   `json:"dot_provider"`
		NextDNSID      *string   `json:"nextdns_id"`
		ControlDID     *string   `json:"controld_id"`
		FreeControlDID *string   `json:"free_controld_id"`
	}
	if err := json.Unmarshal(params, &w); err != nil {
		return nil, ErrInvalidParams
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	c := &m.config
	c.ForceDNS = truthy(p["force_dns"])
	c.RebindProtection = truthy(p["rebind_protection"])
	c.OverrideVPN = truthy(p["override_vpn"])
	setIfPresent(&c.Mode, w.Mode)
	setIfPresent(&c.Proto, w.Proto)
	setIfPresent(&c.DoTProvider, w.DoTProvider)
	setIfPresent(&c.NextDNSID, w.NextDNSID)
	setIfPresent(&c.ControlDID, w.ControlDID)
	setIfPresent(&c.FreeControlDID, w.FreeControlDID)
	if w.Server != nil {
		c.Server = slices.Clone(*w.Server)
	}
	if c.Mode == "auto" {
		c.Server = slices.Clone(m.upstream)
	}
	if c.Mode != "secure" {
		c.Proto = ""
	}
	return writeResult, nil
}

func setIfPresent[T any](dst *T, v *T) {
	if v != nil {
		*dst = *v
	}
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
