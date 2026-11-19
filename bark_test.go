package glinet_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestBark(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentRaw(srv, "bark", "set_config", "logout")

	if err := c.Bark.SetConfig(ctx, glinet.BarkConfig{Enable: true}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if err := c.Bark.Logout(ctx); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	checkSent(t, sent, `set_config {"enable":true,"ipv6_enable":false}`, `logout {}`)

	cfg, err := c.Bark.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if *cfg != (glinet.BarkConfig{Enable: true}) {
		t.Errorf("GetConfig = %+v", *cfg)
	}

	srv.Bark.SetStatus(json.RawMessage(`{"status": 3, "pairing_code": "000000000000000"}`))
	status, err := c.Bark.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if *status != (glinet.BarkStatus{Status: 3, PairingCode: "000000000000000"}) {
		t.Errorf("GetStatus = %+v", *status)
	}
}
