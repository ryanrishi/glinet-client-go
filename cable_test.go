package glinet_test

import (
	"context"
	"encoding/json"
	"testing"
)

// TestCableStatusNestedErrCode checks that the err_code of secondwan in the get_status result of
// the test router is not an error, while a top-level err_code is.
func TestCableStatusNestedErrCode(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	status := fixtureResult(t, "cable.get_status.json")
	srv.Cable.SetStatus(status)

	got, err := c.Cable.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if got.Status != 3 || got.SecondWAN.ErrCode != -4 || got.SecondWAN.ErrMsg != "Unsupported features!" {
		t.Errorf("GetStatus = %+v", *got)
	}
	checkLossless(t, got, status)

	srv.Cable.SetStatus(json.RawMessage(`{"err_code": -4, "err_msg": "Unsupported features!"}`))
	if _, err := c.Cable.GetStatus(ctx); rpcCode(t, err) != -4 {
		t.Errorf("GetStatus with a top-level err_code: error = %v, want code -4", err)
	}
}
