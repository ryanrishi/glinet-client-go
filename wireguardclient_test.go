package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

// sentParams records the params of each call of module.method on srv. The last call is last.
func sentParams(srv *glinettest.Server, module, method string) *[]map[string]any {
	var sent []map[string]any
	next := srv.Handler(module, method)
	srv.Handle(module, method, func(params json.RawMessage) (any, error) {
		var p map[string]any
		_ = json.Unmarshal(params, &p)
		sent = append(sent, p)
		return next(params)
	})
	return &sent
}

const (
	customGroup = 2082
	tunnelPeer  = 2004
)

func TestWireGuardClientReadsDecodeRouterFixtures(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	status, err := c.WireGuardClient.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, status, fixtureResult(t, "wg-client.get_status.json"))
	if status.Status != 1 || status.GroupID != customGroup || status.PeerID != tunnelPeer {
		t.Errorf("GetStatus = %+v", *status)
	}

	groups, err := c.WireGuardClient.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	checkLossless(t, map[string]any{"groups": groups}, fixtureResult(t, "wg-client.get_group_list.json"))

	for groupID, f := range map[int64]string{customGroup: "wg-client.get_config_list.json", 9859: "wg-client.get_config_list.group_type_3.json"} {
		peers, err := c.WireGuardClient.ListPeers(ctx, groupID)
		if err != nil {
			t.Fatalf("ListPeers: %v", err)
		}
		checkLossless(t, map[string]any{"peers": peers}, fixtureResult(t, f), "auth_type")
	}

	all := fixtureResult(t, "wg-client.get_all_config_list.json")
	srv.Handle("wg-client", "get_all_config_list", func(json.RawMessage) (any, error) { return all, nil })
	allPeers, err := c.WireGuardClient.ListAllPeers(ctx)
	if err != nil {
		t.Fatalf("ListAllPeers: %v", err)
	}
	checkLossless(t, map[string]any{"config_list": allPeers}, all)

	setting, err := c.WireGuardClient.GetSetting(ctx, customGroup, tunnelPeer)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	checkLossless(t, setting, fixtureResult(t, "wg-client.get_setting.json"))

	routes, err := c.WireGuardClient.ListRoutes(ctx, customGroup, tunnelPeer)
	if err != nil {
		t.Fatalf("ListRoutes: %v", err)
	}
	checkLossless(t, routes, fixtureResult(t, "wg-client.get_route_list.json"))

	peer, err := c.WireGuardClient.FindPeer(ctx, customGroup, tunnelPeer)
	if err != nil || peer == nil || peer.Name != status.Name {
		t.Errorf("FindPeer = %+v, %v, want the peer of the tunnel", peer, err)
	}
	if peer, err := c.WireGuardClient.FindPeer(ctx, customGroup, 1); err != nil || peer != nil {
		t.Errorf("FindPeer unknown = %+v, %v, want nil, nil", peer, err)
	}
}

