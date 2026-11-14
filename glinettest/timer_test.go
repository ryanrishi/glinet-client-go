package glinettest

import "testing"

// TestTimerLEDMatchesRouterFixtures replays the LED schedule fixtures of the real router.
func TestTimerLEDMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "timer.*_led*.json")
}

// TestTimerRebootMatchesRouterFixtures replays the reboot schedule fixtures of the real router.
func TestTimerRebootMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "timer.*_reboot*.json")
}
