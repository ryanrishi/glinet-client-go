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

func TestSwitchButton(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	handler := srv.Handler("switch-button", "set_config")
	srv.Handle("switch-button", "set_config", func(params json.RawMessage) (any, error) {
		sent = append(sent, string(params))
		return handler(params)
	})

	tests := []struct {
		config glinet.SwitchButtonConfig
		sync   bool
		params string
	}{
		{glinet.SwitchButtonConfig{Func: "wifi", SubFunc: "guest_wifi"}, false, `{"func":"wifi","sub_func":"guest_wifi","sync":false}`},
		{glinet.SwitchButtonConfig{Func: "vpn", SubFunc: "1"}, true, `{"func":"vpn","sub_func":"1","sync":true}`},
		{glinet.SwitchButtonConfig{Func: "led"}, false, `{"func":"led","sync":false}`},
		{glinet.SwitchButtonConfig{Func: "none"}, false, `{"func":"none","sync":false}`},
	}
	for _, tt := range tests {
		sent = nil
		if err := c.SwitchButton.SetConfig(ctx, tt.config, tt.sync); err != nil {
			t.Fatalf("SetConfig(%+v): %v", tt.config, err)
		}
		if len(sent) != 1 || sent[0] != tt.params {
			t.Errorf("params = %v, want %s", sent, tt.params)
		}
		if got := srv.SwitchButton.Config(); got != (glinettest.SwitchButtonConfig{Func: tt.config.Func, SubFunc: tt.config.SubFunc}) {
			t.Errorf("server state = %+v", got)
		}
		res, err := c.SwitchButton.GetConfig(ctx)
		if err != nil {
			t.Fatalf("GetConfig: %v", err)
		}
		if *res != tt.config {
			t.Errorf("GetConfig = %+v, want %+v", *res, tt.config)
		}
	}

	funcs, err := c.SwitchButton.ListFuncs(ctx)
	if err != nil {
		t.Fatalf("ListFuncs: %v", err)
	}
	wifi := glinet.SwitchButtonFunc{Func: "wifi", SubFuncs: []glinet.SwitchButtonFunc{{Func: "main_wifi"}, {Func: "guest_wifi"}}}
	if len(funcs) != 6 || !reflect.DeepEqual(funcs[4], wifi) {
		t.Errorf("ListFuncs = %+v", funcs)
	}
}

func TestSwitchButtonDecodeFirmware470Funcs(t *testing.T) {
	c, srv := newClient(t)
	srv.SwitchButton.SetFuncs(fixtureResult(t, "switch-button.get_funcs.json"))

	funcs, err := c.SwitchButton.ListFuncs(context.Background())
	if err != nil {
		t.Fatalf("ListFuncs: %v", err)
	}
	var names []string
	for _, f := range funcs {
		if f.SubFuncs != nil {
			t.Errorf("%s has SubFuncs %v", f.Func, f.SubFuncs)
		}
		names = append(names, f.Func)
	}
	if want := []string{"adguardhome", "openvpn", "tor", "wireguard"}; !slices.Equal(names, want) {
		t.Errorf("ListFuncs = %v, want %v", names, want)
	}
}
