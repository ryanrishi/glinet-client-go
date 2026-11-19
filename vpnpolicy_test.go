package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestVPNPolicyReadsDecodeRouterFixtures(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	mode, err := c.VPNPolicy.GetProxyMode(ctx)
	if err != nil || mode != glinet.VPNProxyGlobal {
		t.Errorf("GetProxyMode = %d, %v, want %d", mode, err, glinet.VPNProxyGlobal)
	}

	global, err := c.VPNPolicy.GetGlobalPolicy(ctx)
	if err != nil {
		t.Fatalf("GetGlobalPolicy: %v", err)
	}
	checkLossless(t, global, fixtureResult(t, "vpn-policy.get_global_policy.json"))

	domain, err := c.VPNPolicy.GetDomainPolicy(ctx)
	if err != nil {
		t.Fatalf("GetDomainPolicy: %v", err)
	}
	checkLossless(t, domain, fixtureResult(t, "vpn-policy.get_domain_policy.json"))

	mac, err := c.VPNPolicy.GetMACPolicy(ctx)
	if err != nil {
		t.Fatalf("GetMACPolicy: %v", err)
	}
	checkLossless(t, mac, fixtureResult(t, "vpn-policy.get_mac_policy.json"))

	vlans, err := c.VPNPolicy.GetVLANPolicy(ctx)
	if err != nil {
		t.Fatalf("GetVLANPolicy: %v", err)
	}
	checkLossless(t, map[string]any{"vlans": vlans}, fixtureResult(t, "vpn-policy.get_vlan_policy.json"))
}

func TestVPNPolicyWrites(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	if err := c.VPNPolicy.SetProxyMode(ctx, glinet.VPNProxyMAC); err != nil {
		t.Fatalf("SetProxyMode: %v", err)
	}
	if mode, err := c.VPNPolicy.GetProxyMode(ctx); err != nil || mode != glinet.VPNProxyMAC {
		t.Errorf("GetProxyMode = %d, %v, want %d", mode, err, glinet.VPNProxyMAC)
	}

	sent := sentParams(srv, "vpn-policy", "set_global_policy")
	if err := c.VPNPolicy.SetClientPolicy(ctx, glinet.VPNGlobalPolicy{KillSwitch: true, ServicePolicy: true}); err != nil {
		t.Fatalf("SetClientPolicy: %v", err)
	}
	if err := c.VPNPolicy.SetServerPolicy(ctx, false); err != nil {
		t.Fatalf("SetServerPolicy: %v", err)
	}
	want := []map[string]any{
		{"kill_switch": float64(1), "wan_access": float64(0), "service_policy": float64(1)},
		{"vpn_server_policy": float64(0)},
	}
	if !reflect.DeepEqual(*sent, want) {
		t.Errorf("set_global_policy params = %v, want %v", *sent, want)
	}
	global, err := c.VPNPolicy.GetGlobalPolicy(ctx)
	if err != nil || *global != (glinet.VPNGlobalPolicy{KillSwitch: true, ServicePolicy: true}) {
		t.Errorf("GetGlobalPolicy = %+v, %v", global, err)
	}

	mac := glinet.VPNMACPolicy{DefaultPolicy: 0, MACs: []string{"02:00:00:00:00:01"}}
	if err := c.VPNPolicy.SetMACPolicy(ctx, mac); err != nil {
		t.Fatalf("SetMACPolicy: %v", err)
	}
	if got, err := c.VPNPolicy.GetMACPolicy(ctx); err != nil || !reflect.DeepEqual(*got, mac) {
		t.Errorf("GetMACPolicy = %+v, %v, want %+v", got, err, mac)
	}
	if err := c.VPNPolicy.SetMACPolicy(ctx, glinet.VPNMACPolicy{DefaultPolicy: 1}); err != nil {
		t.Fatalf("SetMACPolicy empty: %v", err)
	}
	if got, err := c.VPNPolicy.GetMACPolicy(ctx); err != nil || got.MACs == nil || len(got.MACs) != 0 {
		t.Errorf("GetMACPolicy = %+v, %v, want an empty list", got, err)
	}

	vlans := []glinet.VPNVLAN{{ID: 1, VPN: true}, {ID: 2}}
	if err := c.VPNPolicy.SetVLANPolicy(ctx, vlans); err != nil {
		t.Fatalf("SetVLANPolicy: %v", err)
	}
	if got, err := c.VPNPolicy.GetVLANPolicy(ctx); err != nil || !slices.Equal(got, vlans) {
		t.Errorf("GetVLANPolicy = %+v, %v, want %+v", got, err, vlans)
	}
}

func TestVPNPolicyDomains(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentParams(srv, "vpn-policy", "set_domain_policy")

	policy := glinet.VPNDomainPolicy{DefaultPolicy: 0, Manual: true, Domains: []string{"example.com", "192.168.50.0/24"}, URL: "not sent"}
	if err := c.VPNPolicy.SetDomainPolicy(ctx, policy); err != nil {
		t.Fatalf("SetDomainPolicy: %v", err)
	}
	want := map[string]any{"default_policy": float64(0), "manual": true, "domain_list": "example.com\n192.168.50.0/24"}
	if got := (*sent)[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("set_domain_policy params = %v, want %v", got, want)
	}
	got, err := c.VPNPolicy.GetDomainPolicy(ctx)
	if err != nil || !slices.Equal(got.Domains, policy.Domains) {
		t.Errorf("GetDomainPolicy = %+v, %v, want domains %v", got, err, policy.Domains)
	}

	if err := c.VPNPolicy.SetDomainPolicy(ctx, glinet.VPNDomainPolicy{DefaultPolicy: 1, URL: "http://example.com/list.txt", Domains: []string{"not sent"}}); err != nil {
		t.Fatalf("SetDomainPolicy URL: %v", err)
	}
	want = map[string]any{"default_policy": float64(1), "manual": false, "url": "http://example.com/list.txt"}
	if got := (*sent)[1]; !reflect.DeepEqual(got, want) {
		t.Errorf("set_domain_policy params = %v, want %v", got, want)
	}

	srv.VPNPolicy.SetDomainPolicy(json.RawMessage(`{"default_policy":0,"manual":true,"domain_list":["a.example","b.example"]}`))
	got, err = c.VPNPolicy.GetDomainPolicy(ctx)
	if err != nil || !slices.Equal(got.Domains, []string{"a.example", "b.example"}) {
		t.Errorf("GetDomainPolicy with a list = %+v, %v", got, err)
	}
}
