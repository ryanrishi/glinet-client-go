package glinettest

import "testing"

// TestNetworkMatchesRouterFixtures replays the network fixtures of the real router.
func TestNetworkMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "network.check_wan_cable.json")
}
