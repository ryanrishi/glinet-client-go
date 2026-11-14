package glinettest

import (
	"encoding/json"
	"strings"
	"sync"
)

// DNS serves the hosts file methods of the dns module. Like the router, set_host keeps the
// content as it is and adds "\n" when it does not end with one. It returns HTTP status 500
// when content is missing or is not a string.
type DNS struct {
	mu    sync.Mutex
	hosts string
}

const defaultHosts = "127.0.0.1 localhost\n\n::1     localhost ip6-localhost ip6-loopback\nff02::1 ip6-allnodes\nff02::2 ip6-allrouters\n"

func newDNS() *DNS {
	return &DNS{hosts: defaultHosts}
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
	s.Handle("dns", "get_host", func(json.RawMessage) (any, error) {
		return map[string]string{"content": m.Hosts()}, nil
	})
	s.Handle("dns", "set_host", m.setHost)
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
