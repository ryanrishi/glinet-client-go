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

func TestFirewallRules(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	rules, err := c.Firewall.ListRules(ctx)
	if err != nil {
		t.Fatalf("ListRules: %v", err)
	}
	if len(rules) != 0 {
		t.Fatalf("ListRules = %+v, want empty", rules)
	}

	drop := glinet.Rule{Name: "drop", Src: "wan", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50001", Proto: "tcp", Target: "DROP", Enabled: true}
	input := glinet.Rule{Name: "input", Src: "wan", SrcIP: "203.0.113.5", SrcMAC: "02:00:00:aa:bb:01", SrcPort: "50002", DestPort: "50003-50004", Proto: "tcp udp", Target: "ACCEPT"}

	dropID, err := c.Firewall.AddRule(ctx, drop)
	if err != nil {
		t.Fatalf("AddRule: %v", err)
	}
	inputID, err := c.Firewall.AddRule(ctx, input)
	if err != nil {
		t.Fatalf("AddRule: %v", err)
	}
	if dropID != glinettest.RuleID(0) || inputID != glinettest.RuleID(1) {
		t.Errorf("IDs = %q, %q, want %q, %q", dropID, inputID, glinettest.RuleID(0), glinettest.RuleID(1))
	}

	wantState := []glinettest.Rule{
		{Name: "drop", Src: "wan", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50001", Proto: "tcp", Target: "DROP", Enabled: true},
		{Name: "input", Src: "wan", SrcIP: "203.0.113.5", SrcMAC: "02:00:00:aa:bb:01", SrcPort: "50002", DestPort: "50003-50004", Proto: "tcp udp", Target: "ACCEPT"},
	}
	if got := srv.Firewall.Rules(); !slices.Equal(got, wantState) {
		t.Errorf("server state = %+v, want %+v", got, wantState)
	}

	drop.ID = dropID
	drop.DestIP = ""
	drop.Target = "REJECT"
	drop.Enabled = false
	if err := c.Firewall.SetRule(ctx, drop); err != nil {
		t.Fatalf("SetRule: %v", err)
	}

	got, err := c.Firewall.FindRule(ctx, dropID)
	if err != nil {
		t.Fatalf("FindRule: %v", err)
	}
	if got == nil || *got != drop {
		t.Errorf("FindRule = %+v, want %+v", got, drop)
	}

	if err := c.Firewall.RemoveRule(ctx, dropID); err != nil {
		t.Fatalf("RemoveRule: %v", err)
	}

	rules, err = c.Firewall.ListRules(ctx)
	if err != nil {
		t.Fatalf("ListRules: %v", err)
	}
	input.ID = dropID
	if want := []glinet.Rule{input}; !slices.Equal(rules, want) {
		t.Errorf("ListRules after remove = %+v, want %+v (the ID of the next rule changes)", rules, want)
	}

	got, err = c.Firewall.FindRule(ctx, inputID)
	if err != nil || got != nil {
		t.Errorf("FindRule unknown ID = %+v, %v, want nil, nil", got, err)
	}
}

func TestFirewallRulesAndPortForwardsShareIDs(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()
	rule := glinet.Rule{Name: "rule", Src: "wan", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50001", Target: "DROP"}
	forward := glinet.PortForward{Name: "forward", Proto: "tcp", Src: "wan", SrcDport: "50002", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50002"}

	ruleID, err := c.Firewall.AddRule(ctx, rule)
	if err != nil {
		t.Fatalf("AddRule: %v", err)
	}
	forwardID, err := c.Firewall.AddPortForward(ctx, forward)
	if err != nil {
		t.Fatalf("AddPortForward: %v", err)
	}
	if ruleID != glinettest.RuleID(0) || forwardID != glinettest.PortForwardID(1) {
		t.Fatalf("IDs = %q, %q, want %q, %q", ruleID, forwardID, glinettest.RuleID(0), glinettest.PortForwardID(1))
	}

	if err := c.Firewall.RemoveRule(ctx, ruleID); err != nil {
		t.Fatalf("RemoveRule: %v", err)
	}

	forwards, err := c.Firewall.ListPortForwards(ctx)
	if err != nil {
		t.Fatalf("ListPortForwards: %v", err)
	}
	if len(forwards) != 1 || forwards[0].ID != glinettest.PortForwardID(0) {
		t.Errorf("ListPortForwards after rule remove = %+v, want ID %q", forwards, glinettest.PortForwardID(0))
	}
}

func TestFirewallRuleErrors(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	rules := []glinettest.Rule{
		{Name: "no-dest-port", Src: "wan", Dest: "lan", DestIP: "192.168.8.50", Target: "DROP", Enabled: true},
	}
	forwards := []glinettest.PortForward{
		{Name: "forward", Proto: "tcp", Src: "wan", SrcDport: "50002", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50002"},
	}
	srv.Firewall.SetRules(rules)
	srv.Firewall.SetPortForwards(forwards)

	if err := c.Firewall.SetRule(ctx, glinet.Rule{ID: glinettest.RuleID(5), Name: "x"}); err != nil {
		t.Errorf("SetRule unknown ID: %v", err)
	}

	err := c.Firewall.SetRule(ctx, glinet.Rule{Name: "x"})
	var rpcErr *glinet.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != glinet.CodeInvalidParams || rpcErr.Err != "id: string required" {
		t.Errorf("SetRule without ID: error %v, want %d id: string required", err, glinet.CodeInvalidParams)
	}

	for _, id := range []string{glinettest.RuleID(5), glinettest.RuleID(0)} {
		err := c.Firewall.RemoveRule(ctx, id)
		if err == nil || errors.As(err, &rpcErr) {
			t.Errorf("RemoveRule(%q): error %v, want an HTTP status error", id, err)
		}
	}

	if got := srv.Firewall.Rules(); !slices.Equal(got, rules) {
		t.Errorf("server rules = %+v, want %+v", got, rules)
	}
	if got := srv.Firewall.PortForwards(); !slices.Equal(got, forwards) {
		t.Errorf("server port forwards = %+v, want %+v", got, forwards)
	}
}

func TestFirewallRemoveRuleWithPortForwardID(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.Firewall.SetPortForwards([]glinettest.PortForward{
		{Name: "forward", Proto: "tcp", Src: "wan", SrcDport: "50002", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50002"},
	})

	if err := c.Firewall.RemoveRule(ctx, glinettest.PortForwardID(0)); err != nil {
		t.Fatalf("RemoveRule: %v", err)
	}

	if got := srv.Firewall.PortForwards(); len(got) != 0 {
		t.Errorf("server port forwards = %+v, want empty", got)
	}
}
