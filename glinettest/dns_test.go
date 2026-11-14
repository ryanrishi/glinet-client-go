package glinettest

import "testing"

// TestDNSHostsMatchesRouterFixtures replays the hosts file fixtures of the real router.
func TestDNSHostsMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "dns.*_host*.json")
}

// TestDNSConfigMatchesRouterFixtures replays the DNS config fixtures of the real router.
func TestDNSConfigMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "dns.*_config*.json")
}
