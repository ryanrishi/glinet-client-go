package glinettest

import "testing"

// TestLEDMatchesRouterFixtures replays the led fixtures of the real router.
func TestLEDMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "led.*.json")
}
