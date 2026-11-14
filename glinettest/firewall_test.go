package glinettest

import "testing"

// TestFirewallMatchesRouterFixtures replays the port forward and rule fixtures of the real router.
// The rule fixtures also change port forwards, so both are one sequence.
func TestFirewallMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "firewall.*port_forward*.json", "firewall.*rule*.json")
}
