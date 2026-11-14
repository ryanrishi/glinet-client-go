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
