package glinettest

import "testing"

// TestIGMPMatchesRouterFixtures replays the igmp fixtures of the real router.
func TestIGMPMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "igmp.*.json")
}
