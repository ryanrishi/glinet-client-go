# Probe report: GL.iNet AXT1800, firmware 4.7.0

- Model: `axt1800` (GL-AXT1800, "Slate AX"), board "GL Technologies, Inc. AXT1800"
- Firmware: `4.7.0`, type `release8`, date 2025-03-19, OpenWrt 23.05-SNAPSHOT, kernel 5.4.164
- Probe date: 2026-11-12. Program: `cmd/probe` (read-only allowlist in `cmd/probe/calls.go`).
- Result: 173 read methods in the allowlist. 126 ok, 37 error, 10 skipped (no ID from a previous list response).

## Login

`challenge` returns `{"alg":1,"salt":...,"nonce":...}`. There is no `hash_method`, so the hash is MD5. `alg` 1 = MD5-crypt. `login` returns `{"sid":...,"username":"root"}`. See `challenge.json` and `login.json`.

## Error format

All errors are a JSON-RPC `error` object. No response had `err_code` inside `result`.

- Unknown method: `{"error":{"code":-32601,"message":"Method not found"}}` (37 calls).
- Invalid or expired sid: `{"error":{"code":-32000,"message":"Access denied"}}`. See `invalid_sid.system.get_info.json`.
- Invalid input to a write method: `{"error":{"code":-32602,"message":"Invalid params","err":"invalid mac"}}`. The `err` field has the detail. See `lan.*_static_bind.invalid_*.json`.

## VPN modules

| Module | Result |
|---|---|
| `wg-client` | Exists. All called methods ok (`get_status`, `get_group_list`, `get_all_config_list`, `get_config_list`, `get_route_list`, `get_setting`). |
| `ovpn-client` | Exists. `get_status`, `get_group_list`, `get_all_config_list`, `get_config_list` ok. No group has clients, so `get_route_list` and `get_setting` were skipped. |
| `wg-server` | Exists. All 5 read methods ok. |
| `ovpn-server` | Exists. All 5 read methods ok. |
| `vpn-client` | Does not exist on 4.7.0. All 4 methods return -32601. |

The 4.9.0-only methods also return -32601 on 4.7.0: `system.get_network_status`, `system.get_timezone_list`, `firewall.get_acl_rule_list`, `firewall.get_acl_zone_list`, all `vpn-client.*`. Only `lan.get_wan_info` is ok.

## Differences from the 2023 reference

Many lists were empty on this router, so their item fields could not be compared: `firewall.get_port_forward_list.res`, `firewall.get_rule_list.res`, `firewall.get_wan_access.whitelist`, `wg-server.get_peer_list.peers`, `wg-server.get_route_list.*_route_rules`, `wg-server.get_status.peers`. DMZ is disabled, so `firewall.get_dmz` has only `enabled` (no `dest_ip`). Later phases must get fixtures with data for these (or use the reference shapes and mark them as not verified).

The reference marks `err_code` / `err_msg` as output of many methods. No ok response had them.

### system
- `get_info`: no `disable_guest_during_scan_wifi`. Extra: `device_type`, `hardware_feature.{bluetooth,gps,noled,rs485,slot,submodel,usb3}`, `software_feature.{cellular_upgrade,ksmbd,mlo}`.
- `get_status`: extra `system.cpu`, `wifi[].encryption`, `wifi[].hidden`, `wifi[].mld`. No `system.mcu` and no `service.client_id` (optional in reference).
- `get_timezone_config`: field is `autotimezone_enabled`, not `auto_timezone_enabled`. Extra `timestamp`.
- `get_httpd_mem_status`: only `memory_used`.
- `disk_info`, `get_load`, `get_security_policy`, `get_unixtime`: match.

### lan
- `get_config_list`: shape is `interfaces[]` with `interface`, `ip`, `gateway`, `netmask`, `start`, `end`, `enable`, `dns`, `leasetime`, `lpr`, `ap_isolate`; `transfer_enable` only on `guest`. This agrees with the intent of the reference, but the reference writes some of these paths wrong (`interface.enable`, top-level `ap_isolate`). `enable`, `ap_isolate`, `transfer_enable` are numbers (0/1), not bool. `leasetime` is a string (`"12h"`).
- `get_static_bind_list`: `{"static_bind_list":[{"mac":"02:00:00:aa:bb:01","ip":"192.168.8.50","name":"test-lease"}]}`. All fields are strings. A binding without a name has no `name` key.

### lan static bind (write test, 2026-11-13)

Verified on the router with invented MACs `02:00:00:aa:bb:0x` and IPs outside the DHCP range (`192.168.8.100`-`249`). The list was empty before and after the test. The fixtures `lan.*static_bind*.json` are one sequence; sorted by request `id`, they replay the state changes in order (`glinettest/lan_test.go` does this).

