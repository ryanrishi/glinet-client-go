package glinet_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestDNSHosts(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	initial, err := c.DNS.GetHosts(ctx)
	if err != nil {
		t.Fatalf("GetHosts: %v", err)
	}
	if initial != srv.DNS.Hosts() {
		t.Errorf("GetHosts = %q, want %q", initial, srv.DNS.Hosts())
	}

	if err := c.DNS.SetHosts(ctx, initial+"192.0.2.1 tf-test.invalid"); err != nil {
		t.Fatalf("SetHosts: %v", err)
	}
	if got, want := srv.DNS.Hosts(), initial+"192.0.2.1 tf-test.invalid\n"; got != want {
		t.Errorf("server state = %q, want %q", got, want)
	}
}

func TestDNSConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.DNS.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if res.Mode != "auto" || !res.OverrideVPN || res.RebindProtection || len(res.Server) != 2 {
		t.Errorf("GetConfig = %+v, want the default configuration", *res)
	}

	if err := c.DNS.SetConfig(ctx, glinet.SetDNSConfigRequest{RebindProtection: true, OverrideVPN: true}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	got := srv.DNS.Config()
	if got.Mode != "auto" || !got.RebindProtection || !got.OverrideVPN || got.ForceDNS {
		t.Errorf("server state = %+v, want mode kept and rebind_protection set", got)
	}

	if err := c.Call(ctx, "dns", "set_config", map[string]bool{"rebind_protection": false}, nil); err != nil {
		t.Fatalf("set_config: %v", err)
	}
	if got := srv.DNS.Config(); got.OverrideVPN {
		t.Errorf("server state = %+v, want override_vpn false after a write without it", got)
	}
}

