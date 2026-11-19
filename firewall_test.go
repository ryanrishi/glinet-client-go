package glinet_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
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

func TestFirewallRemoveAll(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	forwards := sentParams(srv, "firewall", "remove_port_forward")
	rules := sentParams(srv, "firewall", "remove_rule")
	srv.Firewall.SetPortForwards([]glinettest.PortForward{
		{Name: "a", Proto: "tcp", Src: "wan", SrcDport: "50002", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50002"},
		{Name: "b", Proto: "udp", Src: "wan", SrcDport: "50003", Dest: "lan", DestIP: "192.168.8.51", DestPort: "50003"},
	})
	rule := glinettest.Rule{Name: "r", Src: "wan", Proto: "tcp", DestPort: "50004", Target: "ACCEPT"}
	srv.Firewall.SetRules([]glinettest.Rule{rule})

	if err := c.Firewall.RemoveAllPortForwards(ctx); err != nil {
		t.Fatalf("RemoveAllPortForwards: %v", err)
	}
	all := map[string]any{"all": true}
	if !reflect.DeepEqual((*forwards)[0], all) {
		t.Errorf("remove_port_forward params = %v, want %v", (*forwards)[0], all)
	}
	if got := srv.Firewall.PortForwards(); len(got) != 0 {
		t.Errorf("server port forwards = %+v, want empty", got)
	}
	if got := srv.Firewall.Rules(); !slices.Equal(got, []glinettest.Rule{rule}) {
		t.Errorf("server rules = %+v, want %+v", got, rule)
	}

	if err := c.Firewall.RemoveAllRules(ctx); err != nil {
		t.Fatalf("RemoveAllRules: %v", err)
	}
	if !reflect.DeepEqual((*rules)[0], all) {
		t.Errorf("remove_rule params = %v, want %v", (*rules)[0], all)
	}
	if got := srv.Firewall.Rules(); len(got) != 0 {
		t.Errorf("server rules = %+v, want empty", got)
	}
}

func TestFirewallDMZ(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	got, err := c.Firewall.GetDMZ(ctx)
	if err != nil {
		t.Fatalf("GetDMZ: %v", err)
	}
	if *got != (glinet.DMZ{}) {
		t.Errorf("GetDMZ = %+v, want disabled without DestIP", got)
	}

	for _, want := range []glinet.DMZ{
		{Enabled: true, DestIP: "192.168.8.50"},
		{DestIP: "192.168.8.50"},
		{},
	} {
		if err := c.Firewall.SetDMZ(ctx, want); err != nil {
			t.Fatalf("SetDMZ(%+v): %v", want, err)
		}
		got, err := c.Firewall.GetDMZ(ctx)
		if err != nil {
			t.Fatalf("GetDMZ: %v", err)
		}
		if *got != want {
			t.Errorf("GetDMZ after SetDMZ(%+v) = %+v", want, got)
		}
		if state := srv.Firewall.DMZ(); state != (glinettest.DMZ{Enabled: want.Enabled, DestIP: want.DestIP}) {
			t.Errorf("server state after SetDMZ(%+v) = %+v", want, state)
		}
	}
}

func TestFirewallDMZPriority(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent map[string]any
	setDMZ := srv.Handler("firewall", "set_dmz")
	srv.Handle("firewall", "set_dmz", func(params json.RawMessage) (any, error) {
		sent = nil
		_ = json.Unmarshal(params, &sent)
		return setDMZ(params)
	})

	enabled := glinet.DMZ{Enabled: true, DestIP: "198.51.100.50", Priority: true}
	if err := c.Firewall.SetDMZ(ctx, enabled); err != nil {
		t.Fatalf("SetDMZ: %v", err)
	}
	if want := map[string]any{"enabled": true, "dest_ip": "198.51.100.50", "priority": true}; !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
	got, err := c.Firewall.GetDMZ(ctx)
	if err != nil {
		t.Fatalf("GetDMZ: %v", err)
	}
	if *got != enabled {
		t.Errorf("GetDMZ = %+v, want %+v", *got, enabled)
	}

	if err := c.Firewall.SetDMZ(ctx, glinet.DMZ{Priority: true}); err != nil {
		t.Fatalf("SetDMZ disable: %v", err)
	}
	if want := map[string]any{"enabled": false}; !reflect.DeepEqual(sent, want) {
		t.Errorf("params = %v, want %v (no priority when disabled)", sent, want)
	}
}

func TestFirewallOrderPortForwards(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent string
	order := srv.Handler("firewall", "order_port_forward")
	srv.Handle("firewall", "order_port_forward", func(params json.RawMessage) (any, error) {
		sent = string(params)
		return order(params)
	})
	srv.Firewall.SetPortForwards([]glinettest.PortForward{
		{Name: "first", Proto: "tcp", Src: "wan", SrcDport: "50001", Dest: "lan", DestIP: "198.51.100.51", DestPort: "50001", Enabled: true},
		{Name: "second", Proto: "tcp", Src: "wan", SrcDport: "50002", Dest: "lan", DestIP: "198.51.100.52", DestPort: "50002", Enabled: true},
		{Name: "third", Proto: "tcp", Src: "wan", SrcDport: "50003", Dest: "lan", DestIP: "198.51.100.53", DestPort: "50003", Enabled: true},
	})

	ids := []string{glinettest.PortForwardID(2), glinettest.PortForwardID(0), glinettest.PortForwardID(1)}
	if err := c.Firewall.OrderPortForwards(ctx, ids); err != nil {
		t.Fatalf("OrderPortForwards: %v", err)
	}
	want := `{"id_list":["` + ids[0] + `","` + ids[1] + `","` + ids[2] + `"]}`
	if sent != want {
		t.Errorf("params = %s, want %s", sent, want)
	}

	forwards, err := c.Firewall.ListPortForwards(ctx)
	if err != nil {
		t.Fatalf("ListPortForwards: %v", err)
	}
	var names []string
	for _, f := range forwards {
		names = append(names, f.Name)
	}
	if want := []string{"third", "first", "second"}; !slices.Equal(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}

	if err := c.Firewall.OrderPortForwards(ctx, nil); err != nil {
		t.Fatalf("OrderPortForwards nil: %v", err)
	}
	if want := `{"id_list":[]}`; sent != want {
		t.Errorf("params = %s, want %s", sent, want)
	}
}

func TestFirewallAddPortForwardAfterDMZ(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.Firewall.SetDMZ(glinettest.DMZ{})
	forward := glinet.PortForward{Name: "forward", Proto: "tcp", Src: "wan", SrcDport: "50002", Dest: "lan", DestIP: "192.168.8.50", DestPort: "50002"}

	id, err := c.Firewall.AddPortForward(ctx, forward)
	if err != nil {
		t.Fatalf("AddPortForward: %v", err)
	}
	if id != glinettest.PortForwardID(1) {
		t.Errorf("AddPortForward = %q, want %q (after the DMZ section)", id, glinettest.PortForwardID(1))
	}

	forwards, err := c.Firewall.ListPortForwards(ctx)
	if err != nil {
		t.Fatalf("ListPortForwards: %v", err)
	}
	if len(forwards) != 1 || forwards[0].ID != glinettest.PortForwardID(0) {
		t.Errorf("ListPortForwards = %+v, want ID %q (the DMZ section moved to the end)", forwards, glinettest.PortForwardID(0))
	}
}

func TestFirewallZonesDecodeRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	zones := fixtureResult(t, "firewall.get_zone_list.json")
	srv.Firewall.SetZones(zones)

	got, err := c.Firewall.GetZones(context.Background())
	if err != nil {
		t.Fatalf("GetZones: %v", err)
	}
	checkLossless(t, got, zones)
}

