package glinet_test

import (
	"context"
	"encoding/json"
	"testing"
)

func TestUSB3Disabled(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	handler := srv.Handler("system", "set_usb3_disable")
	srv.Handle("system", "set_usb3_disable", func(params json.RawMessage) (any, error) {
		sent = append(sent, string(params))
		return handler(params)
	})

	for _, disabled := range []bool{true, false} {
		sent = nil
		if err := c.System.SetUSB3Disabled(ctx, disabled); err != nil {
			t.Fatalf("SetUSB3Disabled(%t): %v", disabled, err)
		}
		want := `{"usb3_disable":false}`
		if disabled {
			want = `{"usb3_disable":true}`
		}
		if len(sent) != 1 || sent[0] != want {
			t.Errorf("params = %v, want %s", sent, want)
		}
		got, err := c.System.GetUSB3Disabled(ctx)
		if err != nil {
			t.Fatalf("GetUSB3Disabled: %v", err)
		}
		if got != disabled || srv.System.USB3Disabled() != disabled {
			t.Errorf("GetUSB3Disabled = %t, server state %t, want %t", got, srv.System.USB3Disabled(), disabled)
		}
	}
}

func TestUSB3DisabledWithoutValue(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("system", "get_usb3_disable", func(json.RawMessage) (any, error) { return json.RawMessage(`{}`), nil })

	got, err := c.System.GetUSB3Disabled(context.Background())
	if err != nil {
		t.Fatalf("GetUSB3Disabled: %v", err)
	}
	if !got {
		t.Error("GetUSB3Disabled = false, want true as the web UI shows")
	}
}