| Call | Params | Result |
|---|---|---|
| `add_static_bind` | `{mac, ip, name?}` | `[]` |
| `set_static_bind` | `{mac, ip, name?}` | `[]` |
| `remove_static_bind` | `{mode: 0, mac}` | `[]` |

Behavior:
- The router identifies a binding by its exact MAC string. `02:00:00:AA:BB:01` and `02:00:00:aa:bb:01` can both exist, and `set` and `remove` with one letter case do not change the other. Exception: the two bindings showed the same `name` after a write to one of them (not in the fixtures; not emulated by `glinettest`). The router uses lower case in `clients.get_list`.
- `add` with an existing MAC replaces that binding. It is not an error.
- `add`, and `set` for a known MAC, remove each other binding with the same IP. The written binding goes to the end of the list.
- `set` without `name` removes the name. `name: ""` is an error on `add` and `set`.
- `set` and `remove` with an unknown MAC return `[]` and change nothing.
- An IP outside the LAN subnet (`10.0.0.5`) and the router IP (`192.168.8.1`) are accepted.
- Names with spaces, `_`, `.`, non-ASCII letters, and 70 characters are accepted.

Errors (all code `-32602`, message `Invalid params`):

| `err` | Cause |
|---|---|
| `invalid mac` | MAC with 5 groups, `-` or `.` separators, a non-hex digit, missing `mac`. Also on `set` and `remove`. |
| `invalid ip` | `192.168.8.256`, `192.168.8`, `not-an-ip`, `""`, missing `ip`. Also on `set`. |
| `invalid name` | `name: ""` on `add` and `set`. |
| `invalid mode` | `remove_static_bind` without `mode`. |

Differences from the 2023 reference:
- `set_static_bind`: `name` is optional, and no `name` removes the name. The reference lists `name` for `set` without marking it optional.
- The reference does not say that `add` replaces a binding with the same MAC or IP, or that `set` and `remove` with an unknown MAC are not errors.
- The error detail is in `error.err`, not in `err_code` / `err_msg`.

Not verified: `remove_static_bind` mode 1 (remove all; not called), other `mode` values, IPv6 addresses, `set` with an unknown MAC and an IP that another binding has, the list position after `add` with an existing MAC, and which check comes first when more than one field is invalid.

The Ethernet link of the test Mac (`en8`) lost its IPv4 address two times for about 25 s during the test sessions. One time this happened before any write of that run. One add and remove with a ping log for 240 s did not cause it. The cause is not known.
- `get_wan_info` (not in reference): ok.

### firewall
- `get_zone_list`: matches.
- `get_dmz`, `get_port_forward_list`, `get_rule_list`, `get_wan_access`: no data to compare (see above). Top-level fields match.

### dns
- `get_config`: no `proto`. Extra: `controld_id`, `dot_provider` (string, `"0"`), `free_controld_id`, `override_vpn`. `server` items are strings like `"wwan 192.168.4.253"`.
- `get_host`, `get_info`: match.

### adguardhome
- `get_config`: extra `dns_enabled` (reference has it only as `set_config` input).

### wifi
- `get_config`: extra `res[].random_bssid`. `res[].channels` items are objects `{channel, dfs, psc}`. `res[].htmodes` is an object (mode → max bandwidth, plus `auto`), as the reference says.
- `get_status`: matches.

### clients
- `get_list`: extra per client: `ipv6`, `last_rx`, `last_tx`, `last_update_rate`, `total_rx_init`, `total_tx_init`.
- `get_status`: matches.

### wg-server
- `get_status`: server fields are nested in `server{initialization,log,rx_bytes,status,tunnel_ip,tx_bytes}`. The reference has them at top level. `peers` is a top-level list (empty here).
- `get_config`, `get_setting`: match.
- `get_peer_list`, `get_route_list`: empty lists; item fields not verified.

### network
- `routes`, `routes6`, `get_dhcp_leases`, `get_hwnat_config`: -32601 Method not found.
- `check_wan_cable`: returns `cable_enabled`, `cable_inserted`, `macclone_enabled`.
- `get_netnat_config`: extra `conflict_features`, `qos_enabled`, `wifi_reload`.
- `get_arp_list`: matches.

## Methods not called

Not read-only or not sure: all `set_*`, `add_*`, `remove_*`, etc. Also these `get_*` methods: `upgrade.check_firmware_online`, `tailscale.get_auth_url`, `wg-client.get_third_config`, `ovpn-client.get_third_config` (send credentials to a third-party VPN service), `wg-client.get_recommend_config`, `ovpn-client.get_recommend_config` (query an external service).

## Redaction

