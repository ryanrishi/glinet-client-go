package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestRepeaterConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	setConfig := srv.Handler("repeater", "set_config")
	srv.Handle("repeater", "set_config", func(params json.RawMessage) (any, error) {
		sent = append(sent, string(params))
		return setConfig(params)
	})

	got, err := c.Repeater.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	want := glinet.RepeaterConfig{Auto: glinet.RepeaterAutoSwitching, DFSSupport: true, MACAddr: "02:00:00:00:00:38"}
	if *got != want {
		t.Errorf("GetConfig = %+v, want %+v", *got, want)
	}

	if err := c.Repeater.SetConfig(ctx, glinet.SetRepeaterConfigRequest{Auto: glinet.RepeaterAutoSwitchingWithoutNetwork, LockBand: "5g"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	got, err = c.Repeater.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Auto != glinet.RepeaterAutoSwitchingWithoutNetwork || got.LockBand != "5g" {
		t.Errorf("GetConfig = %+v, want auto 2 and lock_band 5g", *got)
	}

	if err := c.Repeater.SetConfig(ctx, glinet.SetRepeaterConfigRequest{Auto: glinet.RepeaterNoSwitching}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	got, err = c.Repeater.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Auto != glinet.RepeaterNoSwitching || got.LockBand != "" {
		t.Errorf("GetConfig = %+v, want auto 0 and no lock_band", *got)
	}

	wantSent := []string{`{"auto":2,"lock_band":"5g"}`, `{"auto":0}`}
	if !slices.Equal(sent, wantSent) {
		t.Errorf("params = %v, want %v", sent, wantSent)
	}
}

func TestRepeaterConfigDecodesBoolAuto(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	for _, tt := range []struct {
		result string
		want   glinet.RepeaterSwitchMode
	}{
		{`{"auto":true}`, glinet.RepeaterAutoSwitching},
		{`{"auto":false}`, glinet.RepeaterNoSwitching},
		{`{"auto":2}`, glinet.RepeaterAutoSwitchingWithoutNetwork},
	} {
		srv.Repeater.SetConfig(json.RawMessage(tt.result))
		got, err := c.Repeater.GetConfig(ctx)
		if err != nil {
			t.Fatalf("GetConfig %s: %v", tt.result, err)
		}
		if got.Auto != tt.want {
			t.Errorf("GetConfig %s: Auto = %d, want %d", tt.result, got.Auto, tt.want)
		}
	}

	srv.Repeater.SetConfig(json.RawMessage(`{"auto":"1"}`))
	if _, err := c.Repeater.GetConfig(ctx); err == nil {
		t.Error("GetConfig with auto \"1\": no error")
	}
}

func TestRepeaterDecodesFirmware470Fixtures(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	srv.Repeater.SetConfig(fixtureResult(t, "repeater.get_config.json"))
	config, err := c.Repeater.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Auto != glinet.RepeaterAutoSwitching {
		t.Errorf("Auto = %d, want %d", config.Auto, glinet.RepeaterAutoSwitching)
	}
	checkLossless(t, config, fixtureResult(t, "repeater.get_config.json"), "auto")

	saved := fixtureResult(t, "repeater.get_saved_ap_list.json")
	srv.Repeater.SetSavedNetworks(saved)
	networks, err := c.Repeater.ListSavedNetworks(ctx)
	if err != nil {
		t.Fatalf("ListSavedNetworks: %v", err)
	}
	checkLossless(t, map[string]any{"res": networks}, saved)

	status := fixtureResult(t, "repeater.get_status.json")
	srv.Repeater.SetStatus(status)
	got, err := c.Repeater.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, got, status)
}

func TestRepeaterConnect(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	connect := srv.Handler("repeater", "connect")
	srv.Handle("repeater", "connect", func(params json.RawMessage) (any, error) {
		sent = append(sent, string(params))
		return connect(params)
	})

	static := glinet.RepeaterNetwork{
		SSID:     "example-2",
		Key:      "example-key-2",
		BSSID:    "02:00:00:00:00:40",
		Protocol: "static",
		IP:       "198.51.100.20",
		Netmask:  "255.255.255.0",
		Gateway:  "198.51.100.1",
		DNS:      []string{"198.51.100.1", "198.51.100.2"},
		MACAddr:  &glinet.RepeaterMAC{Mode: "random", MAC: "02:00:00:00:00:41", Update: "time", Period: 168},
		TTL:      64,
		TTLIPv6:  65,
		MTU:      1400,
		Remember: true,
	}
	if err := c.Repeater.Connect(ctx, static); err != nil {
		t.Fatalf("Connect static: %v", err)
	}
	enterprise := glinet.RepeaterNetwork{
		SSID:       "example-eap",
		Key:        "example-password",
		Identity:   "user",
		Protocol:   "dhcp",
		Disguise:   true,
		AutoPortal: true,
		Remember:   true,
	}
	if err := c.Repeater.Connect(ctx, enterprise); err != nil {
		t.Fatalf("Connect enterprise: %v", err)
	}
	open := glinet.RepeaterNetwork{SSID: "example-open", Protocol: "dhcp", MACAddr: &glinet.RepeaterMAC{Mode: "default"}}
	if err := c.Repeater.Connect(ctx, open); err != nil {
		t.Fatalf("Connect open: %v", err)
	}

	want := []string{
		`{"ssid":"example-2","key":"example-key-2","bssid":"02:00:00:00:00:40","protocol":"static","ip":"198.51.100.20","netmask":"255.255.255.0","gateway":"198.51.100.1","dns":["198.51.100.1","198.51.100.2"],"disguise":false,"macaddr":{"mode":"random","macaddr":"02:00:00:00:00:41","update":"time","period":168},"ttl":64,"ttl_ipv6":65,"mtu":1400,"manual":false,"auto_portal":false,"remember":true}`,
		`{"ssid":"example-eap","key":"example-password","identity":"user","protocol":"dhcp","disguise":true,"manual":false,"auto_portal":true,"remember":true}`,
		`{"ssid":"example-open","protocol":"dhcp","disguise":false,"macaddr":{"mode":"default"},"manual":false,"auto_portal":false,"remember":false}`,
	}
	if !slices.Equal(sent, want) {
		t.Errorf("params =\n%v\nwant\n%v", sent, want)
	}

	status, err := c.Repeater.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.State != 2 || status.SSID != "example-open" || !reflect.DeepEqual(*status.Config, open) {
		t.Errorf("GetStatus = %+v, config %+v", *status, status.Config)
	}

	networks, err := c.Repeater.ListSavedNetworks(ctx)
	if err != nil {
		t.Fatalf("ListSavedNetworks: %v", err)
	}
	var ssids []string
	for _, n := range networks {
		ssids = append(ssids, n.SSID)
	}
	if !slices.Equal(ssids, []string{"example", "example-2", "example-eap"}) {
		t.Errorf("saved networks = %v, want example, example-2, example-eap", ssids)
	}
	savedStatic := static
	savedStatic.Remember = false
	if !reflect.DeepEqual(networks[1], savedStatic) {
		t.Errorf("saved network = %+v, want %+v", networks[1], savedStatic)
	}

	changed := static
	changed.Key = "example-key-3"
	if err := c.Repeater.Connect(ctx, changed); err != nil {
		t.Fatalf("Connect changed: %v", err)
	}
	networks, err = c.Repeater.ListSavedNetworks(ctx)
	if err != nil {
		t.Fatalf("ListSavedNetworks: %v", err)
	}
	if len(networks) != 3 || networks[1].Key != "example-key-3" {
		t.Errorf("saved networks = %+v, want example-2 with the new key", networks)
	}
}

func TestRepeaterDisconnectAndRemove(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	for _, method := range []string{"disconnect", "remove_saved_ap"} {
		handler := srv.Handler("repeater", method)
		srv.Handle("repeater", method, func(params json.RawMessage) (any, error) {
			sent = append(sent, string(params))
			return handler(params)
		})
	}

	if err := c.Repeater.Disconnect(ctx); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	status, err := c.Repeater.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.State != 0 || status.Config != nil {
		t.Errorf("GetStatus = %+v, want state 0 and no config", *status)
	}

	if err := c.Repeater.RemoveSavedNetwork(ctx, "example"); err != nil {
		t.Fatalf("RemoveSavedNetwork: %v", err)
	}
	networks, err := c.Repeater.ListSavedNetworks(ctx)
	if err != nil {
		t.Fatalf("ListSavedNetworks: %v", err)
	}
	if len(networks) != 0 {
		t.Errorf("ListSavedNetworks = %+v, want none", networks)
	}

	if want := []string{`{}`, `{"ssid":"example"}`}; !slices.Equal(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
}
