package glinet_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestAccessControlBlockList(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.AccessControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if res.Mode != "black" || res.BlackMAC != nil {
		t.Errorf("GetConfig = %+v, want mode black and no MACs", *res)
	}

	for _, mac := range []string{"02:00:00:00:00:a1", "02:00:00:00:00:A2", "02:00:00:00:00:a1"} {
		if err := c.AccessControl.AddBlockedMAC(ctx, mac); err != nil {
			t.Fatalf("AddBlockedMAC(%s): %v", mac, err)
		}
	}
	if got := srv.AccessControl.Config().BlackMAC; !slices.Equal(got, []string{"02:00:00:00:00:a1", "02:00:00:00:00:A2"}) {
		t.Errorf("server block list = %v", got)
	}

	if err := c.AccessControl.RemoveBlockedMAC(ctx, "02:00:00:00:00:a2"); err != nil {
		t.Fatalf("RemoveBlockedMAC: %v", err)
	}
	if err := c.AccessControl.RemoveBlockedMAC(ctx, "02:00:00:00:00:a1"); err != nil {
		t.Fatalf("RemoveBlockedMAC: %v", err)
	}
	res, err = c.AccessControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !slices.Equal(res.BlackMAC, []string{"02:00:00:00:00:A2"}) {
		t.Errorf("GetConfig = %+v, want only the MAC in the other letter case", *res)
	}
}

func TestAccessControlDecodeRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	result := fixtureResult(t, "black_white_list.get_config.after_add_upper.json")
	srv.Handle("black_white_list", "get_config", func(json.RawMessage) (any, error) { return result, nil })

	res, err := c.AccessControl.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, res, result)
}

func TestAccessControlAllowList(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	for _, method := range []string{"set_config", "set_single_mac"} {
		handler := srv.Handler("black_white_list", method)
		srv.Handle("black_white_list", method, func(params json.RawMessage) (any, error) {
			sent = append(sent, string(params))
			return handler(params)
		})
	}
	srv.AccessControl.SetConfig(glinettest.AccessControlConfig{Mode: "black", BlackMAC: []string{"02:00:00:00:00:B1"}})

	if err := c.AccessControl.SetConfig(ctx, glinet.AccessControlAllowList, []string{"02:00:00:00:00:A1", "02:00:00:00:00:A2"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if err := c.AccessControl.AddAllowedMAC(ctx, "02:00:00:00:00:A3"); err != nil {
		t.Fatalf("AddAllowedMAC: %v", err)
	}
	if err := c.AccessControl.RemoveAllowedMAC(ctx, "02:00:00:00:00:A1"); err != nil {
		t.Fatalf("RemoveAllowedMAC: %v", err)
	}
	want := []string{
		`{"mode":"white","mac":["02:00:00:00:00:A1","02:00:00:00:00:A2"]}`,
		`{"mode":"white","operate":"add","mac":"02:00:00:00:00:A3"}`,
		`{"mode":"white","operate":"del","mac":"02:00:00:00:00:A1"}`,
	}
	if !slices.Equal(sent, want) {
		t.Errorf("params = %v, want %v", sent, want)
	}

	got, err := c.AccessControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Mode != glinet.AccessControlAllowList || !slices.Equal(got.WhiteMAC, []string{"02:00:00:00:00:A2", "02:00:00:00:00:A3"}) || !slices.Equal(got.BlackMAC, []string{"02:00:00:00:00:B1"}) {
		t.Errorf("GetConfig = %+v", *got)
	}

	if err := c.AccessControl.SetConfig(ctx, glinet.AccessControlBlockList, nil); err != nil {
		t.Fatalf("SetConfig block list: %v", err)
	}
	if want := `{"mode":"black","mac":[]}`; sent[len(sent)-1] != want {
		t.Errorf("params = %s, want %s", sent[len(sent)-1], want)
	}
	got, err = c.AccessControl.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if got.Mode != glinet.AccessControlBlockList || got.BlackMAC != nil || len(got.WhiteMAC) != 2 {
		t.Errorf("GetConfig = %+v, want mode black, an empty block list and the allow list", *got)
	}
}
