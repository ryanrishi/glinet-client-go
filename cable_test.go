package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

// TestCableStatusNestedErrCode checks that the err_code of secondwan in the get_status result of
// the test router is not an error, while a top-level err_code is.
func TestCableStatusNestedErrCode(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	status := fixtureResult(t, "cable.get_status.json")
	srv.Cable.SetStatus(status)

	got, err := c.Cable.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if got.Status != 3 || got.SecondWAN.ErrCode != -4 || got.SecondWAN.ErrMsg != "Unsupported features!" {
		t.Errorf("GetStatus = %+v", *got)
	}
	checkLossless(t, got, status)

	srv.Cable.SetStatus(json.RawMessage(`{"err_code": -4, "err_msg": "Unsupported features!"}`))
	if _, err := c.Cable.GetStatus(ctx); rpcCode(t, err) != -4 {
		t.Errorf("GetStatus with a top-level err_code: error = %v, want code -4", err)
	}
}

func TestCableConfigDecodesRouterFixture(t *testing.T) {
	c, _ := newClient(t)

	got, err := c.Cable.GetConfig(context.Background(), glinet.FirstWANPort)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Protocol != "dhcp" {
		t.Errorf("Protocol = %q, want dhcp", got.Protocol)
	}
	checkLossless(t, got, fixtureResult(t, "cable.get_config.json"))
}

func TestCableSetConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent map[string]any
	setConfig := srv.Handler("cable", "set_config")
	srv.Handle("cable", "set_config", func(params json.RawMessage) (any, error) {
		sent = nil
		_ = json.Unmarshal(params, &sent)
		return setConfig(params)
	})

	static := glinet.CableConfig{
		Protocol: "static",
		IPv4:     &glinet.CableIPv4{IP: "192.168.100.30", Netmask: "255.255.255.0", Gateway: "192.168.100.1", DNS: []string{"192.168.100.1"}},
		MTU:      1500,
	}
	if err := c.Cable.SetConfig(ctx, glinet.FirstWANPort, static); err != nil {
		t.Fatalf("SetConfig static: %v", err)
	}
	if sent["secondwan"] != false {
		t.Errorf("params = %v, want secondwan false", sent)
	}
	got, err := c.Cable.GetConfig(ctx, glinet.FirstWANPort)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !reflect.DeepEqual(*got, static) {
		t.Errorf("GetConfig = %+v, want %+v", *got, static)
	}

	pppoe := glinet.CableConfig{Protocol: "pppoe", Username: "user", Password: "secret", VLANID: 7}
	if err := c.Cable.SetConfig(ctx, glinet.FirstWANPort, pppoe); err != nil {
		t.Fatalf("SetConfig pppoe: %v", err)
	}
	want := map[string]any{"protocol": "pppoe", "username": "user", "password": "secret", "vlanid": float64(7), "secondwan": false}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
	got, err = c.Cable.GetConfig(ctx, glinet.FirstWANPort)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !reflect.DeepEqual(*got, pppoe) {
		t.Errorf("GetConfig = %+v, want %+v", *got, pppoe)
	}

	if err := c.Cable.SetConfig(ctx, glinet.FirstWANPort, glinet.CableConfig{Protocol: "pppoe"}); err != nil {
		t.Fatalf("SetConfig pppoe with no account: %v", err)
	}
	want = map[string]any{"protocol": "pppoe", "username": "", "password": "", "secondwan": false}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}

	if err := c.Cable.SetConfig(ctx, glinet.FirstWANPort, glinet.CableConfig{Protocol: "dhcp", Username: "user", Password: "secret"}); err != nil {
		t.Fatalf("SetConfig dhcp: %v", err)
	}
	want = map[string]any{"protocol": "dhcp", "secondwan": false}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
}

func TestCableSecondWANPort(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []map[string]any
	for _, method := range []string{"get_config", "set_config"} {
		handler := srv.Handler("cable", method)
		srv.Handle("cable", method, func(params json.RawMessage) (any, error) {
			var p map[string]any
			_ = json.Unmarshal(params, &p)
			sent = append(sent, p)
			return handler(params)
		})
	}

	static := glinet.CableConfig{
		Protocol: "static",
		IPv4:     &glinet.CableIPv4{IP: "198.51.100.30", Netmask: "255.255.255.0", Gateway: "198.51.100.1", DNS: []string{"198.51.100.1"}},
	}
	if err := c.Cable.SetConfig(ctx, glinet.SecondWANPort, static); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	second, err := c.Cable.GetConfig(ctx, glinet.SecondWANPort)
	if err != nil {
		t.Fatalf("GetConfig second: %v", err)
	}
	if !reflect.DeepEqual(*second, static) {
		t.Errorf("GetConfig second = %+v, want %+v", *second, static)
	}
	first, err := c.Cable.GetConfig(ctx, glinet.FirstWANPort)
	if err != nil {
		t.Fatalf("GetConfig first: %v", err)
	}
	if first.Protocol != "dhcp" || first.IPv4 != nil {
		t.Errorf("GetConfig first = %+v, want the fixture", *first)
	}

	wantSecondWAN := []any{true, true, false}
	for i, p := range sent {
		if p["secondwan"] != wantSecondWAN[i] {
			t.Errorf("call %d: params = %v, want secondwan %v", i, p, wantSecondWAN[i])
		}
	}
}

