package glinet_test

import (
	"context"
	"errors"
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

	got, err := c.LAN.FindStaticBind(ctx, "02:00:00:aa:bb:02")
	if err != nil {
		t.Fatalf("FindStaticBind: %v", err)
	}
	want := glinet.StaticBind{MAC: "02:00:00:AA:BB:02", IP: "192.168.8.51"}
	if got == nil || *got != want {
		t.Errorf("FindStaticBind = %+v, want %+v", got, want)
	}

	got, err = c.LAN.FindStaticBind(ctx, "02:00:00:aa:bb:09")
	if err != nil || got != nil {
		t.Errorf("FindStaticBind unknown MAC = %+v, %v, want nil, nil", got, err)
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
