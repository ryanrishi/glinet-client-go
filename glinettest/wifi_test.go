package glinettest

import "testing"

// TestWifiMatchesRouterFixture replays the get_status and get_config fixtures of the real router.
func TestWifiMatchesRouterFixture(t *testing.T) {
	replayFixtures(t, "wifi.get_status.json", "wifi.get_config.json")
}
