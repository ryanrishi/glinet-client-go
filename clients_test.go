package glinet_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
)

func TestClientsDecodeRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	list := fixtureResult(t, "clients.get_list.json")
	srv.Clients.SetList(list)

	clients, err := c.Clients.ListClients(context.Background())
	if err != nil {
		t.Fatalf("ListClients: %v", err)
	}
	if len(clients) != 51 {
		t.Fatalf("ListClients returned %d clients, want 51", len(clients))
	}
	checkLossless(t, map[string]any{"clients": clients}, list)
}

func TestClientsList(t *testing.T) {
	c, srv := newClient(t)
	srv.Clients.SetList(json.RawMessage(`{"clients": [{"mac": "02:00:00:00:00:01", "name": "", "alias": "printer", "class": "printer", "blocked": true, "limit_rx": 100, "online": true, "total_rx": 5000000000, "ipv6": ["2001:db8::10"], "last_rx": [1, 2], "last_tx": [3]}]}`))

	clients, err := c.Clients.ListClients(context.Background())
	if err != nil {
		t.Fatalf("ListClients: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("ListClients = %+v, want 1 client", clients)
	}
	got := clients[0]
	if got.MAC != "02:00:00:00:00:01" || got.Alias != "printer" || got.Class != "printer" || !got.Blocked || got.LimitRx != 100 || !got.Online || got.TotalRx != 5000000000 ||
		!slices.Equal(got.IPv6, []string{"2001:db8::10"}) || !slices.Equal(got.LastRx, []int64{1, 2}) || !slices.Equal(got.LastTx, []int64{3}) {
		t.Errorf("ListClients = %+v", got)
	}
}
