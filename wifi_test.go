package glinet_test

import (
	"context"
	"testing"
)

func TestWifiStatusDecodesRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	status := fixtureResult(t, "wifi.get_status.json")
	srv.Wifi.SetStatus(status)

	got, err := c.Wifi.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, map[string]any{"res": got}, status)
}
