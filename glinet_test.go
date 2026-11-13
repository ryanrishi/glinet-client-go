package glinet_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

// recorder sits in front of a fake router and records each JSON-RPC method.
// For "call" it records "call:module.method".
type recorder struct {
	mu      sync.Mutex
	methods []string
}

func (r *recorder) record(req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewReader(body))

	var rpc struct {
		Method string            `json:"method"`
		Params []json.RawMessage `json:"params"`
	}
	_ = json.Unmarshal(body, &rpc)

	name := rpc.Method
	if rpc.Method == "call" && len(rpc.Params) >= 3 {
		var module, method string
		_ = json.Unmarshal(rpc.Params[1], &module)
		_ = json.Unmarshal(rpc.Params[2], &method)
		name += ":" + module + "." + method
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.methods = append(r.methods, name)
}

func (r *recorder) Methods() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.methods)
}

func (r *recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.methods = nil
}

func newRecordedClient(t *testing.T) (*glinet.Client, *glinettest.Server, *recorder) {
	t.Helper()
	srv := glinettest.NewServer(t)
	target, _ := url.Parse(srv.URL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	rec := &recorder{}
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.record(r)
		proxy.ServeHTTP(w, r)
	}))
	t.Cleanup(front.Close)

	c, err := glinet.NewClientWithHost(front.URL, srv.Username, []byte(srv.Password))
	if err != nil {
		t.Fatalf("NewClientWithHost: %v", err)
	}
	return c, srv, rec
}

func newClient(t *testing.T) (*glinet.Client, *glinettest.Server) {
	t.Helper()
	srv := glinettest.NewServer(t)
	c, err := glinet.NewClientWithHost(srv.URL, srv.Username, []byte(srv.Password))
	if err != nil {
		t.Fatalf("NewClientWithHost: %v", err)
	}
	return c, srv
}

func rpcCode(t *testing.T, err error) int {
	t.Helper()
	var rpcErr *glinet.RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("error %v is not an *RPCError", err)
	}
	return rpcErr.Code
}

func TestLogin(t *testing.T) {
	c, _, rec := newRecordedClient(t)

	if c.GetSid() == "" {
		t.Error("session ID is empty")
	}
	if got, want := rec.Methods(), []string{"challenge", "login"}; !slices.Equal(got, want) {
		t.Errorf("methods = %v, want %v", got, want)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	srv := glinettest.NewServer(t)

	_, err := glinet.NewClientWithHost(srv.URL, srv.Username, []byte("wrong"))
	if err == nil {
		t.Fatal("expected an error")
	}
	if code := rpcCode(t, err); code != glinet.CodeAccessDenied {
		t.Errorf("code = %d, want %d", code, glinet.CodeAccessDenied)
	}
}

func TestLoginUnsupportedAlgorithm(t *testing.T) {
	tests := map[string]string{
		"alg 5":       `{"alg":5,"salt":"abcdefgh","nonce":"n"}`,
		"hash_method": `{"alg":1,"salt":"abcdefgh","nonce":"n","hash_method":"sha256"}`,
	}

	for name, challenge := range tests {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(body), `"method":"challenge"`) {
					t.Errorf("unexpected request %s", body)
				}
				_, _ = io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":`+challenge+`}`)
			}))
			t.Cleanup(srv.Close)

			_, err := glinet.NewClientWithHost(srv.URL, "root", []byte("password"))
			if err == nil || !strings.Contains(err.Error(), "unsupported login algorithm") {
				t.Errorf("error = %v, want unsupported login algorithm", err)
			}
		})
	}
}

func TestSessionRenewal(t *testing.T) {
	c, srv, rec := newRecordedClient(t)
	oldSid := c.GetSid()
	rec.Reset()
	srv.ExpireSessions()

	if _, err := c.System.GetTimezoneConfig(context.Background()); err != nil {
		t.Fatalf("GetTimezoneConfig: %v", err)
	}

	want := []string{"call:system.get_timezone_config", "challenge", "login", "call:system.get_timezone_config"}
	if got := rec.Methods(); !slices.Equal(got, want) {
		t.Errorf("methods = %v, want %v", got, want)
	}
	if c.GetSid() == oldSid {
		t.Error("session ID did not change")
	}
}

