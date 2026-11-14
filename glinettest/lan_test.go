package glinettest

import "testing"

// TestLANMatchesRouterFixtures replays the static bind fixtures of the real router.
func TestLANMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "lan.*static_bind*.json")
}
