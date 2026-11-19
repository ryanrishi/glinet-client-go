package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
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
	srv.Clients.SetList(json.RawMessage(`{"clients": [{"mac": "02:00:00:00:00:01", "name": "", "alias": "printer", "class": "printer", "blocked": true, "limit_rx": 100, "limit_tx": 0.125, "online": true, "total_rx": 5000000000, "ipv6": ["2001:db8::10"], "last_rx": [1, 2], "last_tx": [3]}]}`))

	clients, err := c.Clients.ListClients(context.Background())
	if err != nil {
		t.Fatalf("ListClients: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("ListClients = %+v, want 1 client", clients)
	}
	got := clients[0]
	if got.MAC != "02:00:00:00:00:01" || got.Alias != "printer" || got.Class != "printer" || !got.Blocked || got.LimitRx != 100 || got.LimitTx != 0.125 || !got.Online || got.TotalRx != 5000000000 ||
		!slices.Equal(got.IPv6, []string{"2001:db8::10"}) || !slices.Equal(got.LastRx, []int64{1, 2}) || !slices.Equal(got.LastTx, []int64{3}) {
		t.Errorf("ListClients = %+v", got)
	}
}

func TestClientsSetInfo(t *testing.T) {
	tests := []struct {
		name      string
		info      glinet.ClientInfo
		wantSent  string
		wantAlias string
		wantClass string
	}{
		{
			name:      "set",
			info:      glinet.ClientInfo{MAC: "02:00:00:00:00:16", Alias: "example", Class: "printer"},
			wantSent:  `{"mac":"02:00:00:00:00:16","alias":"example","class":"printer"}`,
			wantAlias: "example",
			wantClass: "printer",
		},
		{
			name:     "remove",
			info:     glinet.ClientInfo{MAC: "02:00:00:00:00:27"},
			wantSent: `{"mac":"02:00:00:00:00:27","alias":"","class":""}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, srv := newClient(t)
			ctx := context.Background()
			sent := sentParams(srv, "clients", "set_info")

			if err := c.Clients.SetInfo(ctx, tt.info); err != nil {
				t.Fatalf("SetInfo: %v", err)
			}
			checkClientWrite(t, *sent, tt.wantSent)

			got := clientByMAC(t, c, tt.info.MAC)
			if got.Alias != tt.wantAlias || got.Class != tt.wantClass {
				t.Errorf("client = %+v, want alias %q and class %q", got, tt.wantAlias, tt.wantClass)
			}
		})
	}
}

func TestClientsSetInfoUnknownMACChangesNothing(t *testing.T) {
	c, srv := newClient(t)
	before := srv.Clients.List()

	if err := c.Clients.SetInfo(context.Background(), glinet.ClientInfo{MAC: "02:00:00:00:00:99", Alias: "example"}); err != nil {
		t.Fatalf("SetInfo: %v", err)
	}
	var want, got any
	_ = json.Unmarshal(before, &want)
	_ = json.Unmarshal(srv.Clients.List(), &got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("client list changed:\n%s", srv.Clients.List())
	}
}

// checkClientWrite checks that sent has one params object, equal to the JSON object want.
func checkClientWrite(t *testing.T, sent []map[string]any, want string) {
	t.Helper()
	var w map[string]any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	if len(sent) != 1 || !reflect.DeepEqual(sent[0], w) {
		t.Errorf("sent %v, want [%s]", sent, want)
	}
}

func clientByMAC(t *testing.T, c *glinet.Client, mac string) glinet.ClientDevice {
	t.Helper()
	clients, err := c.Clients.ListClients(context.Background())
	if err != nil {
		t.Fatalf("ListClients: %v", err)
	}
	i := slices.IndexFunc(clients, func(d glinet.ClientDevice) bool { return d.MAC == mac })
	if i < 0 {
		t.Fatalf("ListClients has no client %s", mac)
	}
	return clients[i]
}