func TestFirewallWANAccess(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	setWANAccess := srv.Handler("firewall", "set_wan_access")
	srv.Handle("firewall", "set_wan_access", func(params json.RawMessage) (any, error) {
		sent = append(sent, string(params))
		return setWANAccess(params)
	})

	got, err := c.Firewall.GetWANAccess(ctx)
	if err != nil {
		t.Fatalf("GetWANAccess: %v", err)
	}
	if want := (glinet.WANAccess{Whitelist: []glinet.WANAccessSource{}}); !reflect.DeepEqual(*got, want) {
		t.Errorf("GetWANAccess = %+v, want %+v", *got, want)
	}

	enabled := glinet.WANAccess{
		EnableHTTPS:     true,
		EnableSSH:       true,
		EnablePing:      true,
		EnableWhitelist: true,
		Whitelist: []glinet.WANAccessSource{
			{IPAddr: "203.0.113.10", Name: "example"},
			{IPAddr: "2001:db8::10", Name: ""},
		},
	}
	if err := c.Firewall.SetWANAccess(ctx, enabled); err != nil {
		t.Fatalf("SetWANAccess: %v", err)
	}
	got, err = c.Firewall.GetWANAccess(ctx)
	if err != nil {
		t.Fatalf("GetWANAccess: %v", err)
	}
	if !reflect.DeepEqual(*got, enabled) {
		t.Errorf("GetWANAccess = %+v, want %+v", *got, enabled)
	}

	if err := c.Firewall.SetWANAccess(ctx, glinet.WANAccess{}); err != nil {
		t.Fatalf("SetWANAccess disable: %v", err)
	}
	if state := srv.Firewall.WANAccess(); !reflect.DeepEqual(state, glinettest.WANAccess{Whitelist: []glinettest.WANAccessSource{}}) {
		t.Errorf("server state = %+v, want all access disabled and an empty whitelist", state)
	}

	want := []string{
		`{"enable_https":true,"enable_ssh":true,"enable_ping":true,"enable_whitelist":true,"whitelist":[{"ipaddr":"203.0.113.10","name":"example"},{"ipaddr":"2001:db8::10","name":""}]}`,
		`{"enable_https":false,"enable_ssh":false,"enable_ping":false,"enable_whitelist":false,"whitelist":[]}`,
	}
	if !slices.Equal(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}
}

func TestFirewallWANAccessDecodeRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	result := fixtureResult(t, "firewall.get_wan_access.json")
	srv.Handle("firewall", "get_wan_access", func(json.RawMessage) (any, error) { return result, nil })

	res, err := c.Firewall.GetWANAccess(context.Background())
	if err != nil {
		t.Fatalf("GetWANAccess: %v", err)
	}
	checkLossless(t, res, result)
}
