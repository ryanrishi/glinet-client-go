package glinettest

import "testing"

// TestFirewallMatchesRouterFixtures replays the port forward, rule and DMZ fixtures of the real router.
// The rule and DMZ fixtures also change port forwards and the IDs of the others, so all are one sequence.
func TestFirewallMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "firewall.*port_forward*.json", "firewall.*rule*.json", "firewall.*dmz*.json")
}

// TestFirewallZonesMatchRouterFixture replays the get_zone_list fixture of the real router.
func TestFirewallZonesMatchRouterFixture(t *testing.T) {
	replayFixtures(t, "firewall.get_zone_list.json")
}
