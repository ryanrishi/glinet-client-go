package glinettest

import "testing"

func TestLocalAccessConcurrentPartialWrites(t *testing.T) {
	srv := NewServer(t)
	checkConcurrentPartialWrites(t, srv.Handler("local-access", "set_config"), func() any { return srv.LocalAccess.Config() }, map[string][]string{
		"http_port":       numbers(500),
		"https_port":      numbers(500),
		"luci_http_port":  numbers(500),
		"luci_https_port": numbers(500),
		"session_timeout": numbers(500),
		"ssh_port":        numbers(500),
	})
}
