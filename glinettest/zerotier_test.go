package glinettest

import "testing"

// TestZeroTierMatchesRouterFixtures replays the zerotier config fixtures of the real router.
func TestZeroTierMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "zerotier.*_config*.json")
}