func TestCableSecondWANStatus(t *testing.T) {
	c, srv := newClient(t)
	srv.Cable.SetStatus(json.RawMessage(`{"mode": 0, "status": 3, "secondwan": {"mode": 0, "status": 1}}`))

	got, err := c.Cable.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if got.SecondWAN.Status != 1 || got.SecondWAN.ErrCode != 0 {
		t.Errorf("GetStatus = %+v, want second WAN status 1", *got)
	}
}

func TestCablePorts(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var params []string
	handler := srv.Handler("cable", "set_port_config")
	srv.Handle("cable", "set_port_config", func(p json.RawMessage) (any, error) {
		params = append(params, string(p))
		return handler(p)
	})

	ports, err := c.Cable.ListPorts(ctx)
	if err != nil {
		t.Fatalf("ListPorts: %v", err)
	}
	checkLossless(t, map[string]any{"ports": ports}, srv.Cable.Ports())

	writes := []struct {
		mode string
		mac  glinet.CablePortMAC
		want glinet.CablePortMAC
	}{
		{"wan", glinet.CablePortMAC{Mode: "clone", MACAddr: "02:00:00:00:00:a1"}, glinet.CablePortMAC{Mode: "clone", MACAddr: "02:00:00:00:00:a1", DefaultMACAddr: "02:00:00:00:00:02"}},
		{"wan", glinet.CablePortMAC{Mode: "random", MACAddr: "02:00:00:00:00:a2", Update: "none", Period: 24}, glinet.CablePortMAC{Mode: "random", MACAddr: "02:00:00:00:00:a2", DefaultMACAddr: "02:00:00:00:00:02", Update: "none"}},
		{"wan", glinet.CablePortMAC{Mode: "random", MACAddr: "02:00:00:00:00:a3", Update: "reboot", Period: 168}, glinet.CablePortMAC{Mode: "random", MACAddr: "02:00:00:00:00:a3", DefaultMACAddr: "02:00:00:00:00:02", Update: "reboot", Period: 168}},
		{"wan", glinet.CablePortMAC{Mode: "default", MACAddr: "02:00:00:00:00:02", DefaultMACAddr: "02:00:00:00:00:ff"}, glinet.CablePortMAC{Mode: "default", MACAddr: "02:00:00:00:00:02", DefaultMACAddr: "02:00:00:00:00:02"}},
		{"lan", glinet.CablePortMAC{Mode: "clone", MACAddr: "02:00:00:00:00:a4"}, glinet.CablePortMAC{Mode: "default", MACAddr: "02:00:00:00:00:02", DefaultMACAddr: "02:00:00:00:00:02"}},
	}
	for _, w := range writes {
		if err := c.Cable.SetPort(ctx, "wan", w.mode, w.mac); err != nil {
			t.Fatalf("SetPort(%s, %+v): %v", w.mode, w.mac, err)
		}
		ports, err := c.Cable.ListPorts(ctx)
		if err != nil {
			t.Fatalf("ListPorts: %v", err)
		}
		if len(ports) != 1 || ports[0].Mode != w.mode || ports[0].MAC == nil || *ports[0].MAC != w.want {
			t.Errorf("ListPorts after SetPort(%s, %+v) = %+v", w.mode, w.mac, ports)
		}
	}

	want := []string{
		`{"macaddr":{"macaddr":"02:00:00:00:00:a1","mode":"clone"},"mode":"wan","name":"wan"}`,
		`{"macaddr":{"macaddr":"02:00:00:00:00:a2","mode":"random","update":"none"},"mode":"wan","name":"wan"}`,
		`{"macaddr":{"macaddr":"02:00:00:00:00:a3","mode":"random","period":168,"update":"reboot"},"mode":"wan","name":"wan"}`,
		`{"macaddr":{"macaddr":"02:00:00:00:00:02","mode":"default"},"mode":"wan","name":"wan"}`,
		`{"mode":"lan","name":"wan"}`,
	}
	if !slices.Equal(params, want) {
		t.Errorf("params = %v, want %v", params, want)
	}
}
