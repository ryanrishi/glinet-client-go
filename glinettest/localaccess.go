package glinettest

import (
	"encoding/json"
	"sync"
)

// LocalAccessConfig is the configuration of the local-access module.
type LocalAccessConfig struct {
	HTTPPort          int   `json:"http_port"`
	HTTPSPort         int   `json:"https_port"`
	RedirectHTTPS     bool  `json:"redirect_https"`
	LuCIHTTPPort      int   `json:"luci_http_port"`
	LuCIHTTPSPort     int   `json:"luci_https_port"`
	LuCIRedirectHTTPS bool  `json:"luci_redirect_https"`
	SessionTimeout    int   `json:"session_timeout"`
	SSHEnabled        *bool `json:"ssh_enabled,omitempty"`
	SSHPort           int   `json:"ssh_port"`
}

// LocalAccess serves the local-access module. It starts with the configuration of the test
// router with firmware 4.8.3. A configuration with SSHEnabled nil makes get_config return no
// ssh_enabled, as firmware 4.7.0 does.
//
// set_config sets the fields that the params have and keeps the others. A set_config with
// ssh_enabled adds it also when the configuration has no ssh_enabled; what firmware 4.7.0 does
// is not known. set_config returns ErrInvalidParams when a value is null or has the wrong type. It does not check the ports and
// does not return the err_code values of a port in use (-1, -2, -3) or of a weak password (-4);
// a test can add them with Server.Handle. This is an assumption: the write is from the web UI
// code of firmware 4.8.3 and was not tested on a router.
type LocalAccess struct {
	mu     sync.Mutex
	config LocalAccessConfig
}

func newLocalAccess() *LocalAccess {
	sshEnabled := true
	return &LocalAccess{config: LocalAccessConfig{
		HTTPPort:       80,
		HTTPSPort:      443,
		LuCIHTTPPort:   8080,
		LuCIHTTPSPort:  8443,
		SessionTimeout: 300,
		SSHEnabled:     &sshEnabled,
		SSHPort:        22,
	}}
}

// Config returns the configuration.
func (m *LocalAccess) Config() LocalAccessConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.config.clone()
}

// SetConfig sets the configuration.
func (m *LocalAccess) SetConfig(c LocalAccessConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = c.clone()
}

func (c LocalAccessConfig) clone() LocalAccessConfig {
	if c.SSHEnabled != nil {
		sshEnabled := *c.SSHEnabled
		c.SSHEnabled = &sshEnabled
	}
	return c
}

func (m *LocalAccess) register(s *Server) {
	s.Handle("local-access", "get_config", func(json.RawMessage) (any, error) {
		return m.Config(), nil
	})
	s.Handle("local-access", "set_config", m.setConfig)
}

func (m *LocalAccess) setConfig(params json.RawMessage) (any, error) {
	if _, err := objectParams(params); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.config
	c.SSHEnabled = nil // json.Unmarshal must not write through the pointer of m.config.
	if err := json.Unmarshal(params, &c); err != nil {
		return nil, ErrInvalidParams
	}
	if c.SSHEnabled == nil {
		c.SSHEnabled = m.config.SSHEnabled
	}
	m.config = c
	return writeResult, nil
}
