package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

// recordMultiWANWrites records the params of each kmwan write.
func recordMultiWANWrites(srv *glinettest.Server) *[]string {
	var sent []string
	for _, method := range []string{"set_config", "set_interface", "set_sensitivity"} {
		handler := srv.Handler("kmwan", method)
		srv.Handle("kmwan", method, func(params json.RawMessage) (any, error) {
			sent = append(sent, method+" "+string(params))
			return handler(params)
		})
	}
	return &sent
}

func TestMultiWANMode(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := recordMultiWANWrites(srv)

	if err := c.MultiWAN.SetLoadBalance(ctx, []glinet.MultiWANWeight{{Interface: "wan", Weight: 3}, {Interface: "wwan", Weight: 0}}); err != nil {
		t.Fatalf("SetLoadBalance: %v", err)
	}
	got, err := c.MultiWAN.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Mode != glinet.MultiWANLoadBalance || got.Interfaces[0].Weight != 3 || got.Interfaces[1].Weight != 0 || got.Interfaces[2].Weight != 1 || got.Interfaces[0].Metric != 10 {
		t.Errorf("GetConfig after SetLoadBalance = %+v", *got)
	}

	if err := c.MultiWAN.SetFailover(ctx, []string{"wwan", "tethering", "wan"}); err != nil {
		t.Fatalf("SetFailover: %v", err)
	}
	got, err = c.MultiWAN.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	metrics := map[string]int{}
	for _, iface := range got.Interfaces {
		metrics[iface.Interface] = iface.Metric
	}
	if got.Mode != glinet.MultiWANFailover || !reflect.DeepEqual(metrics, map[string]int{"wwan": 1, "tethering": 2, "wan": 3}) || got.Interfaces[0].Weight != 3 {
		t.Errorf("GetConfig after SetFailover = %+v", *got)
	}

	if err := c.MultiWAN.SetLoadBalance(ctx, nil); err != nil {
		t.Fatalf("SetLoadBalance: %v", err)
	}

	want := []string{
		`set_config {"mode":1,"interfaces":[{"interface":"wan","weight":3},{"interface":"wwan","weight":0}]}`,
		`set_config {"mode":0,"interfaces":[{"interface":"wwan","metric":1},{"interface":"tethering","metric":2},{"interface":"wan","metric":3}]}`,
		`set_config {"mode":1,"interfaces":[]}`,
	}
	if !slices.Equal(*sent, want) {
		t.Errorf("params = %v, want %v", *sent, want)
	}
}

