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

func TestServerHandlerWrapsMethod(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	original := srv.Handler("system", "get_timezone_config")
	calls := 0
	srv.Handle("system", "get_timezone_config", func(params json.RawMessage) (any, error) {
		calls++
		return original(params)
	})

	res, err := c.System.GetTimezoneConfig(ctx)
	if err != nil {
		t.Fatalf("GetTimezoneConfig: %v", err)
	}
	if calls != 1 || res.Zonename != srv.System.TimezoneConfig().Zonename {
		t.Errorf("calls = %d, result %+v, want 1 call and the original result", calls, *res)
	}

	if m := srv.Handler("system", "no_such_method"); m != nil {
		t.Error("Handler of an unknown method is not nil")
	}
}

func TestServerInternalServerError(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("test", "fail", func(json.RawMessage) (any, error) {
		return nil, glinettest.ErrInternalServerError
	})

	err := c.Call(context.Background(), "test", "fail", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %v, want HTTP status 500", err)
	}
}

func TestSystemReadsDecodeRouterFixtures(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	info := fixtureResult(t, "system.get_info.json")
	srv.System.SetInfo(info)
	gotInfo, err := c.System.GetInfo(ctx)
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	checkLossless(t, gotInfo, info)

	status := fixtureResult(t, "system.get_status.json")
	srv.System.SetStatus(status)
	gotStatus, err := c.System.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	checkLossless(t, gotStatus, status)

	disk := fixtureResult(t, "system.disk_info.json")
	srv.System.SetDiskInfo(disk)
	gotDisk, err := c.System.GetDiskInfo(ctx)
	if err != nil {
		t.Fatalf("GetDiskInfo: %v", err)
	}
	checkLossless(t, gotDisk, disk)
}

func TestSystemInfo(t *testing.T) {
	c, srv := newClient(t)
	srv.System.SetInfo(json.RawMessage(`{"model": "mt6000", "firmware_version": "4.8.1", "sn": "SN1", "hidden_features": ["x"]}`))

	res, err := c.System.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo: %v", err)
	}
	if res.Model != "mt6000" || res.FirmwareVersion != "4.8.1" || res.SN != "SN1" || !slices.Equal(res.HiddenFeatures, []string{"x"}) {
		t.Errorf("GetInfo = %+v", *res)
	}
}

func TestSystemTimezoneConfig(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	initial := srv.System.TimezoneConfig()

	err := c.System.SetTimezoneConfig(ctx, &glinet.SetSystemTimezoneRequest{
		Zonename: "Europe/Berlin",
		Timezone: "CET-1CEST,M3.5.0,M10.5.0/3",
	})
	if err != nil {
		t.Fatalf("SetTimezoneConfig: %v", err)
	}

	state := srv.System.TimezoneConfig()
	if state.Zonename != "Europe/Berlin" || state.Timezone != "CET-1CEST,M3.5.0,M10.5.0/3" || state.Timestamp != initial.Timestamp {
		t.Errorf("server state = %+v, want new zone and timestamp %d", state, initial.Timestamp)
	}

	res, err := c.System.GetTimezoneConfig(ctx)
	if err != nil {
		t.Fatalf("GetTimezoneConfig: %v", err)
	}
	want := glinet.GetSystemTimezoneConfigResponse{
		Zonename:            "Europe/Berlin",
		TZOffset:            "+0100",
		AutoTimezoneEnabled: true,
		Localtime:           int(initial.Localtime),
		Timestamp:           int(initial.Timestamp),
		Timezone:            "CET-1CEST,M3.5.0,M10.5.0/3",
	}
	if *res != want {
		t.Errorf("GetTimezoneConfig = %+v, want %+v", *res, want)
	}
}

func TestSystemTimezoneConfigLocaltime(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	localtime := int64(1700000100)
	err := c.System.SetTimezoneConfig(ctx, &glinet.SetSystemTimezoneRequest{
		Zonename:  "UTC",
		Timezone:  "UTC0",
		Localtime: &localtime,
	})
	if err != nil {
		t.Fatalf("SetTimezoneConfig: %v", err)
	}

	if got := srv.System.TimezoneConfig().Timestamp; got != localtime {
		t.Errorf("server timestamp = %d, want %d", got, localtime)
	}
}

func TestSystemTimezoneConfigIgnoresAutoTimezone(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	params := map[string]any{"zonename": "UTC", "timezone": "UTC0", "autotimezone_enabled": false}
	if err := c.Call(ctx, "system", "set_timezone_config", params, nil); err != nil {
		t.Fatalf("set_timezone_config: %v", err)
	}

	if !srv.System.TimezoneConfig().AutoTimezoneEnabled {
		t.Error("autotimezone_enabled changed, want it unchanged")
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

func TestCallResultError(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("test", "result_error", func(json.RawMessage) (any, error) {
		return map[string]any{"err_code": -1, "err_msg": "parameter missing"}, nil
	})

	err := c.Call(context.Background(), "test", "result_error", nil, nil)
	var rpcErr *glinet.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != -1 || rpcErr.Message != "parameter missing" {
		t.Errorf("error = %v, want *RPCError -1 parameter missing", err)
	}
}

func TestSystemStatusDecodesLargeByteCounts(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	const big = 8589934592
	srv.System.SetStatus(json.RawMessage(`{"system": {"flash_app": 8589934592, "flash_total": 8589934592, "flash_free": 8589934592, "memory_total": 8589934592, "memory_free": 8589934592, "memory_buff_cache": 8589934592, "timestamp": 8589934592}}`))
	srv.System.SetDiskInfo(json.RawMessage(`{"root": {"total": 8589934592, "used": 8589934592, "free": 8589934592}, "tmp": {"total": 8589934592, "used": 0, "free": 8589934592}}`))

	status, err := c.System.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	s := status.System
	for name, v := range map[string]int64{"flash_app": s.FlashApp, "flash_total": s.FlashTotal, "flash_free": s.FlashFree, "memory_total": s.MemoryTotal, "memory_free": s.MemoryFree, "memory_buff_cache": s.MemoryBuffCache, "timestamp": s.Timestamp} {
		if v != big {
			t.Errorf("%s = %d, want %d", name, v, int64(big))
		}
	}

	disk, err := c.System.GetDiskInfo(ctx)
	if err != nil {
		t.Fatalf("GetDiskInfo: %v", err)
	}
	if disk.Root.Total != big || disk.Tmp.Free != big {
		t.Errorf("GetDiskInfo = %+v, want %d bytes", *disk, int64(big))
	}
}
