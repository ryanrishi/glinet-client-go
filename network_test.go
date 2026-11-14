package glinet_test

import (
	"context"
	"encoding/json"
	"testing"
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

	got, err := c.Network.CheckWANCable(context.Background())
	if err != nil {
		t.Fatalf("CheckWANCable: %v", err)
	}
	if params != `{"secondwan":false}` {
		t.Errorf("params = %s, want the params of the fixture", params)
	}
	checkLossless(t, got, wanCable)
}
