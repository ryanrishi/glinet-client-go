package glinettest

import "testing"

// TestLANMatchesRouterFixtures replays the static bind fixtures of the real router.
func TestLANMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "lan.*static_bind*.json")
}

// TestLANWANInfoMatchesRouterFixture replays the get_wan_info fixture of the real router.
func TestLANWANInfoMatchesRouterFixture(t *testing.T) {
	replayFixtures(t, "lan.get_wan_info.json")
}

// TestLANConfigListMatchesRouterFixture replays the get_config_list fixture of the real router.
func TestLANConfigListMatchesRouterFixture(t *testing.T) {
	replayFixtures(t, "lan.get_config_list.json")
}