func TestWireGuardClientGroupsAndPeers(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	groupID, err := c.WireGuardClient.AddGroup(ctx, "new group")
	if err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	if err := c.WireGuardClient.RenameGroup(ctx, groupID, "office"); err != nil {
		t.Fatalf("RenameGroup: %v", err)
	}
	groups, err := c.WireGuardClient.ListGroups(ctx)
	if err != nil {
		t.Fatalf("ListGroups: %v", err)
	}
	last := groups[len(groups)-1]
	if last.ID != groupID || last.Name != "office" || last.Type != glinet.WireGuardGroupCustom || last.PeerCount != 0 {
		t.Errorf("new group = %+v", last)
	}

	addSent := sentParams(srv, "wg-client", "add_config")
	peer := glinet.WireGuardPeer{
		ID:                  99,
		Name:                "office-1",
		AddressV4:           "10.0.0.2/32",
		PrivateKey:          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		PublicKey:           "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
		PresharedKey:        "not sent while disabled",
		AllowedIPs:          "0.0.0.0/0",
		Endpoint:            "vpn.example.com:51820",
		PersistentKeepalive: 25,
	}
	peerID, err := c.WireGuardClient.AddPeer(ctx, groupID, peer)
	if err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	want := map[string]any{
		"group_id": float64(groupID), "name": "office-1", "address_v4": "10.0.0.2/32", "address_v6": "", "ipv6_enable": false,
		"private_key": peer.PrivateKey, "public_key": peer.PublicKey, "allowed_ips": "0.0.0.0/0",
		"end_point": "vpn.example.com:51820", "presharedkey_enable": false, "persistent_keepalive": float64(25),
	}
	if got := (*addSent)[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("add_config params = %v, want %v", got, want)
	}

	peer.ID = peerID
	peer.PresharedKeyEnable = true
	peer.PresharedKey = "CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC="
	peer.DNS = "10.0.0.1"
	peer.MTU = 1420
	if err := c.WireGuardClient.SetPeer(ctx, groupID, peer); err != nil {
		t.Fatalf("SetPeer: %v", err)
	}
	got, err := c.WireGuardClient.FindPeer(ctx, groupID, peerID)
	if err != nil {
		t.Fatalf("FindPeer: %v", err)
	}
	if got == nil || *got != peer {
		t.Errorf("FindPeer = %+v, want %+v", got, peer)
	}

	setSent := sentParams(srv, "wg-client", "set_config")
	peer.PersistentKeepalive = 0
	if err := c.WireGuardClient.SetPeer(ctx, groupID, peer); err != nil {
		t.Fatalf("SetPeer without keepalive: %v", err)
	}
	sent := (*setSent)[0]
	if sent["persistent_keepalive"] != "" || sent["listen_port"] != "" || sent["mtu"] != float64(1420) {
		t.Errorf("set_config params = %v, want persistent_keepalive and listen_port \"\" and mtu 1420", sent)
	}
	if got, err := c.WireGuardClient.FindPeer(ctx, groupID, peerID); err != nil || got == nil || *got != peer {
		t.Errorf("FindPeer = %+v, %v, want %+v", got, err, peer)
	}

	if err := c.WireGuardClient.RemovePeer(ctx, groupID, peerID); err != nil {
		t.Fatalf("RemovePeer: %v", err)
	}
	if peers, err := c.WireGuardClient.ListPeers(ctx, groupID); err != nil || len(peers) != 0 {
		t.Errorf("ListPeers after RemovePeer = %+v, %v, want empty", peers, err)
	}

	if _, err := c.WireGuardClient.AddPeer(ctx, groupID, peer); err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	if err := c.WireGuardClient.RemoveAllPeers(ctx, groupID); err != nil {
		t.Fatalf("RemoveAllPeers: %v", err)
	}
	if peers, err := c.WireGuardClient.ListPeers(ctx, groupID); err != nil || len(peers) != 0 {
		t.Errorf("ListPeers after RemoveAllPeers = %+v, %v, want empty", peers, err)
	}

	if err := c.WireGuardClient.RemoveGroup(ctx, groupID); err != nil {
		t.Fatalf("RemoveGroup: %v", err)
	}
	_, err = c.WireGuardClient.ListPeers(ctx, groupID)
	if rpcCode(t, err) != glinet.CodeInvalidParams {
		t.Errorf("ListPeers of a removed group: error = %v, want invalid params", err)
	}
}

func TestWireGuardClientAddWithoutNewID(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	srv.Handle("wg-client", "add_group", func(json.RawMessage) (any, error) { return []any{}, nil })
	srv.Handle("wg-client", "add_config", func(json.RawMessage) (any, error) { return []any{}, nil })

	if _, err := c.WireGuardClient.AddGroup(ctx, "x"); err == nil {
		t.Error("AddGroup without a new group: no error")
	}
	if _, err := c.WireGuardClient.AddPeer(ctx, customGroup, glinet.WireGuardPeer{Name: "x"}); err == nil {
		t.Error("AddPeer without a new peer: no error")
	}
}

func TestWireGuardClientErrorHasNoSecret(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("wg-client", "set_config", func(json.RawMessage) (any, error) {
		return map[string]any{"err_code": 20001206, "err_msg": "invalid private key"}, nil
	})

	const key = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	err := c.WireGuardClient.SetPeer(context.Background(), customGroup, glinet.WireGuardPeer{ID: tunnelPeer, PrivateKey: key, PresharedKey: key, PresharedKeyEnable: true})
	if rpcCode(t, err) != 20001206 {
		t.Errorf("SetPeer error = %v, want code 20001206", err)
	}
	if strings.Contains(err.Error(), key) {
		t.Errorf("error %q has the key", err)
	}
}

