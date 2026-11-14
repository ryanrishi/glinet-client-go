package glinet_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestFirewallPortForwards(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	forwards, err := c.Firewall.ListPortForwards(ctx)
	if err != nil {
		t.Fatalf("ListPortForwards: %v", err)
	}
	if len(forwards) != 0 {
		t.Fatalf("ListPortForwards = %+v, want empty", forwards)
	}

	web := glinet.PortForward{Name: "web", Proto: "tcp", Src: "wan", SrcDport: "8080", Dest: "lan", DestIP: "192.168.8.50", DestPort: "80", Enabled: true}
	game := glinet.PortForward{Name: "game", Proto: "tcp udp", Src: "wan", SrcDport: "50004-50005", Dest: "lan", DestIP: "192.168.8.51", DestPort: "50004-50005"}

	webID, err := c.Firewall.AddPortForward(ctx, web)
	if err != nil {
		t.Fatalf("AddPortForward: %v", err)
	}
	gameID, err := c.Firewall.AddPortForward(ctx, game)
	if err != nil {
		t.Fatalf("AddPortForward: %v", err)
	}
	if webID != glinettest.PortForwardID(0) || gameID != glinettest.PortForwardID(1) {
		t.Errorf("IDs = %q, %q, want %q, %q", webID, gameID, glinettest.PortForwardID(0), glinettest.PortForwardID(1))
	}

	wantState := []glinettest.PortForward{
		{Name: "web", Proto: "tcp", Src: "wan", SrcDport: "8080", Dest: "lan", DestIP: "192.168.8.50", DestPort: "80", Enabled: true},
		{Name: "game", Proto: "tcp udp", Src: "wan", SrcDport: "50004-50005", Dest: "lan", DestIP: "192.168.8.51", DestPort: "50004-50005"},
	}
	if got := srv.Firewall.PortForwards(); !slices.Equal(got, wantState) {
		t.Errorf("server state = %+v, want %+v", got, wantState)
	}

	web.ID = webID
	web.SrcDport = "8081"
	web.Enabled = false
	if err := c.Firewall.SetPortForward(ctx, web); err != nil {
		t.Fatalf("SetPortForward: %v", err)
	}

	got, err := c.Firewall.FindPortForward(ctx, webID)
	if err != nil {
		t.Fatalf("FindPortForward: %v", err)
	}
	if got == nil || *got != web {
		t.Errorf("FindPortForward = %+v, want %+v", got, web)
	}

	if err := c.Firewall.RemovePortForward(ctx, webID); err != nil {
		t.Fatalf("RemovePortForward: %v", err)
	}

	forwards, err = c.Firewall.ListPortForwards(ctx)
	if err != nil {
		t.Fatalf("ListPortForwards: %v", err)
	}
	game.ID = webID
	if want := []glinet.PortForward{game}; !slices.Equal(forwards, want) {
		t.Errorf("ListPortForwards after remove = %+v, want %+v (the ID of the next forward changes)", forwards, want)
	}

	got, err = c.Firewall.FindPortForward(ctx, gameID)
	if err != nil || got != nil {
		t.Errorf("FindPortForward unknown ID = %+v, %v, want nil, nil", got, err)
	}
}

func TestFirewallAddPortForwardDuplicate(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	f := glinet.PortForward{Name: "web", Proto: "tcp", Src: "wan", SrcDport: "8080", Dest: "lan", DestIP: "192.168.8.50", DestPort: "80", Enabled: true}

	for range 2 {
		if _, err := c.Firewall.AddPortForward(ctx, f); err != nil {
			t.Fatalf("AddPortForward: %v", err)
		}
	}

	if got := srv.Firewall.PortForwards(); len(got) != 2 || got[0] != got[1] {
		t.Errorf("server state = %+v, want two equal forwards", got)
	}
}

func TestFirewallSetPortForwardEmptyValues(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.Firewall.SetPortForwards([]glinettest.PortForward{
		{Name: "web", Proto: "tcp", Src: "wan", SrcDport: "8080", Dest: "lan", DestIP: "192.168.8.50", DestPort: "80", Enabled: true},
	})

	err := c.Firewall.SetPortForward(ctx, glinet.PortForward{ID: glinettest.PortForwardID(0), Src: "wan", SrcDport: "8080", Dest: "lan", DestIP: "192.168.8.50", Enabled: true})
	if err != nil {
		t.Fatalf("SetPortForward: %v", err)
	}

	want := []glinettest.PortForward{{Src: "wan", SrcDport: "8080", Dest: "lan", DestIP: "192.168.8.50", Enabled: true}}
	if got := srv.Firewall.PortForwards(); !slices.Equal(got, want) {
		t.Errorf("server state = %+v, want %+v", got, want)
	}
}

func TestFirewallPortForwardErrors(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	seed := []glinettest.PortForward{
		{Name: "no-dest-port", Src: "wan", SrcDport: "8080", Dest: "lan", DestIP: "192.168.8.50", Enabled: true},
	}
	srv.Firewall.SetPortForwards(seed)

	if err := c.Firewall.SetPortForward(ctx, glinet.PortForward{ID: glinettest.PortForwardID(5), Name: "x"}); err != nil {
		t.Errorf("SetPortForward unknown ID: %v", err)
	}

	err := c.Firewall.SetPortForward(ctx, glinet.PortForward{Name: "x"})
	var rpcErr *glinet.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != glinet.CodeInvalidParams || rpcErr.Err != "id: string required" {
		t.Errorf("SetPortForward without ID: error %v, want %d id: string required", err, glinet.CodeInvalidParams)
	}

	for _, id := range []string{glinettest.PortForwardID(5), glinettest.PortForwardID(0)} {
		err := c.Firewall.RemovePortForward(ctx, id)
		if err == nil || errors.As(err, &rpcErr) {
			t.Errorf("RemovePortForward(%q): error %v, want an HTTP status error", id, err)
		}
	}

	if got := srv.Firewall.PortForwards(); !slices.Equal(got, seed) {
		t.Errorf("server state = %+v, want %+v", got, seed)
	}
}
