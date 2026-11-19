package glinet_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
)

// TestReadsDecodeFirmware483Fixtures checks that the typed reads that firmware 4.8.3 still has
// keep each value of its fixtures.
func TestReadsDecodeFirmware483Fixtures(t *testing.T) {
	type read func(context.Context, *glinet.Client) (any, error)
	list := func(key string, f func(context.Context, *glinet.Client) (any, error)) read {
		return func(ctx context.Context, c *glinet.Client) (any, error) {
			v, err := f(ctx, c)
			return map[string]any{key: v}, err
		}
	}
	tests := []struct {
		method  string
		read    read
		omitted []string
	}{
		{"system.get_info", func(ctx context.Context, c *glinet.Client) (any, error) { return c.System.GetInfo(ctx) }, nil},
		{"system.get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.System.GetStatus(ctx) }, nil},
		{"system.disk_info", func(ctx context.Context, c *glinet.Client) (any, error) { return c.System.GetDiskInfo(ctx) }, nil},
		{"system.get_timezone_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.System.GetTimezoneConfig(ctx) }, nil},
		{"adguardhome.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.AdGuard.GetAdGuardConfig(ctx) }, nil},
		{"black_white_list.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.AccessControl.GetConfig(ctx) }, nil},
		{"cable.get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Cable.GetStatus(ctx) }, nil},
		{"cable.get_config", func(ctx context.Context, c *glinet.Client) (any, error) {
			return c.Cable.GetConfig(ctx, glinet.FirstWANPort)
		}, nil},
		{"clients.get_list", list("clients", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Clients.ListClients(ctx) }), nil},
		{"ddns.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.DDNS.GetConfig(ctx) }, nil},
		{"ddns.get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.DDNS.GetStatus(ctx) }, nil},
		{"dns.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.DNS.GetConfig(ctx) }, nil},
		{"dns.get_info", func(ctx context.Context, c *glinet.Client) (any, error) { return c.DNS.GetInfo(ctx) }, nil},
		{"fan.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Fan.GetConfig(ctx) }, nil},
		{"fan.get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Fan.GetStatus(ctx) }, nil},
		{"firewall.get_dmz", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Firewall.GetDMZ(ctx) }, nil},
		{"firewall.get_wan_access", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Firewall.GetWANAccess(ctx) }, nil},
		{"firewall.get_zone_list", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Firewall.GetZones(ctx) }, nil},
		{"firewall.get_port_forward_list", list("res", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Firewall.ListPortForwards(ctx) }), nil},
		{"firewall.get_rule_list", list("res", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Firewall.ListRules(ctx) }), nil},
		{"igmp.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.IGMP.GetConfig(ctx) }, nil},
		{"ipv6.get_ipv6", func(ctx context.Context, c *glinet.Client) (any, error) { return c.IPv6.GetConfig(ctx) }, nil},
		{"lan.get_config_list", list("interfaces", func(ctx context.Context, c *glinet.Client) (any, error) { return c.LAN.ListConfigs(ctx) }), nil},
		{"lan.get_static_bind_list", list("static_bind_list", func(ctx context.Context, c *glinet.Client) (any, error) { return c.LAN.ListStaticBinds(ctx) }), nil},
		{"lan.get_wan_info", list("wan_info", func(ctx context.Context, c *glinet.Client) (any, error) { return c.LAN.GetWANInfo(ctx) }), nil},
		{"led.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.LED.GetConfig(ctx) }, nil},
		{"local-access.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.LocalAccess.GetConfig(ctx) }, nil},
		{"kmwan.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.MultiWAN.GetConfig(ctx) }, nil},
		{"kmwan.get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.MultiWAN.GetStatus(ctx) }, nil},
		{"network.check_wan_cable", func(ctx context.Context, c *glinet.Client) (any, error) {
			return c.Network.CheckWANCable(ctx, glinet.FirstWANPort)
		}, nil},
		{"netmode.get_mode", func(ctx context.Context, c *glinet.Client) (any, error) { return c.NetworkMode.GetMode(ctx) }, nil},
		{"network.get_arp_list", list("entries", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Network.ListARPEntries(ctx) }), nil},
		{"switch-button.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.SwitchButton.GetConfig(ctx) }, nil},
		{"switch-button.get_funcs", list("funcs", func(ctx context.Context, c *glinet.Client) (any, error) { return c.SwitchButton.GetFuncs(ctx) }), nil},
		{"tailscale.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Tailscale.GetConfig(ctx) }, nil},
		{"tailscale.get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Tailscale.GetStatus(ctx) }, nil},
		{"timer.get_led", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Timer.GetLEDSchedule(ctx) }, nil},
		{"timer.get_reboot", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Timer.GetRebootSchedule(ctx) }, nil},
		{"wifi.get_status", list("res", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Wifi.GetStatus(ctx) }), nil},
		{"wifi.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.Wifi.GetConfig(ctx) }, []string{"bandmode"}},
		{"zerotier.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.ZeroTier.GetConfig(ctx) }, nil},
		{"wg-client.get_group_list", list("groups", func(ctx context.Context, c *glinet.Client) (any, error) { return c.WireGuardClient.ListGroups(ctx) }), nil},
		{"wg-client.get_all_config_list", list("config_list", func(ctx context.Context, c *glinet.Client) (any, error) { return c.WireGuardClient.ListAllPeers(ctx) }), nil},
		{"wg-client.get_config_list", list("peers", func(ctx context.Context, c *glinet.Client) (any, error) {
			return c.WireGuardClient.ListPeers(ctx, 9859)
		}), []string{"auth_type"}},
		{"wg-server.get_config", func(ctx context.Context, c *glinet.Client) (any, error) { return c.WireGuardServer.GetConfig(ctx) }, nil},
		{"wg-server.get_status", func(ctx context.Context, c *glinet.Client) (any, error) { return c.WireGuardServer.GetStatus(ctx) }, nil},
		{"wg-server.get_peer_list", list("peers", func(ctx context.Context, c *glinet.Client) (any, error) { return c.WireGuardServer.ListPeers(ctx) }), nil},
		{"wg-server.get_setting", func(ctx context.Context, c *glinet.Client) (any, error) { return c.WireGuardServer.GetSetting(ctx) }, nil},
		{"wg-server.get_route_list", func(ctx context.Context, c *glinet.Client) (any, error) { return c.WireGuardServer.ListRoutes(ctx) }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			c, srv := newClient(t)
			result := firmwareFixtureResult(t, "4.8.3", tt.method+".json")
			module, method, _ := strings.Cut(tt.method, ".")
			srv.Handle(module, method, func(json.RawMessage) (any, error) { return result, nil })

			got, err := tt.read(context.Background(), c)
			if err != nil {
				t.Fatal(err)
			}
			checkLossless(t, got, result, tt.omitted...)
		})
	}
}
