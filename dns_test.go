package glinet_test

import (
	"context"
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
