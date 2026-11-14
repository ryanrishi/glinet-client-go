package glinettest

import "testing"

// TestFirewallMatchesRouterFixtures replays the port forward fixtures of the real router.
func TestFirewallMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "firewall.*port_forward*.json")
}