func TestWireGuardClientTunnel(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	if err := c.WireGuardClient.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if status, err := c.WireGuardClient.GetStatus(ctx); err != nil || status.Status != 0 {
		t.Errorf("GetStatus after Stop = %+v, %v, want status 0", status, err)
	}

	if err := c.WireGuardClient.Start(ctx, customGroup, 3578); err != nil {
		t.Fatalf("Start: %v", err)
	}
	status, err := c.WireGuardClient.GetStatus(ctx)
	if err != nil || status.Status != 1 || status.GroupID != customGroup || status.PeerID != 3578 {
		t.Errorf("GetStatus after Start = %+v, %v", status, err)
	}

	setting := glinet.WireGuardClientSetting{LocalAccess: true, Masq: false, MTU: 1380}
	if err := c.WireGuardClient.SetSetting(ctx, customGroup, 3578, setting); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if got, err := c.WireGuardClient.GetSetting(ctx, customGroup, 3578); err != nil || *got != setting {
		t.Errorf("GetSetting = %+v, %v, want %+v", got, err, setting)
	}
}

func TestWireGuardClientRoutes(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentParams(srv, "wg-client", "add_route")

	route := glinet.VPNRoute{ID: 7, Dest: "192.168.50.0/24", Gateway: "10.0.0.1", Metric: 10}
	if err := c.WireGuardClient.AddRoute(ctx, customGroup, tunnelPeer, route); err != nil {
		t.Fatalf("AddRoute: %v", err)
	}
	want := map[string]any{"group_id": float64(customGroup), "peer_id": float64(tunnelPeer), "dest": "192.168.50.0/24", "gateway": "10.0.0.1", "metric": float64(10), "route_flag": float64(4)}
	if got := (*sent)[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("add_route params = %v, want %v", got, want)
	}
	if err := c.WireGuardClient.AddRoute(ctx, customGroup, tunnelPeer, glinet.VPNRoute{Dest: "fd00:1::/64"}); err != nil {
		t.Fatalf("AddRoute IPv6: %v", err)
	}

	routes, err := c.WireGuardClient.ListRoutes(ctx, customGroup, tunnelPeer)
	if err != nil {
		t.Fatalf("ListRoutes: %v", err)
	}
	if len(routes.IPv4) != 1 || len(routes.IPv6) != 1 || routes.IPv4[0].Dest != route.Dest {
		t.Fatalf("ListRoutes = %+v", routes)
	}

	route = routes.IPv4[0]
	route.Metric = 20
	if err := c.WireGuardClient.SetRoute(ctx, customGroup, tunnelPeer, route); err != nil {
		t.Fatalf("SetRoute: %v", err)
	}
	if err := c.WireGuardClient.RemoveRoute(ctx, customGroup, tunnelPeer, routes.IPv6[0]); err != nil {
		t.Fatalf("RemoveRoute: %v", err)
	}
	routes, err = c.WireGuardClient.ListRoutes(ctx, customGroup, tunnelPeer)
	if err != nil {
		t.Fatalf("ListRoutes: %v", err)
	}
	if len(routes.IPv4) != 1 || routes.IPv4[0] != route || len(routes.IPv6) != 0 {
		t.Errorf("ListRoutes = %+v, want only %+v", routes, route)
	}
}

func TestWireGuardClientStartSetsEndpoint(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	peerID, err := c.WireGuardClient.AddPeer(ctx, customGroup, glinet.WireGuardPeer{Name: "other", AddressV4: "10.0.0.3/32", Endpoint: "vpn-test.example.com:51821"})
	if err != nil {
		t.Fatalf("AddPeer: %v", err)
	}

	for _, tt := range []struct {
		peerID int64
		domain string
		port   int
	}{
		{peerID, "vpn-test.example.com", 51821},
		{3578, "vpn-1.example.com", 51820},
	} {
		if err := c.WireGuardClient.Start(ctx, customGroup, tt.peerID); err != nil {
			t.Fatalf("Start: %v", err)
		}
		status, err := c.WireGuardClient.GetStatus(ctx)
		if err != nil || status.PeerID != tt.peerID || status.Domain != tt.domain || status.Port != tt.port {
			t.Errorf("GetStatus after Start of %d = %+v, %v, want domain %s port %d", tt.peerID, status, err, tt.domain, tt.port)
		}
	}
}