Replaced with stable fake values (same real value → same fake value, type kept): password, cipher password, sid, salt, nonce, login hash, Wi-Fi keys, private keys, preshared keys, tokens, `sn`, `sn_bak`, `ddns` / `device_id`, MAC addresses (`02:00:00:00:xx:xx`), public IPv4 (`203.0.113.x`), global/ULA/link-local IPv6 (`2001:db8::x`, `fd00::x`, `fe80::x`), WireGuard endpoint hostnames (`host-N.example.com`), SSIDs (`ssid-N`, also the SSIDs of other networks in `repeater.*`), hostnames and client names (`hostname`, `alias`, and `name` in objects that have a `mac`: `hostname-N`), WireGuard public keys (base64 of 32 bytes from SHA-256 of a counter; empty values stay empty). A fake value replaces the real value also inside other strings, such as log lines, when the real value has 4 or more characters. The number N depends on the order of the probe run, so it can change between runs.

Kept: private IPv4, well-known public DNS resolvers, VPN group names, VPN peer names and VPN usernames (`wg-client`, `ovpn-client`), time zone.

## Results

| Method | Result | Detail |
|---|---|---|
| `acl.get_acl_list` | skipped | no ID from a previous list response |
| `acl.get_group_list` | ok | |
| `adguardhome.get_config` | ok | |
| `bark.get_config` | error | -32601 Method not found |
| `bark.get_status` | error | -32601 Method not found |
| `black_white_list.get_config` | ok | |
| `cable.get_config` | ok | |
| `cable.get_status` | ok | |
| `clients.get_list` | ok | |
| `clients.get_status` | ok | |
| `cloud-batch-manage.get_2b_config` | error | -32601 Method not found |
| `cloud-batch-manage.get_batch_config` | error | -32601 Method not found |
| `cloud.get_batch_config` | ok | |
| `cloud.get_config` | ok | |
| `ddns.get_config` | ok | |
| `ddns.get_status` | ok | |
| `dlna.get_config` | error | -32601 Method not found |
| `dns.get_config` | ok | |
| `dns.get_host` | ok | |
| `dns.get_info` | ok | |
| `edgerouter.get_config` | ok | |
| `edgerouter.get_status` | ok | |
| `fan.get_config` | ok | |
| `fan.get_status` | ok | |
| `firewall.get_acl_rule_list` | error | -32601 Method not found |
| `firewall.get_acl_zone_list` | error | -32601 Method not found |
| `firewall.get_dmz` | ok | |
| `firewall.get_port_forward_list` | ok | |
| `firewall.get_rule_list` | ok | |
| `firewall.get_wan_access` | ok | |
| `firewall.get_zone_list` | ok | |
| `igmp.get_config` | ok | |
| `ipv6.get_ipv6` | ok | |
| `kmwan.get_config` | ok | |
| `kmwan.get_status` | ok | |
| `lan.get_config_list` | ok | |
| `lan.get_static_bind_list` | ok | |
| `lan.get_wan_info` | ok | |
| `led.get_config` | ok | |
| `local-access.get_config` | ok | |
| `logread.get_config` | error | -32601 Method not found |
| `logread.get_crash_log` | ok | |
| `logread.get_kernel_log` | ok | |
| `logread.get_nginx_log` | ok | |
| `logread.get_system_log` | ok | |
| `macclone.get_mac` | error | -32601 Method not found |
| `mcu.get_battery_config` | error | -32601 Method not found |
| `mcu.get_oled_config` | error | -32601 Method not found |
| `modem.get_cell_tower` | ok | |
| `modem.get_cells_info` | skipped | no ID from a previous list response |
| `modem.get_config` | skipped | no ID from a previous list response |
| `modem.get_debug_msg` | skipped | no ID from a previous list response |
| `modem.get_info` | ok | |
| `modem.get_name` | ok | |
| `modem.get_profile_list` | ok | |
| `modem.get_signals` | ok | |
| `modem.get_sim_info` | ok | |
| `modem.get_sim_signal` | ok | |
| `modem.get_slot_config` | skipped | no ID from a previous list response |
| `modem.get_sms_list` | ok | |
| `modem.get_status` | ok | |
| `modem.get_traffic_config` | skipped | no ID from a previous list response |
| `mwan3.get_config` | error | -32601 Method not found |
| `mwan3.get_status` | error | -32601 Method not found |
| `nas-web.get_disk_list` | ok | |
| `nas-web.get_file_list` | ok | |
| `nas-web.get_nas_ser` | ok | |
| `nas-web.get_proto_config` | ok | |
| `nas-web.get_share_list` | ok | |
| `nas-web.get_status` | ok | |
| `nas-web.get_user_list` | ok | |
| `netmode.get_mode` | ok | |
| `network.check_wan_cable` | ok | |
| `network.get_arp_list` | ok | |
| `network.get_dhcp_leases` | error | -32601 Method not found |
| `network.get_hwnat_config` | error | -32601 Method not found |
| `network.get_netnat_config` | ok | |
| `network.routes` | error | -32601 Method not found |
| `network.routes6` | error | -32601 Method not found |
| `otbr.get_bbr_status` | error | -32601 Method not found |
| `otbr.get_joiner_list` | error | -32601 Method not found |
| `otbr.get_neighbor_list` | error | -32601 Method not found |
| `otbr.get_network_data` | error | -32601 Method not found |
| `otbr.get_srp_server_config` | error | -32601 Method not found |
| `otbr.get_srp_server_service` | error | -32601 Method not found |
| `otbr.get_status` | error | -32601 Method not found |
| `ovpn-client.get_all_config_list` | ok | |
| `ovpn-client.get_config_list` | ok | |
| `ovpn-client.get_group_list` | ok | |
| `ovpn-client.get_route_list` | skipped | no ID from a previous list response |
| `ovpn-client.get_setting` | skipped | no ID from a previous list response |
| `ovpn-client.get_status` | ok | |
| `ovpn-server.get_config` | ok | |
| `ovpn-server.get_route_list` | ok | |
| `ovpn-server.get_setting` | ok | |
| `ovpn-server.get_status` | ok | |
| `ovpn-server.get_user_list` | ok | |
| `parental-control.get_app_list` | ok | |
| `parental-control.get_brief` | skipped | no ID from a previous list response |
| `parental-control.get_config` | ok | |
| `parental-control.get_status` | ok | |
| `plugins.get_config` | ok | |
| `plugins.get_list` | ok | |
| `plugins.get_package_info` | skipped | no ID from a previous list response |
| `plugins.get_repository_status` | ok | |
| `qos.get_bandwidth_config` | ok | |
| `qos.get_channel_bandwidth_ratio` | ok | |
| `qos.get_client_list` | ok | |
| `qos.get_config` | ok | |
| `qos.get_device_group` | ok | |
| `reboot.get_config` | error | -32601 Method not found |
| `repeater.get_config` | ok | |
| `repeater.get_saved_ap_list` | ok | |
| `repeater.get_status` | ok | |
| `rs485.get_config` | error | -32601 Method not found |
| `rs485.get_connect_status` | error | -32601 Method not found |
| `rs485.get_forward_config` | error | -32601 Method not found |
| `rs485.get_forward_log` | error | -32601 Method not found |
| `rs485.get_tcp_clients` | error | -32601 Method not found |
| `rtty.get_config` | ok | |
| `s2s.get_status` | ok | |
| `samba.get_config` | error | -32601 Method not found |
| `sms-forward.get_config` | ok | |
| `switch-button.get_config` | ok | |
| `switch-button.get_funcs` | ok | |
| `system.disk_info` | ok | |
| `system.get_httpd_mem_status` | ok | |
| `system.get_info` | ok | |
| `system.get_load` | ok | |
| `system.get_network_status` | error | -32601 Method not found |
| `system.get_security_policy` | ok | |
| `system.get_status` | ok | |
| `system.get_timezone_config` | ok | |
| `system.get_timezone_list` | error | -32601 Method not found |
| `system.get_unixtime` | ok | |
| `tailscale.get_config` | ok | |
| `tailscale.get_exit_node_list` | ok | |
| `tailscale.get_status` | ok | |
| `tethering.get_status` | ok | |
| `timer.get_disk` | ok | |
| `timer.get_led` | ok | |
| `timer.get_reboot` | ok | |
| `timer.get_wifi` | ok | |
| `tor.get_config` | ok | |
| `tor.get_status` | ok | |
| `ui.get_lang` | ok | |
| `ui.get_menu_list` | ok | |
| `upgrade.get_config` | ok | |
| `upgrade.get_online_upgrade_status` | ok | |
| `vpn-client.get_all_config_list` | error | -32601 Method not found |
| `vpn-client.get_connection_methods` | error | -32601 Method not found |
| `vpn-client.get_status` | error | -32601 Method not found |
| `vpn-client.get_tunnel` | error | -32601 Method not found |
| `vpn-policy.get_domain_policy` | ok | |
| `vpn-policy.get_global_policy` | ok | |
| `vpn-policy.get_mac_policy` | ok | |
| `vpn-policy.get_proxy_mode` | ok | |
| `vpn-policy.get_vlan_policy` | ok | |
| `wg-client.get_all_config_list` | ok | |
| `wg-client.get_config_list` | ok | |
| `wg-client.get_group_list` | ok | |
| `wg-client.get_route_list` | ok | |
| `wg-client.get_setting` | ok | |
| `wg-client.get_status` | ok | |
| `wg-server.get_config` | ok | |
| `wg-server.get_peer_list` | ok | |
| `wg-server.get_route_list` | ok | |
| `wg-server.get_setting` | ok | |
| `wg-server.get_status` | ok | |
| `wifi.get_config` | ok | |
| `wifi.get_status` | ok | |
| `zerotier.get_config` | ok | |
| `zerotier.get_status` | ok | |