func TestMultiWANSetInterface(t *testing.T) {
	ipv4 := []string{"192.0.2.10", "192.0.2.11"}
	ipv6 := []string{"2001:db8::10", "2001:db8::11", "2001:db8::12"}
	tests := []struct {
		name     string
		tracking glinet.MultiWANTracking
		want     string
	}{
		{
			name:     "no check",
			tracking: glinet.MultiWANTracking{Interface: "wan", TrackMode: glinet.MultiWANTrackModeStrict, TrackIPv4: ipv4},
			want:     `{"interface":"wan","enable_check":false}`,
		},
		{
			name: "IPv4",
			tracking: glinet.MultiWANTracking{
				Interface: "wan", EnableCheck: true, TrackMode: glinet.MultiWANTrackModeLowData, TrackProto: glinet.MultiWANTrackProtoIPv4,
				TrackMethod: glinet.MultiWANTrackMethodPing, TrackIPv4: ipv4, TrackIPv6: ipv6, EnableSSL: true,
			},
			want: `{"interface":"wan","enable_check":true,"track_proto":0,"track_method":0,"track_mode":0,"track_ipv4":["192.0.2.10","192.0.2.11"]}`,
		},
		{
			name: "IPv6",
			tracking: glinet.MultiWANTracking{
				Interface: "wwan", EnableCheck: true, TrackMode: glinet.MultiWANTrackModeNormal, TrackProto: glinet.MultiWANTrackProtoIPv6, TrackIPv4: ipv4, TrackIPv6: ipv6,
			},
			want: `{"interface":"wwan","enable_check":true,"track_proto":1,"track_method":0,"track_mode":1,"track_ipv6":["2001:db8::10","2001:db8::11","2001:db8::12"]}`,
		},
		{
			name: "both with method 1",
			tracking: glinet.MultiWANTracking{
				Interface: "tethering", EnableCheck: true, TrackMode: glinet.MultiWANTrackModeStrict, TrackProto: glinet.MultiWANTrackProtoBoth,
				TrackMethod: 1, TrackIPv4: ipv4,
			},
			want: `{"interface":"tethering","enable_check":true,"track_proto":2,"track_method":1,"track_mode":2,"track_ipv4":["192.0.2.10","192.0.2.11"],"track_ipv6":[],"enable_ssl":false}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, srv := newClient(t)
			ctx := context.Background()
			sent := recordMultiWANWrites(srv)
			before := srv.MultiWAN.Config()

			if err := c.MultiWAN.SetInterface(ctx, tt.tracking); err != nil {
				t.Fatalf("SetInterface: %v", err)
			}
			if want := []string{"set_interface " + tt.want}; !slices.Equal(*sent, want) {
				t.Errorf("params = %v, want %v", *sent, want)
			}

			got, err := c.MultiWAN.GetConfig(ctx)
			if err != nil {
				t.Fatalf("GetConfig: %v", err)
			}
			var iface glinet.MultiWANInterface
			for _, i := range got.Interfaces {
				if i.Interface == tt.tracking.Interface {
					iface = i
				}
			}
			var old glinettest.MultiWANInterface
			for _, i := range before.Interfaces {
				if i.Interface == tt.tracking.Interface {
					old = i
				}
			}
			if iface.EnableCheck != tt.tracking.EnableCheck || iface.Metric != old.Metric {
				t.Errorf("interface = %+v", iface)
			}
			if tt.tracking.EnableCheck && (iface.TrackProto != tt.tracking.TrackProto || iface.TrackMode != tt.tracking.TrackMode || iface.TrackMethod != tt.tracking.TrackMethod) {
				t.Errorf("interface = %+v, want the track settings of %+v", iface, tt.tracking)
			}
			if !tt.tracking.EnableCheck && iface.TrackMode != old.TrackMode {
				t.Errorf("interface = %+v, want the old track mode %d", iface, old.TrackMode)
			}
		})
	}
}

func TestMultiWANSensitivity(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := recordMultiWANWrites(srv)

	got, err := c.MultiWAN.GetSensitivity(ctx)
	if err != nil {
		t.Fatalf("GetSensitivity: %v", err)
	}
	if *got != (glinet.MultiWANSensitivity{Level: glinet.MultiWANSensitivityMedium, Interval: 5}) {
		t.Errorf("GetSensitivity = %+v", *got)
	}

	if err := c.MultiWAN.SetSensitivity(ctx, glinet.MultiWANSensitivity{Level: glinet.MultiWANSensitivityCustom, Interval: 0.5}); err != nil {
		t.Fatalf("SetSensitivity: %v", err)
	}
	if err := c.MultiWAN.SetSensitivity(ctx, glinet.MultiWANSensitivity{Level: glinet.MultiWANSensitivityHigh, Interval: 30}); err != nil {
		t.Fatalf("SetSensitivity: %v", err)
	}
	want := []string{
		`set_sensitivity {"sensitivity":{"level":"custom","val":0.5}}`,
		`set_sensitivity {"sensitivity":{"level":"high","val":30}}`,
	}
	if !slices.Equal(*sent, want) {
		t.Errorf("params = %v, want %v", *sent, want)
	}

	got, err = c.MultiWAN.GetSensitivity(ctx)
	if err != nil {
		t.Fatalf("GetSensitivity: %v", err)
	}
	if *got != (glinet.MultiWANSensitivity{Level: glinet.MultiWANSensitivityHigh, Interval: 30}) {
		t.Errorf("GetSensitivity = %+v", *got)
	}

	srv.Handle("kmwan", "get_sensitivity", func(json.RawMessage) (any, error) { return map[string]any{}, nil })
	got, err = c.MultiWAN.GetSensitivity(ctx)
	if err != nil {
		t.Fatalf("GetSensitivity: %v", err)
	}
	if *got != (glinet.MultiWANSensitivity{}) {
		t.Errorf("GetSensitivity without sensitivity = %+v, want the zero value", *got)
	}
}

func TestMultiWANStatus(t *testing.T) {
	c, _ := newClient(t)

	got, err := c.MultiWAN.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	want := []glinet.MultiWANInterfaceStatus{
		{Interface: "wan", StatusV4: glinet.MultiWANStatusOffline, StatusV6: glinet.MultiWANStatusOffline},
		{Interface: "wwan", StatusV4: glinet.MultiWANStatusOnline, StatusV6: glinet.MultiWANStatusOffline},
		{Interface: "tethering", StatusV4: glinet.MultiWANStatusOffline, StatusV6: glinet.MultiWANStatusOffline},
	}
	if !slices.Equal(got.Interfaces, want) {
		t.Errorf("GetStatus = %+v, want %+v", got.Interfaces, want)
	}
}

func TestMultiWANDecodeRouterFixtures(t *testing.T) {
	tests := []struct {
		method string
		read   func(context.Context, *glinet.Client) (any, error)
	}{
		{"get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.MultiWAN.GetConfig(ctx) }},
		{"get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.MultiWAN.GetStatus(ctx) }},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			c, srv := newClient(t)
			result := fixtureResult(t, "kmwan."+tt.method+".json")
			srv.Handle("kmwan", tt.method, func(json.RawMessage) (any, error) { return result, nil })

			got, err := tt.read(context.Background(), c)
			if err != nil {
				t.Fatal(err)
			}
			checkLossless(t, got, result)
		})
	}
}
