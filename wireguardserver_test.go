package glinet_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestWireGuardServerReadsDecodeRouterFixtures(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	config, err := c.WireGuardServer.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, config, fixtureResult(t, "wg-server.get_config.json"))

	status, err := c.WireGuardServer.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, status, fixtureResult(t, "wg-server.get_status.json"))

	peers, err := c.WireGuardServer.ListPeers(ctx)
	if err != nil {
		t.Fatalf("ListPeers: %v", err)
	}
	checkLossless(t, map[string]any{"peers": peers}, fixtureResult(t, "wg-server.get_peer_list.json"))

	setting, err := c.WireGuardServer.GetSetting(ctx)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	checkLossless(t, setting, fixtureResult(t, "wg-server.get_setting.json"))

	routes, err := c.WireGuardServer.ListRoutes(ctx)
	if err != nil {
		t.Fatalf("ListRoutes: %v", err)
	}
	checkLossless(t, routes, fixtureResult(t, "wg-server.get_route_list.json"))
}

func TestWireGuardServerConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentParams(srv, "wg-server", "set_config")

	if err := c.WireGuardServer.Initialize(ctx); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	config, err := c.WireGuardServer.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !config.Initialization || config.PublicKey == "" {
		t.Errorf("GetConfig after Initialize = %+v", *config)
	}

	if err := c.WireGuardServer.SetConfig(ctx, glinet.WireGuardServerConfig{AddressV4: "10.0.1.1/24", Port: 51821, PublicKey: "not sent"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	want := map[string]any{"address_v4": "10.0.1.1/24", "address_v6": "", "port": float64(51821)}
	if got := (*sent)[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("set_config params = %v, want %v", got, want)
	}

	key := "DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD="
	if err := c.WireGuardServer.SetConfig(ctx, glinet.WireGuardServerConfig{AddressV4: "10.0.1.1/24", Port: 51821, PrivateKey: key, AddressV6: "fd00:1::1/64"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	want = map[string]any{"address_v4": "10.0.1.1/24", "port": float64(51821), "private_key": key, "ipv6_enable": false, "address_v6": "fd00:1::1/64"}
	if got := (*sent)[1]; !reflect.DeepEqual(got, want) {
		t.Errorf("set_config params = %v, want %v", got, want)
	}

	if err := c.WireGuardServer.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if status, err := c.WireGuardServer.GetStatus(ctx); err != nil || status.Server.Status != 1 {
		t.Errorf("GetStatus after Start = %+v, %v", status, err)
	}
	if err := c.WireGuardServer.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if status, err := c.WireGuardServer.GetStatus(ctx); err != nil || status.Server.Status != 0 {
		t.Errorf("GetStatus after Stop = %+v, %v", status, err)
	}

	setting := glinet.WireGuardServerSetting{ClientToClient: true, Masq: true, MTU: 1400}
	if err := c.WireGuardServer.SetSetting(ctx, setting); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}
	if got, err := c.WireGuardServer.GetSetting(ctx); err != nil || *got != setting {
		t.Errorf("GetSetting = %+v, %v, want %+v", got, err, setting)
	}
}

func TestWireGuardServerPeers(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentParams(srv, "wg-server", "add_peer")

	peer := glinet.WireGuardServerPeer{ID: 5, Name: "phone", ClientIP: "not sent", PresharedKeyEnable: true, AllowedIPs: "0.0.0.0/0"}
	id, err := c.WireGuardServer.AddPeer(ctx, peer)
	if err != nil {
		t.Fatalf("AddPeer: %v", err)
	}
	want := map[string]any{"name": "phone", "presharedkey_enable": true, "presharedkey": "", "allowed_ips": "0.0.0.0/0", "persistent_keepalive": "", "mtu": ""}
	if got := (*sent)[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("add_peer params = %v, want %v", got, want)
	}

	peer.ID = id
	peer.DNS = "10.0.0.1"
	peer.MTU = 1420
	if err := c.WireGuardServer.SetPeer(ctx, peer); err != nil {
		t.Fatalf("SetPeer: %v", err)
	}
	peers, err := c.WireGuardServer.ListPeers(ctx)
	if err != nil {
		t.Fatalf("ListPeers: %v", err)
	}
	if len(peers) != 1 || peers[0].ID != id || peers[0].DNS != "10.0.0.1" || peers[0].MTU != 1420 || peers[0].ClientIP == "not sent" {
		t.Errorf("ListPeers = %+v", peers)
	}

	peer.MTU = 0
	if err := c.WireGuardServer.SetPeer(ctx, peer); err != nil {
		t.Fatalf("SetPeer without MTU: %v", err)
	}
	if peers, err := c.WireGuardServer.ListPeers(ctx); err != nil || peers[0].MTU != 0 {
		t.Errorf("ListPeers after SetPeer without MTU = %+v, %v", peers, err)
	}

	config, err := c.WireGuardServer.GetPeerConfig(ctx, id)
	if err != nil {
		t.Fatalf("GetPeerConfig: %v", err)
	}
	if config.Address != peers[0].ClientIP || config.AllowedIPs != "0.0.0.0/0" {
		t.Errorf("GetPeerConfig = %+v", *config)
	}

	if err := c.WireGuardServer.RemovePeer(ctx, id); err != nil {
		t.Fatalf("RemovePeer: %v", err)
	}
	if peers, err := c.WireGuardServer.ListPeers(ctx); err != nil || len(peers) != 0 {
		t.Errorf("ListPeers after RemovePeer = %+v, %v", peers, err)
	}
	err = c.WireGuardServer.RemovePeer(ctx, id)
	if rpcCode(t, err) != glinet.CodeInvalidParams {
		t.Errorf("RemovePeer unknown: error = %v, want invalid params", err)
	}
}

func TestWireGuardServerRemoveAllPeers(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentParams(srv, "wg-server", "remove_peer")
	for _, name := range []string{"a", "b"} {
		if _, err := c.WireGuardServer.AddPeer(ctx, glinet.WireGuardServerPeer{Name: name}); err != nil {
			t.Fatalf("AddPeer %s: %v", name, err)
		}
	}

	if err := c.WireGuardServer.RemoveAllPeers(ctx); err != nil {
		t.Fatalf("RemoveAllPeers: %v", err)
	}
	if want := (map[string]any{"all": true}); !reflect.DeepEqual((*sent)[0], want) {
		t.Errorf("remove_peer params = %v, want %v", (*sent)[0], want)
	}
	if peers, err := c.WireGuardServer.ListPeers(ctx); err != nil || len(peers) != 0 {
		t.Errorf("ListPeers after RemoveAllPeers = %+v, %v, want empty", peers, err)
	}
}

func TestWireGuardServerRoutes(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	sent := sentParams(srv, "wg-server", "set_route")

	if err := c.WireGuardServer.AddRoute(ctx, glinet.VPNRoute{Dest: "192.168.60.0/24", Scope: "link"}); err != nil {
		t.Fatalf("AddRoute: %v", err)
	}
	routes, err := c.WireGuardServer.ListRoutes(ctx)
	if err != nil || len(routes.IPv4) != 1 {
		t.Fatalf("ListRoutes = %+v, %v", routes, err)
	}
	route := routes.IPv4[0]
	route.MTU = 1400
	if err := c.WireGuardServer.SetRoute(ctx, route); err != nil {
		t.Fatalf("SetRoute: %v", err)
	}
	want := map[string]any{"rule_id": float64(route.ID), "dest": "192.168.60.0/24", "gateway": "", "metric": float64(0), "mtu": float64(1400), "scope": "link", "route_flag": float64(4)}
	if got := (*sent)[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("set_route params = %v, want %v", got, want)
	}
	if routes, err := c.WireGuardServer.ListRoutes(ctx); err != nil || routes.IPv4[0] != route {
		t.Errorf("ListRoutes = %+v, %v, want %+v", routes, err, route)
	}
	if err := c.WireGuardServer.RemoveRoute(ctx, route); err != nil {
		t.Fatalf("RemoveRoute: %v", err)
	}
	if routes, err := c.WireGuardServer.ListRoutes(ctx); err != nil || len(routes.IPv4) != 0 {
		t.Errorf("ListRoutes after RemoveRoute = %+v, %v", routes, err)
	}
}

func TestWireGuardServerPeerAddressesAreUnique(t *testing.T) {
	c, _ := newClient(t)
	ctx := context.Background()

	a, err := c.WireGuardServer.AddPeer(ctx, glinet.WireGuardServerPeer{Name: "a"})
	if err != nil {
		t.Fatalf("AddPeer a: %v", err)
	}
	if _, err := c.WireGuardServer.AddPeer(ctx, glinet.WireGuardServerPeer{Name: "b"}); err != nil {
		t.Fatalf("AddPeer b: %v", err)
	}
	if err := c.WireGuardServer.RemovePeer(ctx, a); err != nil {
		t.Fatalf("RemovePeer a: %v", err)
	}
	if _, err := c.WireGuardServer.AddPeer(ctx, glinet.WireGuardServerPeer{Name: "c"}); err != nil {
		t.Fatalf("AddPeer c: %v", err)
	}

	peers, err := c.WireGuardServer.ListPeers(ctx)
	if err != nil {
		t.Fatalf("ListPeers: %v", err)
	}
	if len(peers) != 2 || peers[0].ClientIP == peers[1].ClientIP {
		t.Errorf("ListPeers = %+v, want 2 peers with different client IPs", peers)
	}
}

func TestWireGuardServerLargeIDsAndMetrics(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	const peer, metric = 4294967297, 4294967295
	replyParams(srv, "wg-server", "add_peer", `{"peer_id": 4294967297}`)
	removePeer := replyParams(srv, "wg-server", "remove_peer", `[]`)
	addRoute := sentParams(srv, "wg-server", "add_route")

	id, err := c.WireGuardServer.AddPeer(ctx, glinet.WireGuardServerPeer{Name: "example"})
	if err != nil || id != peer {
		t.Fatalf("AddPeer = %d, %v, want %d", id, err, int64(peer))
	}
	if err := c.WireGuardServer.RemovePeer(ctx, id); err != nil {
		t.Fatalf("RemovePeer: %v", err)
	}
	if got, want := (*removePeer)[0], map[string]any{"peer_id": float64(peer)}; !reflect.DeepEqual(got, want) {
		t.Errorf("remove_peer params = %v, want %v", got, want)
	}

	if err := c.WireGuardServer.AddRoute(ctx, glinet.VPNRoute{Dest: "192.168.60.0/24", Metric: metric}); err != nil {
		t.Fatalf("AddRoute: %v", err)
	}
	if got := (*addRoute)[0]["metric"]; got != float64(metric) {
		t.Errorf("add_route metric = %v, want %d", got, int64(metric))
	}
	routes, err := c.WireGuardServer.ListRoutes(ctx)
	if err != nil || len(routes.IPv4) != 1 || routes.IPv4[0].Metric != metric {
		t.Errorf("ListRoutes = %+v, %v, want metric %d", routes, err, int64(metric))
	}
}
