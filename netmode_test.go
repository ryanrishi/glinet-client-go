package glinet_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestNetworkModeGetModeDecodesRouterFixture(t *testing.T) {
	c, _ := newClient(t)

	got, err := c.NetworkMode.GetMode(context.Background())
	if err != nil {
		t.Fatalf("GetMode: %v", err)
	}
	if got.Mode != glinet.NetworkModeRouter {
		t.Errorf("Mode = %q, want router", got.Mode)
	}
	checkLossless(t, got, fixtureResult(t, "netmode.get_mode.json"))
}

func TestNetworkModeSetMode(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var params []string
	handler := srv.Handler("netmode", "set_mode")
	srv.Handle("netmode", "set_mode", func(p json.RawMessage) (any, error) {
		params = append(params, string(p))
		return handler(p)
	})

	tests := []struct {
		settings glinet.NetworkModeSettings
		want     glinet.NetworkModeConfig
		lanIP    string
	}{
		{glinet.NetworkModeSettings{Mode: glinet.NetworkModeAccessPoint, SSID: "ignored", Mask: "24"}, glinet.NetworkModeConfig{Mode: "ap"}, ""},
		{glinet.NetworkModeSettings{Mode: glinet.NetworkModeRouter}, glinet.NetworkModeConfig{Mode: "router"}, glinettest.NetworkModeLANIP},
		{glinet.NetworkModeSettings{Mode: glinet.NetworkModeExtender, SSID: "example"}, glinet.NetworkModeConfig{Mode: "relay"}, ""},
		{glinet.NetworkModeSettings{Mode: glinet.NetworkModeWDS, SSID: "example", Key: "example-key", Identity: "example-user"}, glinet.NetworkModeConfig{Mode: "wds"}, ""},
		{glinet.NetworkModeSettings{Mode: glinet.NetworkModePassthrough, Mask: "auto"}, glinet.NetworkModeConfig{Mode: "passthrough", Mask: "auto"}, ""},
		{glinet.NetworkModeSettings{Mode: glinet.NetworkModePassthrough, Mask: "24", MAC: "02:00:00:00:00:01"}, glinet.NetworkModeConfig{Mode: "passthrough", Mask: "24", MAC: "02:00:00:00:00:01"}, ""},
	}
	for _, tt := range tests {
		res, err := c.NetworkMode.SetMode(ctx, tt.settings)
		if err != nil {
			t.Fatalf("SetMode(%+v): %v", tt.settings, err)
		}
		if res.LANIP != tt.lanIP {
			t.Errorf("SetMode(%+v) LANIP = %q, want %q", tt.settings, res.LANIP, tt.lanIP)
		}
		got, err := c.NetworkMode.GetMode(ctx)
		if err != nil {
			t.Fatalf("GetMode: %v", err)
		}
		if *got != tt.want {
			t.Errorf("GetMode = %+v, want %+v", *got, tt.want)
		}
	}

	want := []string{
		`{"mode":"ap"}`,
		`{"mode":"router"}`,
		`{"mode":"relay","ssid":"example"}`,
		`{"identity":"example-user","key":"example-key","mode":"wds","ssid":"example"}`,
		`{"mac":"","mask":"auto","mode":"passthrough"}`,
		`{"mac":"02:00:00:00:00:01","mask":"24","mode":"passthrough"}`,
	}
	if !slices.Equal(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}
}

func TestNetworkModeSetModeError(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("netmode", "set_mode", func(json.RawMessage) (any, error) {
		return map[string]any{"err_code": 4, "err_msg": "example"}, nil
	})

	_, err := c.NetworkMode.SetMode(context.Background(), glinet.NetworkModeSettings{Mode: glinet.NetworkModeExtender, SSID: "example", Key: "wrong"})
	if code := rpcCode(t, err); code != 4 {
		t.Errorf("code = %d, want 4", code)
	}
}
