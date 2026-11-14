package glinettest

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"
)

// TestClientsDefaultIsFromRouterFixture checks that each default client is a client in the
// get_list fixture of the test router.
func TestClientsDefaultIsFromRouterFixture(t *testing.T) {
	data, err := os.ReadFile("../testdata/fixtures/axt1800_4.7.0/clients.get_list.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Response struct {
			Result struct {
				Clients []any `json:"clients"`
			} `json:"result"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}

	var list struct {
		Clients []any `json:"clients"`
	}
	if err := json.Unmarshal(newClients().List(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Clients) == 0 {
		t.Fatal("no default clients")
	}
	for _, c := range list.Clients {
		if !slices.ContainsFunc(f.Response.Result.Clients, func(fc any) bool { return reflect.DeepEqual(c, fc) }) {
			t.Errorf("default client %v is not in the fixture", c)
		}
	}
}
