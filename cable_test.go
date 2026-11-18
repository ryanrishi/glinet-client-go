package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
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

	got, err := c.Cable.GetConfig(context.Background())
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
	if err := c.Cable.SetConfig(ctx, static); err != nil {
		t.Fatalf("SetConfig static: %v", err)
	}
	if sent["secondwan"] != false {
		t.Errorf("params = %v, want secondwan false", sent)
	}
	got, err := c.Cable.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !reflect.DeepEqual(*got, static) {
		t.Errorf("GetConfig = %+v, want %+v", *got, static)
	}

	pppoe := glinet.CableConfig{Protocol: "pppoe", Username: "user", Password: "secret", VLANID: 7}
	if err := c.Cable.SetConfig(ctx, pppoe); err != nil {
		t.Fatalf("SetConfig pppoe: %v", err)
	}
	want := map[string]any{"protocol": "pppoe", "username": "user", "password": "secret", "vlanid": float64(7), "secondwan": false}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
	got, err = c.Cable.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !reflect.DeepEqual(*got, pppoe) {
		t.Errorf("GetConfig = %+v, want %+v", *got, pppoe)
	}

	if err := c.Cable.SetConfig(ctx, glinet.CableConfig{Protocol: "pppoe"}); err != nil {
		t.Fatalf("SetConfig pppoe with no account: %v", err)
	}
	want = map[string]any{"protocol": "pppoe", "username": "", "password": "", "secondwan": false}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}

	if err := c.Cable.SetConfig(ctx, glinet.CableConfig{Protocol: "dhcp", Username: "user", Password: "secret"}); err != nil {
		t.Fatalf("SetConfig dhcp: %v", err)
	}
	want = map[string]any{"protocol": "dhcp", "secondwan": false}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
}