func TestSessionRenewalAfterCallerClearsPassword(t *testing.T) {
	srv := glinettest.NewServer(t)
	password := []byte(srv.Password)
	c, err := glinet.NewClientWithHost(srv.URL, srv.Username, password)
	if err != nil {
		t.Fatalf("NewClientWithHost: %v", err)
	}
	clear(password)
	srv.ExpireSessions()

	if _, err := c.System.GetTimezoneConfig(context.Background()); err != nil {
		t.Fatalf("GetTimezoneConfig: %v", err)
	}
}

func TestSessionRenewalOnlyOnce(t *testing.T) {
	c, srv, rec := newRecordedClient(t)
	srv.Handle("system", "get_timezone_config", func(json.RawMessage) (any, error) {
		return nil, glinettest.ErrAccessDenied
	})
	rec.Reset()

	_, err := c.System.GetTimezoneConfig(context.Background())
	if code := rpcCode(t, err); code != glinet.CodeAccessDenied {
		t.Errorf("code = %d, want %d", code, glinet.CodeAccessDenied)
	}

	want := []string{"call:system.get_timezone_config", "challenge", "login", "call:system.get_timezone_config"}
	if got := rec.Methods(); !slices.Equal(got, want) {
		t.Errorf("methods = %v, want %v", got, want)
	}
}

func TestCallMethodNotFound(t *testing.T) {
	c, _ := newClient(t)

	err := c.Call(context.Background(), "system", "no_such_method", nil, nil)
	if code := rpcCode(t, err); code != glinet.CodeMethodNotFound {
		t.Errorf("code = %d, want %d", code, glinet.CodeMethodNotFound)
	}
}

func TestCallParams(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("test", "echo", func(params json.RawMessage) (any, error) {
		return params, nil
	})

	tests := []struct {
		params any
		want   string
	}{
		{nil, `{}`},
		{map[string]int{"a": 1}, `{"a":1}`},
	}

	for _, tt := range tests {
		var res json.RawMessage
		if err := c.Call(context.Background(), "test", "echo", tt.params, &res); err != nil {
			t.Fatalf("Call: %v", err)
		}
		if string(res) != tt.want {
			t.Errorf("params %v: router got %s, want %s", tt.params, res, tt.want)
		}
	}
}

func TestSystemGetStatus(t *testing.T) {
	c, _ := newClient(t)

	res, err := c.System.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if res.System.LanIp != "192.168.8.1" || res.System.Uptime != 1234.56 || len(res.Wifi) != 2 {
		t.Errorf("unexpected status %+v", res)
	}
}

func TestSystemTimezoneConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	err := c.System.SetTimezoneConfig(ctx, &glinet.SetSystemTimezoneRequest{
		Zonename:  "Europe/Berlin",
		Timezone:  "CET-1CEST,M3.5.0,M10.5.0/3",
		Localtime: 1700000100,
	})
	if err != nil {
		t.Fatalf("SetTimezoneConfig: %v", err)
	}

	state := srv.System.TimezoneConfig()
	if state.Zonename != "Europe/Berlin" || state.Timezone != "CET-1CEST,M3.5.0,M10.5.0/3" || state.Localtime != 1700000100 {
		t.Errorf("server state = %+v", state)
	}

	res, err := c.System.GetTimezoneConfig(ctx)
	if err != nil {
		t.Fatalf("GetTimezoneConfig: %v", err)
	}
	want := glinet.GetSystemTimezoneConfigResponse{
		Zonename:            "Europe/Berlin",
		TZOffset:            state.TZOffset,
		AutoTimezoneEnabled: true,
		Localtime:           1700000100,
		Timezone:            "CET-1CEST,M3.5.0,M10.5.0/3",
	}
	if *res != want {
		t.Errorf("GetTimezoneConfig = %+v, want %+v", *res, want)
	}
}

func TestAdGuardConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.AdGuard.GetAdGuardConfig(ctx)
	if err != nil {
		t.Fatalf("GetAdGuardConfig: %v", err)
	}
	if res.Enabled || !res.DNSEnabled {
		t.Errorf("GetAdGuardConfig = %+v, want enabled=false dns_enabled=true", *res)
	}

	if err := c.AdGuard.SetAdGuardConfig(ctx, &glinet.SetAdGuardConfigRequest{Enabled: true, DNSEnabled: false}); err != nil {
		t.Fatalf("SetAdGuardConfig: %v", err)
	}
	if got := srv.AdGuardHome.Config(); !got.Enabled || got.DNSEnabled {
		t.Errorf("server state = %+v, want enabled=true dns_enabled=false", got)
	}
}
