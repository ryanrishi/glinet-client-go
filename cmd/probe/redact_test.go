package main

import (
	"encoding/json"
	"testing"
)

func TestRedactAdjacentIPv4(t *testing.T) {
	r := newRedactor()
	got := r.redactString("45.67.89.10,56.78.90.11 192.168.8.1")
	want := "203.0.113.1,203.0.113.2 192.168.8.1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRedactIPv4InsideLongerNumber(t *testing.T) {
	r := newRedactor()
	for _, s := range []string{"1.2.3.4.5", "1234.5.6.7", "4.7.0"} {
		if got := r.redactString(s); got != s {
			t.Errorf("redactString(%q) = %q", s, got)
		}
	}
}

func TestRedactPinCode(t *testing.T) {
	r := newRedactor()
	v := map[string]any{"profiles": []any{map[string]any{"pin_code": "1234", "apn": "internet"}}}
	r.collect(v)
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"profiles":[{"apn":"internet","pin_code":"redacted-pin_code-1"}]}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactPrivateIPv4(t *testing.T) {
	r := newRedactor()
	got := r.redactString("192.168.77.20/24 via 192.168.77.1 dns 172.16.5.9 lan 192.168.8.1 guest 192.168.9.1 ts 100.100.1.2 again 192.168.77.255")
	want := "10.200.1.20/24 via 10.200.1.1 dns 10.200.2.9 lan 192.168.8.1 guest 192.168.9.1 ts 10.200.3.2 again 10.200.1.255"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRedactVPNNames(t *testing.T) {
	r := newRedactor()
	v := map[string]any{
		"groups": []any{
			map[string]any{"group_id": float64(7), "group_name": "NordVPN", "username": "vpn-account"},
			map[string]any{"group_id": float64(8), "group_name": "Office", "peers": []any{
				map[string]any{"peer_id": float64(1), "name": "laptop-tunnel", "listen_port": float64(12345)},
				map[string]any{"peer_id": float64(2), "name": "backup", "listen_port": float64(51820)},
			}},
		},
		"service": []any{map[string]any{"name": "wgclient", "peer_id": float64(1), "status": float64(1)}},
		"log":     "Office: laptop-tunnel up",
	}
	r.collect(v)
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"groups":[{"group_id":7,"group_name":"NordVPN","username":"user-1"},{"group_id":8,"group_name":"group-1","peers":[{"listen_port":40001,"name":"peer-1","peer_id":1},{"listen_port":51820,"name":"peer-2","peer_id":2}]}],"log":"group-1: peer-1 up","service":[{"name":"wgclient","peer_id":1,"status":1}]}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactFirewallRuleName(t *testing.T) {
	r := newRedactor()
	v := map[string]any{"res": []any{
		map[string]any{"id": "cfg01", "name": "game-server", "src": "wan", "dest": "lan"},
		map[string]any{"id": "cfg02", "name": "allow-ssh", "src": "wan", "target": "ACCEPT"},
	}}
	r.collect(v)
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"res":[{"dest":"lan","id":"cfg01","name":"rule-1","src":"wan"},{"id":"cfg02","name":"rule-2","src":"wan","target":"ACCEPT"}]}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactNameThatIsPartOfAKey(t *testing.T) {
	r := newRedactor()
	v := map[string]any{"username": "user", "log": "login of user"}
	r.collect(v)
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"log":"login of user-1","username":"user-1"}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactTimezone(t *testing.T) {
	r := newRedactor()
	v := map[string]any{
		"zonename": "Europe/Berlin", "timezone": "CET-1CEST,M3.5.0,M10.5.0/3", "tzoffset": "+0100",
		"localtime": float64(1003600), "timestamp": float64(1000000), "autotimezone_enabled": true,
	}
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"autotimezone_enabled":true,"localtime":1000000,"timestamp":1000000,"timezone":"UTC0","tzoffset":"+0000","zonename":"UTC"}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactVPNTunnelAndStatus(t *testing.T) {
	r := newRedactor()
	tunnels := []any{
		map[string]any{"tunnel_id": float64(10), "name": "to-office", "to": map[string]any{"domain_list": "intranet.corp.test\nwiki.corp.test"}},
		map[string]any{"tunnel_id": float64(100), "name": "last sort default policy"},
	}
	status := []any{map[string]any{"peer_id": float64(1), "peer_name": "office-peer", "domain": []any{"vpn.corp.test"}}}
	r.collect(tunnels)
	r.collect(status)
	out := string(mustJSON(r.scrub(r.redact(tunnels)))) + string(mustJSON(r.scrub(r.redact(status))))
	want := `[{"name":"tunnel-1","to":{"domain_list":"host-1.example.com\nhost-2.example.com"},"tunnel_id":10},{"name":"last sort default policy","tunnel_id":100}]` +
		`[{"domain":["host-3.example.com"],"peer_id":1,"peer_name":"peer-1"}]`
	if out != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactConnectionUUID(t *testing.T) {
	r := newRedactor()
	v := map[string]any{"wpa_status": map[string]any{"uuid": "6f1c2a7e-3b9d-4e8f-a1c2-9d8e7f6a5b4c", "mode": "station"}}
	r.collect(v)
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"wpa_status":{"mode":"station","uuid":"00000000-0000-4000-8000-000000000001"}}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactRebootSchedule(t *testing.T) {
	r := newRedactor()
	v := map[string]any{"enable": true, "hour": "04", "min": "30", "week": []any{float64(0), float64(6)}}
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"enable":true,"hour":"02","min":"15","week":[3,5]}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestRedactLogTimes(t *testing.T) {
	r := newRedactor()
	v := map[string]any{"log": "Mon Nov 16 21:07:45 2026 usbmuxd: [21:07:45.250] up\n2026/11/16 21:07:46 [error] mac 02:00:00:00:00:01"}
	out := mustJSON(r.scrub(r.redact(v)))
	want := `{"log":"Thu Jan  1 00:00:00 2026 usbmuxd: [00:00:00.250] up\n2026/01/01 00:00:00 [error] mac 02:00:00:00:00:01"}`
	if string(out) != want {
		t.Errorf("got %s, want %s", out, want)
	}
}

func TestScrubKeepsJSONValid(t *testing.T) {
	r := newRedactor()
	r.addSecret("true", "password")
	v := map[string]any{"enabled": true, "note": "is true", "count": float64(3)}
	out := mustJSON(r.scrub(r.redact(v)))

	var decoded map[string]any
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded["enabled"] != true {
		t.Errorf("enabled = %v, want true", decoded["enabled"])
	}
	if decoded["note"] != "is redacted-password-1" {
		t.Errorf("note = %v", decoded["note"])
	}
}
