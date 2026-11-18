package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestWifiStatusDecodesRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	status := fixtureResult(t, "wifi.get_status.json")
	srv.Wifi.SetStatus(status)

	got, err := c.Wifi.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, map[string]any{"res": got}, status)
}

func TestWifiConfigDecodesRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	config := fixtureResult(t, "wifi.get_config.json")
	srv.Wifi.SetConfig(config)

	got, err := c.Wifi.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, got, config)

	radio, iface := got.FindIface("guest2g")
	if radio == nil || radio.Device != "radio1" || iface.Name != "guest2g" || !iface.Guest {
		t.Errorf("FindIface guest2g = %+v, %+v", radio, iface)
	}
	if want := (glinet.WifiHTModes{MaxWidth: map[string]int{"11a/n/ac": 80, "11a/n/ac/ax": 80, "11ac/ax": 80, "11n/ac/ax": 80}}); !reflect.DeepEqual(got.Radios[0].HTModes, want) {
		t.Errorf("radio0 HTModes = %+v, want %+v", got.Radios[0].HTModes, want)
	}
	if radio, iface := got.FindIface("wlan9"); radio != nil || iface != nil {
		t.Errorf("FindIface wlan9 = %+v, %+v, want nil", radio, iface)
	}
}

func TestWifiSetConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent map[string]any
	setConfig := srv.Handler("wifi", "set_config")
	srv.Handle("wifi", "set_config", func(params json.RawMessage) (any, error) {
		sent = nil
		_ = json.Unmarshal(params, &sent)
		return setConfig(params)
	})

	enabled := true
	if err := c.Wifi.SetConfig(ctx, glinet.SetWifiConfigRequest{IfaceName: "guest2g", Enabled: &enabled}); err != nil {
		t.Fatalf("SetConfig enabled: %v", err)
	}
	if want := map[string]any{"iface_name": "guest2g", "enabled": true}; !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}

	hidden, channel, randomBSSID := true, 0, false
	main := glinet.SetWifiConfigRequest{
		IfaceName: "default_radio1", SSID: "test-ssid", Encryption: "sae-mixed", Hidden: &hidden, Key: "test-key-1",
		Device: "radio1", HWMode: "11n/ax", Channel: &channel, HTMode: "20", TxPower: "Low", RandomBSSID: &randomBSSID,
	}
	if err := c.Wifi.SetConfig(ctx, main); err != nil {
		t.Fatalf("SetConfig main: %v", err)
	}
	if sent["channel"] != float64(0) || sent["random_bssid"] != false {
		t.Errorf("params = %v, want channel 0 and random_bssid false", sent)
	}

	got, err := c.Wifi.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	radio, iface := got.FindIface("default_radio1")
	wantIface := glinet.WifiIface{Name: "default_radio1", SSID: "test-ssid", Key: "test-key-1", Encryption: "sae-mixed", Hidden: true}
	if *iface != wantIface {
		t.Errorf("default_radio1 = %+v, want %+v", *iface, wantIface)
	}
	if radio.HWMode != "11n/ax" || radio.Channel != 0 || radio.HTMode != "20" || radio.TxPower != "Low" || radio.RandomBSSID {
		t.Errorf("radio1 = %+v", *radio)
	}
	if _, guest := got.FindIface("guest2g"); !guest.Enabled {
		t.Errorf("guest2g = %+v, want enabled", *guest)
	}
	if radio0, _ := got.FindIface("default_radio0"); radio0.TxPower != "Medium" || radio0.Channel != 48 {
		t.Errorf("radio0 = %+v, want no change", *radio0)
	}

	err = c.Wifi.SetConfig(ctx, glinet.SetWifiConfigRequest{IfaceName: "wlan9", Enabled: &enabled})
	if rpcCode(t, err) != glinettest.ErrInvalidParams.Code {
		t.Errorf("SetConfig wlan9: error = %v, want invalid params", err)
	}
}

func TestWifiSetConfig6GOnlyPSC(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent map[string]any
	setConfig := srv.Handler("wifi", "set_config")
	srv.Handle("wifi", "set_config", func(params json.RawMessage) (any, error) {
		sent = nil
		_ = json.Unmarshal(params, &sent)
		return setConfig(params)
	})

	onlyPSC := true
	if err := c.Wifi.SetConfig(ctx, glinet.SetWifiConfigRequest{IfaceName: "default_radio0", SSID: "example", Encryption: "none", Device: "radio0", Chan6GOnlyPSC: &onlyPSC}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if sent["chan_6g_only_psc"] != true {
		t.Errorf("params = %v, want chan_6g_only_psc true", sent)
	}
	got, err := c.Wifi.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if radio, _ := got.FindIface("default_radio0"); radio.Chan6GOnlyPSC == nil || !*radio.Chan6GOnlyPSC {
		t.Errorf("radio0 = %+v, want Chan6GOnlyPSC true", *radio)
	}
	if radio, _ := got.FindIface("default_radio1"); radio.Chan6GOnlyPSC != nil {
		t.Errorf("radio1 = %+v, want no Chan6GOnlyPSC", *radio)
	}
}

func TestWifiSetTxPower(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	if err := c.Wifi.SetTxPower(ctx, "radio1", "High"); err != nil {
		t.Fatalf("SetTxPower: %v", err)
	}
	got, err := c.Wifi.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Radios[1].TxPower != "High" || got.Radios[0].TxPower != "Medium" {
		t.Errorf("txpower = %q, %q, want Medium, High", got.Radios[0].TxPower, got.Radios[1].TxPower)
	}

	if err := c.Wifi.SetTxPower(ctx, "radio9", "High"); rpcCode(t, err) != glinettest.ErrInvalidParams.Code {
		t.Errorf("SetTxPower radio9: error = %v, want invalid params", err)
	}
}
