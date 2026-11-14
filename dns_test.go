package glinet_test

import (
	"context"
	"testing"
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
