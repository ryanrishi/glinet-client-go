package glinettest

import "testing"

// TestIPv6MatchesRouterFixture replays the get_ipv6 fixture of the real router.
func TestIPv6MatchesRouterFixture(t *testing.T) {
	replayFixtures(t, "ipv6.get_ipv6.json")
}
