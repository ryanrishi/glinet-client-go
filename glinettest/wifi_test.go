package glinettest

import "testing"

// TestWifiMatchesRouterFixture replays the get_status fixture of the real router.
func TestWifiMatchesRouterFixture(t *testing.T) {
	replayFixtures(t, "wifi.get_status.json")
}
