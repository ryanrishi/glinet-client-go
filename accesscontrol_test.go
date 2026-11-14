package glinet_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
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
