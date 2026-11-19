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

func TestFirewallDMZDoesNotSharePriority(t *testing.T) {
	srv := NewServer(t)
	higher := true
	srv.Firewall.SetDMZ(DMZ{Enabled: true, DestIP: "198.51.100.50", Priority: &higher})

	higher = false
	*srv.Firewall.DMZ().Priority = false
	if got := srv.Firewall.DMZ().Priority; got == nil || !*got {
		t.Errorf("Priority = %v, want true: a change through a pointer of SetDMZ or DMZ changed the state", got)
	}
}
