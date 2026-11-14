package glinettest

import "testing"

// TestAccessControlMatchesRouterFixtures replays the black_white_list fixtures of the real router.
func TestAccessControlMatchesRouterFixtures(t *testing.T) {
	replayFixtures(t, "black_white_list.*.json")
}