func TestWireGuardClientConcurrentAdds(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	const n = 8
	ids := make([]int64, 2*n)
	errs := make([]error, 2*n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(2)
		go func() {
			defer wg.Done()
			ids[i], errs[i] = c.WireGuardClient.AddGroup(ctx, "group")
		}()
		go func() {
			defer wg.Done()
			ids[n+i], errs[n+i] = c.WireGuardClient.AddPeer(ctx, customGroup, glinet.WireGuardPeer{Name: "peer"})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("call %d: %v", i, err)
		}
	}
	slices.Sort(ids)
	if len(slices.Compact(slices.Clone(ids))) != len(ids) {
		t.Errorf("IDs = %v, want all different", ids)
	}
}

// replyParams replaces module.method on srv with a method that records its params and returns
// result.
func replyParams(srv *glinettest.Server, module, method, result string) *[]map[string]any {
	var sent []map[string]any
	srv.Handle(module, method, func(params json.RawMessage) (any, error) {
		var p map[string]any
		_ = json.Unmarshal(params, &p)
		sent = append(sent, p)
		return json.RawMessage(result), nil
	})
	return &sent
}

func TestWireGuardClientLargeIDsAndMetrics(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	const group, peer, rule, metric = 4294967296, 4294967297, 4294967298, 4294967295
	replyParams(srv, "wg-client", "get_group_list", `{"groups": [{"group_id": 4294967296, "group_name": "example", "group_type": 2}]}`)
	replyParams(srv, "wg-client", "get_config_list", `{"peers": [{"peer_id": 4294967297, "name": "example"}]}`)
	replyParams(srv, "wg-client", "get_status", `{"status": 1, "group_id": 4294967296, "peer_id": 4294967297}`)
	replyParams(srv, "wg-client", "get_setting", `{"default_metric": 4294967295}`)
	replyParams(srv, "wg-client", "get_route_list", `{"ipv4_route_rules": [{"rule_id": 4294967298, "dest": "192.168.50.0/24", "metric": 4294967295}], "ipv6_route_rules": []}`)
	setSetting := replyParams(srv, "wg-client", "set_setting", `[]`)
	setRoute := replyParams(srv, "wg-client", "set_route", `[]`)
	removeConfig := replyParams(srv, "wg-client", "remove_config", `[]`)

	groups, err := c.WireGuardClient.ListGroups(ctx)
	if err != nil || len(groups) != 1 || groups[0].ID != group {
		t.Errorf("ListGroups = %+v, %v, want group_id %d", groups, err, int64(group))
	}
	if p, err := c.WireGuardClient.FindPeer(ctx, group, peer); err != nil || p == nil {
		t.Errorf("FindPeer = %+v, %v, want peer_id %d", p, err, int64(peer))
	}
	if s, err := c.WireGuardClient.GetStatus(ctx); err != nil || s.GroupID != group || s.PeerID != peer {
		t.Errorf("GetStatus = %+v, %v", s, err)
	}
	setting, err := c.WireGuardClient.GetSetting(ctx, group, peer)
	if err != nil || setting.DefaultMetric != metric {
		t.Fatalf("GetSetting = %+v, %v, want default_metric %d", setting, err, int64(metric))
	}
	routes, err := c.WireGuardClient.ListRoutes(ctx, group, peer)
	if err != nil || len(routes.IPv4) != 1 || routes.IPv4[0].ID != rule || routes.IPv4[0].Metric != metric {
		t.Fatalf("ListRoutes = %+v, %v", routes, err)
	}

	if err := c.WireGuardClient.SetSetting(ctx, group, peer, *setting); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if err := c.WireGuardClient.SetRoute(ctx, group, peer, routes.IPv4[0]); err != nil {
		t.Fatalf("SetRoute: %v", err)
	}
	if err := c.WireGuardClient.RemovePeer(ctx, group, peer); err != nil {
		t.Fatalf("RemovePeer: %v", err)
	}
	for name, tt := range map[string]struct {
		got, want map[string]any
	}{
		"set_setting":   {(*setSetting)[0], map[string]any{"group_id": float64(group), "peer_id": float64(peer), "default_metric": float64(metric), "local_access": false, "masq": false, "mtu": float64(0)}},
		"set_route":     {(*setRoute)[0], map[string]any{"group_id": float64(group), "peer_id": float64(peer), "rule_id": float64(rule), "dest": "192.168.50.0/24", "metric": float64(metric), "route_flag": float64(4)}},
		"remove_config": {(*removeConfig)[0], map[string]any{"group_id": float64(group), "peer_id": float64(peer)}},
	} {
		if !reflect.DeepEqual(tt.got, tt.want) {
			t.Errorf("%s params = %v, want %v", name, tt.got, tt.want)
		}
	}
}
