package glinet_test

import (
	"context"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestLEDConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.LED.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !res.Enabled {
		t.Errorf("GetConfig = %+v, want enabled", *res)
	}

	if err := c.LED.SetConfig(ctx, glinet.LEDConfig{Enabled: false}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if got := srv.LED.Config(); got.Enabled {
		t.Errorf("server state = %+v, want disabled", got)
	}

	if err := c.LED.SetConfig(ctx, glinet.LEDConfig{Enabled: true}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if got := srv.LED.Config(); got != (glinettest.LEDConfig{Enabled: true}) {
		t.Errorf("server state = %+v, want enabled", got)
	}
}
