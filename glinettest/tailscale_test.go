package glinettest

import "testing"

// TestTailscaleMatchesRouterFixtures replays the tailscale config fixtures of the real router.
func TestTailscaleMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "tailscale.*_config*.json")
}
