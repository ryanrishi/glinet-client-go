package glinet_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestIGMPConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var params []string
	handler := srv.Handler("igmp", "set_config")
	srv.Handle("igmp", "set_config", func(p json.RawMessage) (any, error) {
		params = append(params, string(p))
		return handler(p)
	})

	got, err := c.IGMP.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, got, fixtureResult(t, "igmp.get_config.json"))

	for _, config := range []glinet.IGMPConfig{{Enabled: true, Version: 2}, {Enabled: false, Version: 1}} {
		if err := c.IGMP.SetConfig(ctx, config); err != nil {
			t.Fatalf("SetConfig: %v", err)
		}
		got, err := c.IGMP.GetConfig(ctx)
		if err != nil {
			t.Fatalf("GetConfig: %v", err)
		}
		if *got != config {
			t.Errorf("GetConfig = %+v, want %+v", *got, config)
		}
	}
	want := []string{`{"enable":true,"version":2}`, `{"enable":false,"version":1}`}
	if !slices.Equal(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}
}
