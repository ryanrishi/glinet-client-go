package glinettest

import "testing"

// TestCableMatchesRouterFixture replays the get_status fixture of the real router.
func TestCableMatchesRouterFixture(t *testing.T) {
	replayFixtures(t, "cable.get_status.json")
}
