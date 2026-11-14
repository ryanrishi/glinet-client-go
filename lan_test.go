package glinet_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestLANStaticBinds(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	binds, err := c.LAN.ListStaticBinds(ctx)
	if err != nil {
		t.Fatalf("ListStaticBinds: %v", err)
	}
	if len(binds) != 0 {
		t.Fatalf("ListStaticBinds = %+v, want empty", binds)
	}

	if err := c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.50", Name: "printer"}); err != nil {
		t.Fatalf("AddStaticBind: %v", err)
	}
	if err := c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.51"}); err != nil {
		t.Fatalf("AddStaticBind without name: %v", err)
	}
	want := []glinettest.StaticBind{
		{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.50", Name: "printer"},
		{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.51"},
	}
	if got := srv.LAN.StaticBinds(); !slices.Equal(got, want) {
		t.Errorf("server state = %+v, want %+v", got, want)
	}

	if err := c.LAN.SetStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.60", Name: "printer-2"}); err != nil {
		t.Fatalf("SetStaticBind: %v", err)
	}
	if err := c.LAN.SetStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.61"}); err != nil {
		t.Fatalf("SetStaticBind without name: %v", err)
	}

	binds, err = c.LAN.ListStaticBinds(ctx)
	if err != nil {
		t.Fatalf("ListStaticBinds: %v", err)
	}
	wantBinds := []glinet.StaticBind{
		{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.60", Name: "printer-2"},
		{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.61"},
	}
	if !slices.Equal(binds, wantBinds) {
		t.Errorf("ListStaticBinds = %+v, want %+v", binds, wantBinds)
	}

	if err := c.LAN.RemoveStaticBind(ctx, "02:00:00:aa:bb:01"); err != nil {
		t.Fatalf("RemoveStaticBind: %v", err)
	}
	want = []glinettest.StaticBind{{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.61"}}
	if got := srv.LAN.StaticBinds(); !slices.Equal(got, want) {
		t.Errorf("server state after remove = %+v, want %+v", got, want)
	}
}

func TestLANAddStaticBindExistingMAC(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.LAN.SetStaticBinds([]glinettest.StaticBind{{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.50", Name: "old"}})

	if err := c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.52"}); err != nil {
		t.Fatalf("AddStaticBind: %v", err)
	}

	want := []glinettest.StaticBind{{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.52"}}
	if got := srv.LAN.StaticBinds(); !slices.Equal(got, want) {
		t.Errorf("server state = %+v, want %+v", got, want)
	}
}

func TestLANStaticBindExistingIP(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.LAN.SetStaticBinds([]glinettest.StaticBind{
		{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.50"},
		{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.51"},
		{MAC: "02:00:00:aa:bb:03", IP: "192.168.8.52"},
	})

	if err := c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:04", IP: "192.168.8.50"}); err != nil {
		t.Fatalf("AddStaticBind: %v", err)
	}
	if err := c.LAN.SetStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.52"}); err != nil {
		t.Fatalf("SetStaticBind: %v", err)
	}

	want := []glinettest.StaticBind{
		{MAC: "02:00:00:aa:bb:04", IP: "192.168.8.50"},
		{MAC: "02:00:00:aa:bb:02", IP: "192.168.8.52"},
	}
	if got := srv.LAN.StaticBinds(); !slices.Equal(got, want) {
		t.Errorf("server state = %+v, want %+v", got, want)
	}
}

func TestLANUnknownMAC(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	seed := []glinettest.StaticBind{{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.50", Name: "kept"}}
	srv.LAN.SetStaticBinds(seed)

	if err := c.LAN.SetStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:09", IP: "192.168.8.59", Name: "x"}); err != nil {
		t.Errorf("SetStaticBind unknown MAC: %v", err)
	}
	if err := c.LAN.RemoveStaticBind(ctx, "02:00:00:aa:bb:09"); err != nil {
		t.Errorf("RemoveStaticBind unknown MAC: %v", err)
	}
	if err := c.LAN.RemoveStaticBind(ctx, "02:00:00:AA:BB:01"); err != nil {
		t.Errorf("RemoveStaticBind other letter case: %v", err)
	}
	if got := srv.LAN.StaticBinds(); !slices.Equal(got, seed) {
		t.Errorf("server state = %+v, want %+v", got, seed)
	}
}

func TestLANFindStaticBind(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.LAN.SetStaticBinds([]glinettest.StaticBind{
		{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.50", Name: "a"},
		{MAC: "02:00:00:AA:BB:02", IP: "192.168.8.51"},
	})

	got, err := c.LAN.FindStaticBind(ctx, "02:00:00:AA:BB:02")
	if err != nil {
		t.Fatalf("FindStaticBind: %v", err)
	}
	want := glinet.StaticBind{MAC: "02:00:00:AA:BB:02", IP: "192.168.8.51"}
	if got == nil || *got != want {
		t.Errorf("FindStaticBind = %+v, want %+v", got, want)
	}

	for _, mac := range []string{"02:00:00:aa:bb:02", "02:00:00:AA:BB:01", "02:00:00:aa:bb:09"} {
		got, err = c.LAN.FindStaticBind(ctx, mac)
		if err != nil || got != nil {
			t.Errorf("FindStaticBind(%q) = %+v, %v, want nil, nil", mac, got, err)
		}
	}
}

func TestLANStaticBindErrors(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
		want string
	}{
		{"add invalid mac", func() error {
			return c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02-00-00-aa-bb-01", IP: "192.168.8.50"})
		}, "invalid mac"},
		{"add invalid ip", func() error {
			return c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:01", IP: "192.168.8.256"})
		}, "invalid ip"},
		{"add no ip", func() error {
			return c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:01"})
		}, "invalid ip"},
		{"set invalid mac", func() error {
			return c.LAN.SetStaticBind(ctx, glinet.StaticBind{MAC: "bad", IP: "192.168.8.50"})
		}, "invalid mac"},
		{"set invalid ip", func() error {
			return c.LAN.SetStaticBind(ctx, glinet.StaticBind{MAC: "02:00:00:aa:bb:01", IP: "bad"})
		}, "invalid ip"},
		{"remove invalid mac", func() error {
			return c.LAN.RemoveStaticBind(ctx, "bad")
		}, "invalid mac"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			var rpcErr *glinet.RPCError
			if !errors.As(err, &rpcErr) {
				t.Fatalf("error %v is not an *RPCError", err)
			}
			if rpcErr.Code != glinet.CodeInvalidParams || rpcErr.Err != tt.want {
				t.Errorf("error = %+v, want code %d err %q", *rpcErr, glinet.CodeInvalidParams, tt.want)
			}
		})
	}

	if got := srv.LAN.StaticBinds(); len(got) != 0 {
		t.Errorf("server state = %+v, want empty", got)
	}
}

func TestLANWANInfoDecodesRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	wanInfo := fixtureResult(t, "lan.get_wan_info.json")
	srv.LAN.SetWANInfo(wanInfo)

	got, err := c.LAN.GetWANInfo(context.Background())
	if err != nil {
		t.Fatalf("GetWANInfo: %v", err)
	}
	checkLossless(t, map[string]any{"wan_info": got}, wanInfo)
}

func TestLANConfigsDecodeRouterFixture(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	configs, err := c.LAN.ListConfigs(ctx)
	if err != nil {
		t.Fatalf("ListConfigs: %v", err)
	}
	checkLossless(t, map[string]any{"interfaces": configs}, fixtureResult(t, "lan.get_config_list.json"))

	guest, err := c.LAN.GetConfig(ctx, "guest")
	if err != nil {
		t.Fatalf("GetConfig guest: %v", err)
	}
	if guest == nil || guest.IP != "192.168.9.1" || guest.Enable != 0 || guest.APIsolate != 1 || guest.TransferEnable == nil {
		t.Errorf("GetConfig guest = %+v", guest)
	}
	if lan := configs[0]; lan.Interface != "lan" || lan.TransferEnable != nil || lan.LeaseTime != "12h" {
		t.Errorf("lan = %+v, want no transfer_enable and leasetime 12h", lan)
	}

	none, err := c.LAN.GetConfig(ctx, "wan")
	if err != nil || none != nil {
		t.Errorf("GetConfig wan = %+v, %v, want nil, nil", none, err)
	}
}

func TestLANSetAddressAndDHCP(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent map[string]any
	setConfig := srv.Handler("lan", "set_config")
	srv.Handle("lan", "set_config", func(params json.RawMessage) (any, error) {
		sent = nil
		_ = json.Unmarshal(params, &sent)
		return setConfig(params)
	})

	if err := c.LAN.SetAddress(ctx, glinet.LANAddress{Interface: "guest", IP: "192.168.10.1", Netmask: "255.255.255.0"}); err != nil {
		t.Fatalf("SetAddress: %v", err)
	}
	want := map[string]any{"interface": "guest", "ip": "192.168.10.1", "netmask": "255.255.255.0", "ap_isolate": float64(0)}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("SetAddress params = %v, want %v", sent, want)
	}

	dhcp := glinet.LANDHCP{Interface: "guest", Enable: 1, Start: "192.168.10.100", End: "192.168.10.200", LeaseTime: "120m"}
	if err := c.LAN.SetDHCP(ctx, dhcp); err != nil {
		t.Fatalf("SetDHCP: %v", err)
	}
	want = map[string]any{"interface": "guest", "enable": float64(1), "start": "192.168.10.100", "end": "192.168.10.200", "leasetime": "120m", "gateway": "", "dns": []any{}}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("SetDHCP params = %v, want %v", sent, want)
	}

	guest, err := c.LAN.GetConfig(ctx, "guest")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	zero := 0
	wantGuest := glinet.LANConfig{
		Interface: "guest", IP: "192.168.10.1", Netmask: "255.255.255.0", Enable: 1,
		Start: "192.168.10.100", End: "192.168.10.200", LeaseTime: "120m",
		DNS: []string{}, LPR: []string{}, TransferEnable: &zero,
	}
	if !reflect.DeepEqual(*guest, wantGuest) {
		t.Errorf("GetConfig = %+v, want %+v", *guest, wantGuest)
	}

	if err := c.LAN.SetDHCP(ctx, glinet.LANDHCP{Interface: "guest", Start: "192.168.10.2"}); err != nil {
		t.Fatalf("SetDHCP disable: %v", err)
	}
	want = map[string]any{"interface": "guest", "enable": float64(0)}
	if !reflect.DeepEqual(sent, want) {
		t.Errorf("SetDHCP disable params = %v, want %v", sent, want)
	}

	err = c.LAN.SetAddress(ctx, glinet.LANAddress{Interface: "wan", IP: "192.168.11.1", Netmask: "255.255.255.0"})
	if rpcCode(t, err) != glinettest.ErrInvalidParams.Code {
		t.Errorf("SetAddress wan: error = %v, want invalid params", err)
	}
}
