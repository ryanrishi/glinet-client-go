package glinettest

import "testing"

// TestNetworkModeMatchesRouterFixtures replays the netmode fixtures of the real router.
func TestNetworkModeMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "netmode.*.json")
}
