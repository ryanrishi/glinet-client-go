package glinettest

import "testing"

// TestCableMatchesRouterFixture replays the get_status and get_config fixtures of the real router.
func TestCableMatchesRouterFixture(t *testing.T) {
	replayFixtures(t, "cable.get_status.json", "cable.get_config.json")
}
