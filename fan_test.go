package glinet_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestFan(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	handler := srv.Handler("fan", "set_config")
	srv.Handle("fan", "set_config", func(params json.RawMessage) (any, error) {
		sent = append(sent, string(params))
		return handler(params)
	})

	res, err := c.Fan.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if *res != (glinet.FanConfig{Temperature: 75, WarnTemperature: 75}) {
		t.Errorf("GetConfig = %+v", *res)
	}

	if err := c.Fan.SetTemperature(ctx, 80); err != nil {
		t.Fatalf("SetTemperature: %v", err)
	}
	if want := `{"temperature":80}`; len(sent) != 1 || sent[0] != want {
		t.Errorf("params = %v, want %s", sent, want)
	}
	if got := srv.Fan.Config(); got != (glinettest.FanConfig{Temperature: 80, WarnTemperature: 75}) {
		t.Errorf("server state = %+v", got)
	}

	srv.Fan.SetStatus(json.RawMessage(`{"speed":3600,"status":true}`))
	status, err := c.Fan.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if *status != (glinet.FanStatus{Speed: 3600, Running: true}) {
		t.Errorf("GetStatus = %+v", *status)
	}
}
