package glinet_test

import (
	"bufio"
	"context"
	"os"
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
