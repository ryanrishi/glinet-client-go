package glinet_test

import (
	"context"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestTailscaleConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.Tailscale.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, res, fixtureResult(t, "tailscale.get_config.json"))

	if err := c.Tailscale.SetConfig(ctx, glinet.SetTailscaleConfigRequest{LANEnabled: true, ExitNodeIP: "100.64.0.1"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	want := glinettest.TailscaleConfig{LANEnabled: true, ExitNodeIP: "100.64.0.1", LANIP: "192.168.8.0/24"}
	if got := srv.Tailscale.Config(); got != want {
		t.Errorf("server state = %+v, want %+v", got, want)
	}

	if err := c.Tailscale.SetConfig(ctx, glinet.SetTailscaleConfigRequest{WANEnabled: true}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	res, err = c.Tailscale.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if *res != (glinet.TailscaleConfig{WANEnabled: true, LANIP: "192.168.8.0/24"}) {
		t.Errorf("GetConfig = %+v, want only wan_enabled and no exit node", *res)
	}
}
