package glinet_test

import (
	"encoding/json"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestFlagDecodesNumbersAndBools(t *testing.T) {
	for input, want := range map[string]bool{
		`{"nat_enable":1,"sip_enable":0}`:        true,
		`{"nat_enable":true,"sip_enable":false}`: true,
		`{"nat_enable":0,"sip_enable":null}`:     false,
		`{"nat_enable":false}`:                   false,
		`{}`:                                     false,
	} {
		var got glinet.NetworkAdvancedConfig
		if err := json.Unmarshal([]byte(input), &got); err != nil || got.FullConeNATEnabled != want || got.SIPALGEnabled {
			t.Errorf("Unmarshal(%s) = %+v, %v, want FullConeNATEnabled %v", input, got, err, want)
		}
	}

	for _, input := range []string{`{"nat_enable":2}`, `{"nat_enable":"1"}`} {
		var got glinet.NetworkAdvancedConfig
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("Unmarshal(%s) = %+v, want an error", input, got)
		}
	}
}

func TestFlagEncodesNumbers(t *testing.T) {
	cases := map[string]any{
		`{"nat_enable":1,"sip_enable":0}`:                                           glinet.NetworkAdvancedConfig{FullConeNATEnabled: true},
		`{"kill_switch":0,"wan_access":1,"service_policy":0,"vpn_server_policy":1}`: glinet.VPNGlobalPolicy{WANAccess: true, VPNServerPolicy: true},
		`{"id":2,"vpn":1}`: glinet.VPNVLAN{ID: 2, VPN: true},
	}
	for want, value := range cases {
		got, err := json.Marshal(value)
		if err != nil || string(got) != want {
			t.Errorf("Marshal(%+v) = %s, %v, want %s", value, got, err, want)
		}
	}
}

func TestLANConfigFlags(t *testing.T) {
	var got glinet.LANConfig
	if err := json.Unmarshal([]byte(`{"interface":"guest","ap_isolate":1,"wan_isolate":true,"enable":0}`), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Interface != "guest" || !got.APIsolate || !got.WANIsolate || got.DHCPEnabled {
		t.Errorf("Unmarshal = %+v", got)
	}

	data, err := json.Marshal(glinet.LANConfig{Interface: "lan", DHCPEnabled: true})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if m["ap_isolate"] != float64(0) || m["enable"] != float64(1) || m["interface"] != "lan" {
		t.Errorf("Marshal = %s", data)
	}
	if _, ok := m["wan_isolate"]; ok {
		t.Errorf("Marshal = %s, want no wan_isolate", data)
	}
}
