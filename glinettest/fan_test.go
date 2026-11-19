package glinettest

import "testing"

// TestFanMatchesRouterFixtures replays the fan fixtures of the real router.
func TestFanMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "fan.*.json")
}
