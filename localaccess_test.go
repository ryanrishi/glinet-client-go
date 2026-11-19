package glinet_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestLocalAccessConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent string
	setConfig := srv.Handler("local-access", "set_config")
	srv.Handle("local-access", "set_config", func(params json.RawMessage) (any, error) {
		sent = string(params)
		return setConfig(params)
	})

	got, err := c.LocalAccess.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	initial := glinet.LocalAccessConfig{HTTPPort: 80, HTTPSPort: 443, LuCIHTTPPort: 8080, LuCIHTTPSPort: 8443, SessionTimeout: 300, SSHEnabled: true, SSHPort: 22}
	if *got != initial {
		t.Errorf("GetConfig = %+v, want %+v", *got, initial)
	}

	changed := glinet.LocalAccessConfig{
		HTTPPort:          8000,
		HTTPSPort:         8001,
		RedirectHTTPS:     true,
		LuCIHTTPPort:      8002,
		LuCIHTTPSPort:     8003,
		LuCIRedirectHTTPS: true,
		SessionTimeout:    3600,
		SSHPort:           2222,
	}
	if err := c.LocalAccess.SetConfig(ctx, changed); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}
	if want := `{"http_port":8000,"https_port":8001,"redirect_https":true,"luci_http_port":8002,"luci_https_port":8003,"luci_redirect_https":true,"session_timeout":3600,"ssh_enabled":false,"ssh_port":2222}`; sent != want {
		t.Errorf("params = %s, want %s", sent, want)
	}
	if state := srv.LocalAccess.Config(); state != glinettest.LocalAccessConfig(changed) {
		t.Errorf("server state = %+v, want %+v", state, changed)
	}
	got, err = c.LocalAccess.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if *got != changed {
		t.Errorf("GetConfig = %+v, want %+v", *got, changed)
	}
}

func TestLocalAccessSetConfigPortInUse(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("local-access", "set_config", func(json.RawMessage) (any, error) {
		return map[string]any{"err_code": -1, "err_msg": "http port is occupied"}, nil
	})

	err := c.LocalAccess.SetConfig(context.Background(), glinet.LocalAccessConfig{HTTPPort: 8000, HTTPSPort: 443, LuCIHTTPPort: 8080, LuCIHTTPSPort: 8443, SessionTimeout: 300, SSHPort: 22})
	var rpcErr *glinet.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != -1 {
		t.Errorf("SetConfig error = %v, want an *RPCError with code -1", err)
	}
}

func TestLocalAccessDecodeRouterFixture(t *testing.T) {
	c, srv := newClient(t)
	result := fixtureResult(t, "local-access.get_config.json")
	srv.Handle("local-access", "get_config", func(json.RawMessage) (any, error) { return result, nil })

	res, err := c.LocalAccess.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	checkLossless(t, res, result)
}
