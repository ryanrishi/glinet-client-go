package glinettest

import "testing"

// TestDDNSMatchesRouterFixtures replays the ddns get_config fixture of the real router.
func TestDDNSMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "ddns.get_config.json")
}
