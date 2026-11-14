package glinet_test

import (
	"context"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestZeroTierConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.ZeroTier.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, res, fixtureResult(t, "zerotier.get_config.json"))

	srv.ZeroTier.SetConfig(glinettest.ZeroTierConfig{ID: "0123456789abcdef", LANEnabled: true})
	if err := c.ZeroTier.SetConfig(ctx, glinet.ZeroTierConfig{WANEnabled: true}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	res, err = c.ZeroTier.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if *res != (glinet.ZeroTierConfig{ID: "0123456789abcdef", LANEnabled: true}) {
		t.Errorf("GetConfig = %+v, want the values kept by a disabled write", *res)
	}
}

func TestZeroTierConfigEnable(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	enabled := glinet.ZeroTierConfig{Enabled: true, ID: "0123456789abcdef", LANEnabled: true, WANEnabled: true}
	if err := c.ZeroTier.SetConfig(ctx, enabled); err != nil {
		t.Fatalf("SetConfig enabled: %v", err)
	}
	res, err := c.ZeroTier.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if *res != enabled {
		t.Errorf("GetConfig = %+v, want %+v", *res, enabled)
	}

	if err := c.ZeroTier.SetConfig(ctx, glinet.ZeroTierConfig{}); err != nil {
		t.Fatalf("SetConfig disabled: %v", err)
	}
	res, err = c.ZeroTier.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	disabled := enabled
	disabled.Enabled = false
	if *res != disabled {
		t.Errorf("GetConfig = %+v, want %+v", *res, disabled)
	}
}

func TestZeroTierStatus(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	res, err := c.ZeroTier.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if res.Status != nil {
		t.Errorf("Status = %d while disabled, want nil", *res.Status)
	}

	if err := c.ZeroTier.SetConfig(ctx, glinet.ZeroTierConfig{Enabled: true, ID: "0123456789abcdef"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	res, err = c.ZeroTier.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, res, fixtureResult(t, "zerotier.get_status.enabled.json"))
}

func TestZeroTierConfigEmptyID(t *testing.T) {
	c, srv := newClient(t)
	initial := glinettest.ZeroTierConfig{ID: "0123456789abcdef"}
	srv.ZeroTier.SetConfig(initial)

	err := c.ZeroTier.SetConfig(context.Background(), glinet.ZeroTierConfig{Enabled: true})
	if code := rpcCode(t, err); code != 1 {
		t.Errorf("code = %d, want 1", code)
	}
	if got := srv.ZeroTier.Config(); got != initial {
		t.Errorf("server state = %+v, want %+v", got, initial)
	}
}