func TestDNSSetConfigModes(t *testing.T) {
	upstream := []string{"wwan 192.168.100.53", "wg 10.66.0.1"}
	tests := []struct {
		name     string
		req      glinet.SetDNSConfigRequest
		wantSent string
		want     glinet.DNSConfig
	}{
		{
			name:     "manual",
			req:      glinet.SetDNSConfigRequest{Mode: "manual", OverrideVPN: true, Server: []string{"1.1.1.1", "9.9.9.9", "2001:db8::53"}},
			wantSent: `{"mode":"manual","force_dns":false,"rebind_protection":false,"override_vpn":true,"server":["1.1.1.1","9.9.9.9","2001:db8::53"]}`,
			want:     glinet.DNSConfig{Mode: "manual", OverrideVPN: true, Server: []string{"1.1.1.1", "9.9.9.9", "2001:db8::53"}, DoTProvider: "0"},
		},
		{
			name:     "proxy",
			req:      glinet.SetDNSConfigRequest{Mode: "proxy", ForceDNS: true, Server: []string{"192.0.2.10#5353"}},
			wantSent: `{"mode":"proxy","force_dns":true,"rebind_protection":false,"override_vpn":false,"server":["192.0.2.10#5353"]}`,
			want:     glinet.DNSConfig{Mode: "proxy", ForceDNS: true, Server: []string{"192.0.2.10#5353"}, DoTProvider: "0"},
		},
		{
			name:     "secure DoT Cloudflare",
			req:      glinet.SetDNSConfigRequest{Mode: "secure", RebindProtection: true, Proto: "DoT", DoTProvider: "0"},
			wantSent: `{"mode":"secure","force_dns":false,"rebind_protection":true,"override_vpn":false,"proto":"DoT","dot_provider":"0"}`,
			want:     glinet.DNSConfig{Mode: "secure", RebindProtection: true, Server: upstream, Proto: "DoT", DoTProvider: "0"},
		},
		{
			name:     "secure DoT NextDNS",
			req:      glinet.SetDNSConfigRequest{Mode: "secure", Proto: "DoT", DoTProvider: "1", NextDNSID: "abc123"},
			wantSent: `{"mode":"secure","force_dns":false,"rebind_protection":false,"override_vpn":false,"proto":"DoT","dot_provider":"1","nextdns_id":"abc123"}`,
			want:     glinet.DNSConfig{Mode: "secure", Server: upstream, Proto: "DoT", DoTProvider: "1", NextDNSID: "abc123"},
		},
		{
			name:     "secure DoT Control D paid",
			req:      glinet.SetDNSConfigRequest{Mode: "secure", Proto: "DoT", DoTProvider: "2", ControlDID: "abc123"},
			wantSent: `{"mode":"secure","force_dns":false,"rebind_protection":false,"override_vpn":false,"proto":"DoT","dot_provider":"2","controld_id":"abc123"}`,
			want:     glinet.DNSConfig{Mode: "secure", Server: upstream, Proto: "DoT", DoTProvider: "2", ControlDID: "abc123"},
		},
		{
			name:     "secure DoT Control D free",
			req:      glinet.SetDNSConfigRequest{Mode: "secure", Proto: "DoT", DoTProvider: "3", FreeControlDID: "p0"},
			wantSent: `{"mode":"secure","force_dns":false,"rebind_protection":false,"override_vpn":false,"proto":"DoT","dot_provider":"3","free_controld_id":"p0"}`,
			want:     glinet.DNSConfig{Mode: "secure", Server: upstream, Proto: "DoT", DoTProvider: "3", FreeControlDID: "p0"},
		},
		{
			name:     "secure DoH",
			req:      glinet.SetDNSConfigRequest{Mode: "secure", OverrideVPN: true, Proto: "DoH", Server: []string{"example-doh", "example-doh-ipv6"}},
			wantSent: `{"mode":"secure","force_dns":false,"rebind_protection":false,"override_vpn":true,"server":["example-doh","example-doh-ipv6"],"proto":"DoH"}`,
			want:     glinet.DNSConfig{Mode: "secure", OverrideVPN: true, Server: []string{"example-doh", "example-doh-ipv6"}, Proto: "DoH", DoTProvider: "0"},
		},
		{
			name:     "auto",
			req:      glinet.SetDNSConfigRequest{Mode: "auto", OverrideVPN: true},
			wantSent: `{"mode":"auto","force_dns":false,"rebind_protection":false,"override_vpn":true}`,
			want:     glinet.DNSConfig{Mode: "auto", OverrideVPN: true, Server: upstream, DoTProvider: "0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, srv := newClient(t)
			ctx := context.Background()
			sent := sentParams(srv, "dns", "set_config")

			if err := c.DNS.SetConfig(ctx, tt.req); err != nil {
				t.Fatalf("SetConfig: %v", err)
			}
			var want map[string]any
			if err := json.Unmarshal([]byte(tt.wantSent), &want); err != nil {
				t.Fatal(err)
			}
			if got := (*sent)[0]; !reflect.DeepEqual(got, want) {
				t.Errorf("params = %v, want %v", got, want)
			}

			got, err := c.DNS.GetConfig(ctx)
			if err != nil {
				t.Fatalf("GetConfig: %v", err)
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("GetConfig = %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func TestDNSSetConfigAutoAfterManual(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	initial, err := c.DNS.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if err := c.DNS.SetConfig(ctx, glinet.SetDNSConfigRequest{Mode: "manual", Server: []string{"1.1.1.1"}}); err != nil {
		t.Fatalf("SetConfig manual: %v", err)
	}
	if err := c.DNS.SetConfig(ctx, glinet.SetDNSConfigRequest{Mode: "auto"}); err != nil {
		t.Fatalf("SetConfig auto: %v", err)
	}
	got, err := c.DNS.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Mode != "auto" || !reflect.DeepEqual(got.Server, initial.Server) {
		t.Errorf("GetConfig = %+v, want mode auto and the upstream servers %v", *got, initial.Server)
	}
}

func TestDNSSetConfigInvalidType(t *testing.T) {
	c, _ := newClient(t)

	err := c.Call(context.Background(), "dns", "set_config", map[string]any{"mode": "manual", "server": "1.1.1.1"}, nil)
	var rpcErr *glinet.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != glinet.CodeInvalidParams {
		t.Errorf("set_config with a string server: error %v, want %d", err, glinet.CodeInvalidParams)
	}
}

func TestDNSGetInfo(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	info, err := c.DNS.GetInfo(ctx)
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if len(info.Protos) != 4 || info.DNSCryptVersion != 2 || len(info.ServerList) == 0 {
		t.Errorf("GetInfo = %+v", *info)
	}
	want := glinet.EncryptedServer{Name: "example-doh", Proto: "DoH", DNSSEC: true, NoFilter: true, NoLog: true}
	if info.ServerList[0] != want {
		t.Errorf("ServerList[0] = %+v, want %+v", info.ServerList[0], want)
	}

	result := fixtureResult(t, "dns.get_info.json")
	srv.Handle("dns", "get_info", func(json.RawMessage) (any, error) { return result, nil })
	info, err = c.DNS.GetInfo(ctx)
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	checkLossless(t, info, result)
}
