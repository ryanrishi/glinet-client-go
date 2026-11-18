package glinet_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

func TestNetworkWANCableDecodesRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	wanCable := fixtureResult(t, "network.check_wan_cable.json")
	srv.Network.SetWANCable(wanCable)

	var params string
	handler := srv.Handler("network", "check_wan_cable")
	srv.Handle("network", "check_wan_cable", func(p json.RawMessage) (any, error) {
		params = string(p)
		return handler(p)
	})

	got, err := c.Network.CheckWANCable(context.Background(), glinet.FirstWANPort)
	if err != nil {
		t.Fatalf("CheckWANCable: %v", err)
	}
	if params != `{"secondwan":false}` {
		t.Errorf("params = %s, want the params of the fixture", params)
	}
	checkLossless(t, got, wanCable)

	if _, err := c.Network.CheckWANCable(context.Background(), glinet.SecondWANPort); err != nil {
		t.Fatalf("CheckWANCable second: %v", err)
	}
	if params != `{"secondwan":true}` {
		t.Errorf("params = %s, want secondwan true", params)
	}
}

func TestNetworkARPListDecodesRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	arpList := fixtureResult(t, "network.get_arp_list.json")
	srv.Network.SetARPList(arpList)

	got, err := c.Network.ListARPEntries(context.Background())
	if err != nil {
		t.Fatalf("ListARPEntries: %v", err)
	}
	checkLossless(t, map[string]any{"entries": got}, arpList)
}
