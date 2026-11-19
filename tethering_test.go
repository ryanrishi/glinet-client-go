package glinet_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestTetheringDecodesRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	status := fixtureResult(t, "tethering.get_status.json")
	srv.Tethering.SetStatus(status)

	got, err := c.Tethering.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if got.Status != 2 || got.Devices == nil || len(got.Devices) != 0 || got.IPv4 != nil {
		t.Errorf("GetStatus = %+v, want status 2 and an empty device list", *got)
	}
	checkLossless(t, got, status)
}

func TestTetheringStatusDevices(t *testing.T) {
	c, srv := newClient(t)
	srv.Tethering.SetStatus(json.RawMessage(`{
		"status": 1,
		"devices": [{"device": "usb0", "type": 0, "use": true}, {"device": "eth9", "use": false}],
		"ipv4": {"ip": "192.0.2.20/24", "gateway": "192.0.2.1", "dns": ["192.0.2.1"]}
	}`))

	got, err := c.Tethering.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if len(got.Devices) != 2 || got.Devices[0].Type == nil || *got.Devices[0].Type != 0 || !got.Devices[0].Use || got.Devices[1].Type != nil {
		t.Errorf("Devices = %+v", got.Devices)
	}
	if got.IPv4 == nil || got.IPv4.IP != "192.0.2.20/24" || !slices.Equal(got.IPv4.DNS, []string{"192.0.2.1"}) {
		t.Errorf("IPv4 = %+v", got.IPv4)
	}
}

func TestTetheringConnect(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	for _, method := range []string{"set_connect", "disconnect"} {
		handler := srv.Handler("tethering", method)
		srv.Handle("tethering", method, func(params json.RawMessage) (any, error) {
			sent = append(sent, string(params))
			return handler(params)
		})
	}

	config := glinet.TetheringConfig{Device: "usb0", TTL: 64, TTLIPv6: 65, MTU: 1400}
	if err := c.Tethering.Connect(ctx, config); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	got, err := c.Tethering.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if *got != config {
		t.Errorf("GetConfig = %+v, want %+v", *got, config)
	}
	status, err := c.Tethering.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.Status != 2 {
		t.Errorf("Status = %d, want 2", status.Status)
	}

	if err := c.Tethering.Connect(ctx, glinet.TetheringConfig{Device: "usb0"}); err != nil {
		t.Fatalf("Connect without options: %v", err)
	}
	if err := c.Tethering.Disconnect(ctx); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	status, err = c.Tethering.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.Status != 0 {
		t.Errorf("Status = %d, want 0", status.Status)
	}

	want := []string{
		`{"device":"usb0","ttl":64,"ttl_ipv6":65,"mtu":1400}`,
		`{"device":"usb0"}`,
		`{}`,
	}
	if !slices.Equal(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
}
