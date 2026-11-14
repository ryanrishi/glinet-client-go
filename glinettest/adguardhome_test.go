package glinettest

import "testing"

// TestAdGuardHomeMatchesRouterFixtures replays the adguardhome fixtures of the real router.
func TestAdGuardHomeMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "adguardhome.*.json")
}
