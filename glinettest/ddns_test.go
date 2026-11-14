package glinettest

import "testing"

// TestDDNSMatchesRouterFixtures replays the ddns config fixtures of the real router.
func TestDDNSMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "ddns.*_config*.json")
}
