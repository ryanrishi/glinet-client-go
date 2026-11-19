package glinettest

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestWANAccessConcurrentPartialWrites(t *testing.T) {
	srv := NewServer(t)
	checkConcurrentPartialWrites(t, srv.Handler("firewall", "set_wan_access"), func() any { return srv.Firewall.WANAccess() }, map[string][]string{
		"enable_https":     toggles(500),
		"enable_ssh":       toggles(500),
		"enable_ping":      toggles(500),
		"enable_whitelist": toggles(500),
	})
}

func TestWANAccessInvalidWriteKeepsWhitelist(t *testing.T) {
	srv := NewServer(t)
	want := WANAccess{Whitelist: []WANAccessSource{{IPAddr: "203.0.113.10", Name: "example"}}}
	srv.Firewall.SetWANAccess(want)

	_, err := srv.Handler("firewall", "set_wan_access")(json.RawMessage(`{"whitelist":[{"ipaddr":"198.51.100.20","name":"changed"}],"enable_ssh":"wrong type"}`))
	if !errors.Is(err, ErrInvalidParams) {
		t.Fatalf("set_wan_access error = %v, want ErrInvalidParams", err)
	}
	if got := srv.Firewall.WANAccess(); !reflect.DeepEqual(got, want) {
		t.Errorf("WANAccess = %+v, want %+v", got, want)
	}
}
