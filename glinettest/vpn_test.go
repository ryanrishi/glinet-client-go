package glinettest

import "testing"

// TestWireGuardClientMatchesRouterFixtures replays the wg-client read fixtures of the real router.
// get_all_config_list is not replayed: the router leaves some empty peer fields out of it.
func TestWireGuardClientMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "wg-client.get_status.json", "wg-client.get_group_list.json", "wg-client.get_config_list*.json",
		"wg-client.get_setting*.json", "wg-client.get_route_list*.json")
}

// TestWireGuardServerMatchesRouterFixtures replays the wg-server read fixtures of the real router.
func TestWireGuardServerMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "wg-server.*.json")
}

// TestVPNPolicyMatchesRouterFixtures replays the vpn-policy read fixtures of the real router.
func TestVPNPolicyMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "vpn-policy.*.json")
}
