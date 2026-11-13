package glinet_test

import (
	"bufio"
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

// TestRouterReadOnly calls read methods on a real router.
// It runs only when GLINET_ROUTER_TEST=1. GLINET_HOST, GLINET_USERNAME and GLINET_PASSWORD
// come from the environment or from the .env file.
func TestRouterReadOnly(t *testing.T) {
	if os.Getenv("GLINET_ROUTER_TEST") != "1" {
		t.Skip("set GLINET_ROUTER_TEST=1 to run against a real router")
	}

	c := newRouterClient(t)
	ctx := context.Background()

	tz, err := c.System.GetTimezoneConfig(ctx)
	if err != nil {
		t.Fatalf("system.get_timezone_config: %v", err)
	}
	if tz.Zonename == "" || tz.Timezone == "" {
		t.Errorf("system.get_timezone_config: empty zonename or timezone: %+v", *tz)
	}
	t.Logf("system.get_timezone_config: %+v", *tz)

	ag, err := c.AdGuard.GetAdGuardConfig(ctx)
	if err != nil {
		t.Fatalf("adguardhome.get_config: %v", err)
	}
	t.Logf("adguardhome.get_config: %+v", *ag)
}

// TestRouterStaticBind adds, updates and removes one static bind with an invented MAC on a real router.
// It runs only when GLINET_ROUTER_WRITE_TEST=1. It does not change other bindings.
func TestRouterStaticBind(t *testing.T) {
	if os.Getenv("GLINET_ROUTER_WRITE_TEST") != "1" {
		t.Skip("set GLINET_ROUTER_WRITE_TEST=1 to write to a real router")
	}

	const mac = "02:00:00:aa:bb:01"
	c := newRouterClient(t)
	ctx := context.Background()

	initial, err := c.LAN.ListStaticBinds(ctx)
	if err != nil {
		t.Fatalf("ListStaticBinds: %v", err)
	}
	for _, b := range initial {
		if strings.EqualFold(b.MAC, mac) || b.IP == "192.168.8.50" || b.IP == "192.168.8.51" {
			t.Fatalf("router already has a binding with the test MAC or IP: %+v", b)
		}
	}

	t.Cleanup(func() {
		if err := c.LAN.RemoveStaticBind(context.Background(), mac); err != nil {
			t.Errorf("cleanup RemoveStaticBind: %v", err)
		}
		final, err := c.LAN.ListStaticBinds(context.Background())
		if err != nil {
			t.Fatalf("cleanup ListStaticBinds: %v", err)
		}
		if !slices.Equal(final, initial) {
			t.Errorf("static binds after test = %+v, want %+v", final, initial)
		}
	})

	if err := c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: mac, IP: "192.168.8.50", Name: "test-lease"}); err != nil {
		t.Fatalf("AddStaticBind: %v", err)
	}
	got, err := c.LAN.FindStaticBind(ctx, strings.ToUpper(mac))
	if err != nil {
		t.Fatalf("FindStaticBind: %v", err)
	}
	if want := (glinet.StaticBind{MAC: mac, IP: "192.168.8.50", Name: "test-lease"}); got == nil || *got != want {
		t.Fatalf("FindStaticBind after add = %+v, want %+v", got, want)
	}

	if err := c.LAN.SetStaticBind(ctx, glinet.StaticBind{MAC: mac, IP: "192.168.8.51"}); err != nil {
		t.Fatalf("SetStaticBind: %v", err)
	}
	got, err = c.LAN.FindStaticBind(ctx, mac)
	if err != nil {
		t.Fatalf("FindStaticBind: %v", err)
	}
	if want := (glinet.StaticBind{MAC: mac, IP: "192.168.8.51"}); got == nil || *got != want {
		t.Fatalf("FindStaticBind after set = %+v, want %+v", got, want)
	}

	err = c.LAN.AddStaticBind(ctx, glinet.StaticBind{MAC: "02-00-00-aa-bb-01", IP: "192.168.8.50"})
	var rpcErr *glinet.RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != glinet.CodeInvalidParams || rpcErr.Err != "invalid mac" {
		t.Errorf("AddStaticBind invalid MAC error = %v, want -32602 invalid mac", err)
	}

	if err := c.LAN.RemoveStaticBind(ctx, mac); err != nil {
		t.Fatalf("RemoveStaticBind: %v", err)
	}
	got, err = c.LAN.FindStaticBind(ctx, mac)
	if err != nil || got != nil {
		t.Errorf("FindStaticBind after remove = %+v, %v, want nil, nil", got, err)
	}
}

func newRouterClient(t *testing.T) *glinet.Client {
	t.Helper()
	env := readDotEnv(t)
	get := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		if v := env[key]; v != "" {
			return v
		}
		return fallback
	}

	c, err := glinet.NewClientWithHost(get("GLINET_HOST", "http://192.168.8.1"), get("GLINET_USERNAME", "root"), []byte(get("GLINET_PASSWORD", "")))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return c
}

func readDotEnv(t *testing.T) map[string]string {
	t.Helper()
	env := map[string]string{}

	f, err := os.Open(".env")
	if os.IsNotExist(err) {
		return env
	}
	if err != nil {
		t.Fatalf("open .env: %v", err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		key, value, ok := strings.Cut(strings.TrimSpace(s.Text()), "=")
		if !ok || strings.HasPrefix(key, "#") {
			continue
		}
		env[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("read .env: %v", err)
	}

	return env
}
