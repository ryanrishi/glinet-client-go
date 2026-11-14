package glinet_test

import (
	"context"
	"testing"

	"github.com/ryanrishi/glinet-client-go/glinettest"
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

func TestDDNSConfigMissingParams(t *testing.T) {
	c, srv := newClient(t)
	srv.DDNS.SetConfig(glinettest.DDNSConfig{Enabled: true})

	err := c.Call(context.Background(), "ddns", "set_config", map[string]any{}, nil)
	if code := rpcCode(t, err); code != -1 {
		t.Errorf("code = %d, want -1", code)
	}
	if !srv.DDNS.Config().Enabled {
		t.Error("set_config without enable_ddns changed the configuration")
	}
}

func TestDDNSStatus(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.DDNS.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, res, fixtureResult(t, "ddns.get_status.disabled.json"))
	if res.Address != "" {
		t.Errorf("Address = %q while disabled, want none", res.Address)
	}

	srv.DDNS.SetConfig(glinettest.DDNSConfig{Enabled: true})
	res, err = c.DDNS.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, res, fixtureResult(t, "ddns.get_status.enabled.json"), "ddns")
	if res.Address != glinettest.DDNSAddress {
		t.Errorf("Address = %q while enabled, want %s", res.Address, glinettest.DDNSAddress)
	}
}
