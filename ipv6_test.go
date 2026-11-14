package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestIPv6ConfigDecodesRouterFixture(t *testing.T) {
	c, _ := newClient(t)

	got, err := c.IPv6.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	want := glinet.IPv6Config{LANMode: "nat6", LANDNSMode: true}
	if *got != want {
		t.Errorf("GetConfig = %+v, want %+v", *got, want)
	}
	checkLossless(t, got, fixtureResult(t, "ipv6.get_ipv6.json"))
}

func TestIPv6SetConfigSendsWebUIParams(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent map[string]any
	setIPv6 := srv.Handler("ipv6", "set_ipv6")
	srv.Handle("ipv6", "set_ipv6", func(params json.RawMessage) (any, error) {
		sent = nil
		_ = json.Unmarshal(params, &sent)
		return setIPv6(params)
	})

	tests := []struct {
		name   string
		config glinet.IPv6Config
		want   map[string]any
	}{
		{
			name:   "disable",
			config: glinet.IPv6Config{LANMode: "static", LANIP: "fd00::1/64", LANDNS1: "fd00::53"},
			want:   map[string]any{"enable": false},
		},
		{
			name:   "nat6 with automatic DNS",
			config: glinet.IPv6Config{Enable: true, LANMode: "nat6", LANDNSMode: true, LANIP: "fd00::1/64", LANDNS1: "fd00::53"},
			want:   map[string]any{"enable": true, "lan_mode": "nat6", "lan_dns_mode": true},
		},
		{
			name:   "static with manual DNS",
			config: glinet.IPv6Config{Enable: true, LANMode: "static", LANIP: "fd00::1/64", LANDNS1: "fd00::53"},
			want:   map[string]any{"enable": true, "lan_mode": "static", "lan_dns_mode": false, "lan_ip": "fd00::1/64", "lan_dns1": "fd00::53"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.IPv6.SetConfig(ctx, tt.config); err != nil {
				t.Fatalf("SetConfig: %v", err)
			}
			if !reflect.DeepEqual(sent, tt.want) {
				t.Errorf("params = %v, want %v", sent, tt.want)
			}
		})
	}

	got, err := c.IPv6.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	want := glinet.IPv6Config{Enable: true, LANMode: "static", LANIP: "fd00::1/64", LANDNS1: "fd00::53"}
	if *got != want {
		t.Errorf("GetConfig after the writes = %+v, want %+v", *got, want)
	}
}
