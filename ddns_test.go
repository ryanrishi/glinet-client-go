package glinet_test

import (
	"context"
	"testing"
)

func TestDDNSConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.DDNS.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, res, fixtureResult(t, "ddns.get_config.json"))

	if err := c.DDNS.SetEnabled(ctx, true); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	if got := srv.DDNS.Config(); !got.Enabled || got.DeviceID != res.DeviceID {
		t.Errorf("server state = %+v, want enabled and the device ID kept", got)
	}

	if err := c.DDNS.SetEnabled(ctx, false); err != nil {
		t.Fatalf("SetEnabled: %v", err)
	}
	if got := srv.DDNS.Config(); got.Enabled {
		t.Errorf("server state = %+v, want disabled", got)
	}
}
