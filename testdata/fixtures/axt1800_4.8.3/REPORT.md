# Probe report: GL.iNet AXT1800, firmware 4.8.3

- Model: `axt1800` (GL-AXT1800, "Slate AX"). The same router as `axt1800_4.7.0`, after a firmware upgrade.
- Firmware: `4.8.3`.
- Probe date: 2026-11-18. Program: `cmd/probe` (read-only allowlist in `cmd/probe/calls.go`), with the redaction changes of this version (see "Redaction").
- Result: 173 read methods in the allowlist. 120 ok, 43 error, 10 skipped (no ID from a previous list response). 4.7.0: 126 ok, 37 error, 10 skipped.
- Calls: the probe ran 3 times. The outputs of the first two runs were audited and then deleted (see "Redaction"). Each run sent `challenge`, `login`, 163 read calls and one call with an invalid sid. All calls were reads. No write method was called on 4.8.3.

## Login

`challenge` returns `{"alg":1,"salt":...,"nonce":...,"hash-method":"sha256"}`. The login hash is SHA-256, not MD5. Client `v0.10.2` and later support this. `login` returns `{"sid":...,"username":"root"}` as on 4.7.0. An invalid sid gives `-32000 Access denied` as on 4.7.0.

## Methods that are gone

These methods return `{"code":-32601,"message":"Method not found"}` on 4.8.3 and were ok on 4.7.0. For the `wg-client` methods, the nginx log of the router shows `dlsym: Symbol not found`, so the module has no such function. The params are not the cause.

| Method | Client | Effect |
|---|---|---|
| `wg-client.get_status` | `WireGuardClient.GetStatus` | returns `*RPCError` -32601 |
| `wg-client.get_setting` | `WireGuardClient.GetSetting` | returns `*RPCError` -32601 |
| `wg-client.get_route_list` | `WireGuardClient.ListRoutes` | returns `*RPCError` -32601 |
| `vpn-policy.get_proxy_mode`, `get_global_policy`, `get_domain_policy`, `get_mac_policy`, `get_vlan_policy` | all reads of `VPNPolicy` | return `*RPCError` -32601; the whole `vpn-policy` module is gone |
| `ovpn-client.get_status` | not in the client | |
| `cloud.get_batch_config` | not in the client | |

The other `wg-client` reads (`get_group_list`, `get_all_config_list`, `get_config_list`) are still ok.

## New methods

The `vpn-client` module is on 4.8.3. All 4 called methods are ok. The client does not call them.

| Method | Result shape |
|---|---|
| `vpn-client.get_status` | `{"mode":int, "status_list":[{tunnel_id, type ("wireguard"), via (interface, for example "wgclient1"), status, enabled, group_id, peer_id, name, group_name, peer_name, domain:[string], port, ipv4, rx_bytes, tx_bytes}]}` |
| `vpn-client.get_tunnel` | `{"global_enabled":bool, "tunnels":[...], "default_tunnels":[...]}`. A tunnel has `id` (a UCI section), `tunnel_id`, `name`, `enabled`, `killswitch`, `from{type}`, `to{type, manual, domain_list, domain_list_len}`, `via{type, via, group_id, peer_id}` and `options{local_access, masq, mtu, service_policy, options_in_used, tunnel_using_options}`. The default tunnel has `via{via:"novpn"}` and no `options`. |
| `vpn-client.get_all_config_list` | `{"configs":{"wireguard":[group with peers], "openvpn":[...]}}`. The WireGuard groups have the fields of `wg-client.get_all_config_list`. |
| `vpn-client.get_connection_methods` | `{"connection_methods":["lan","wan","guest"]}` |

## WireGuard client status

On 4.8.3, the status of the WireGuard client tunnel is in `vpn-client.get_status`, in the `status_list` item with `type` `wireguard`. `vpn-client.get_tunnel` has the same tunnel. It has `via.group_id` and `via.peer_id` of the peer, `killswitch`, and the options that `wg-client.get_setting` had on 4.7.0 (`local_access`, `masq`, `mtu`). No 4.8.3 read that the probe sent has the route rules of `wg-client.get_route_list`.

On 4.8.3, `system.get_status.service` has no `wgclient` item and no `ovpnclient` item. On 4.7.0 it had `wgclient` (with `group_id` and `peer_id`) and `ovpnclient`. The other services are the same.

## Changed result shapes

Only fields and types are compared, not values that are runtime data.

| Method | Change | Client |
|---|---|---|
| `system.get_info` | `hardware_feature` has the new keys `screen` (bool), `simo` (bool), `usb_power` (string) | `HardwareFeature.Screen`, `SIMO`, `USBPower` |
| `system.get_status` | no `wgclient` and no `ovpnclient` items in `service` | no decode error |
| `wg-client.get_all_config_list` | each group has the new key `group_type` | `WireGuardGroupPeers.Type` (0 on 4.7.0) |
| `wg-server.get_config` | no `ipv6_enable` | `WireGuardServerConfig.IPv6Enable` is nil; the cause is not known |
| `ovpn-server.get_config` | no `ipv6_enable` | not in the client |
| `wifi.get_config` | new top-level key `bandmode`, an empty array | not in `WifiConfig`, because the item type is not known |
| `repeater.get_config` | `auto` is a number (1), not a bool | `RepeaterConfig.Auto` reads both; `true` is 1 and `false` is 0 (assumption) |
| `switch-button.get_funcs` | `funcs` items are objects `{func, sub_func?:[{func}]}`, not strings; `vpn` with the tunnel IDs as `sub_func` replaces `wireguard` and `openvpn` | not in the client |
| `switch-button.get_config` | new key `sub_func` (a tunnel ID as a string when `func` is `vpn`) | not in the client |
| `local-access.get_config` | new key `ssh_enabled` (bool) | `LocalAccessConfig.SSHEnabled` (nil on 4.7.0) |
| `switch-button.get_funcs` | `funcs` items are objects `{func, sub_func?:[{func}]}`, not strings; `vpn` with the tunnel IDs as `sub_func` replaces `wireguard` and `openvpn` | `SwitchButton.GetFuncs` decodes both shapes |
| `switch-button.get_config` | new key `sub_func` (a tunnel ID as a string when `func` is `vpn`) | `SwitchButtonConfig.SubFunc` |
| `local-access.get_config` | new key `ssh_enabled` (bool) | not in the client |
| `cloud.get_config` | new key `aw_api_url` (string) | not in the client |
| `modem.get_profile_list` | result is `{"err_code":20002001,"err_msg":"Parameter not found"}`, not a profile list (the router has no modem) | not in the client |
| `challenge` | new key `hash-method` | supported since `v0.10.2` |

All other ok methods have the same fields and types as on 4.7.0.

These differences come from the router state, not from the firmware:

- `firewall.get_dmz` has `priority` (bool). On 4.7.0 it appeared after the first `set_dmz` of the write test. `DMZ.Priority` (client `v0.12.0`) has it; it is nil when `get_dmz` has no `priority`.
- `zerotier.get_config` has `id`: the made-up network ID of the 4.7.0 write test.
- `wg-client.get_config_list` has no `ipv6_enable` and no `mtu`, because the probe read the group of type 3 on 4.8.3 and the custom group (type 2) on 4.7.0. The `get_all_config_list` peers of the custom group still have `ipv6_enable` and `mtu`.

## Effect on the client and glinettest

- `TestReadsDecodeFirmware483Fixtures` (`firmware_4_8_3_test.go`) checks that each typed read that 4.8.3 still has keeps each value of its 4.8.3 fixture. Omitted keys: `wifi.get_config` `bandmode`, and `wg-client.get_config_list` `auth_type`.
- The typed reads that do not work on 4.8.3 are `WireGuardClient.GetStatus`, `GetSetting`, `ListRoutes` and all reads of `VPNPolicy`. A replacement needs typed methods for `vpn-client` (status, and tunnels with their options and policies). The client does not have them.
- `glinettest` emulates firmware 4.7.0. It has no `vpn-client` handlers.

## Effect on the provider

`terraform-provider-glinet` uses only these client methods for VPN:

- the `wg-client` group and peer methods (`ListGroups`, `ListPeers`, `AddGroup`, `RenameGroup`, `RemoveGroup`, `AddPeer`, `SetPeer`, `RemovePeer`);
- `WireGuardServer` (`GetConfig`, `ListPeers`, and the peer writes).

It does not use the methods that are gone. The reads that it uses decode on 4.8.3. On 4.8.3, `glinet_system_status` shows no `wgclient` service and no `ovpnclient` service.

## DNS modes

`DNSService.SetConfig` (client `v0.11.0`) sends the fields of each DNS mode. The formats of the `manual`, `proxy` and `secure` writes come from the web UI code of firmware 4.8.3 (the DNS view, `handleApplyDnsConfig`) and were NOT tested on a router. Only `auto` writes were tested, on firmware 4.7.0 (see the 4.7.0 report).

| `mode` | Fields that the web UI sends in addition to `force_dns`, `rebind_protection`, `override_vpn` |
|---|---|
| `auto` | none |
| `manual` | `server`: 1 to 4 IPv4 or IPv6 addresses; the web UI refuses a list without an IPv4 address |
| `proxy` | `server`: one item `"<ip>#<port>"`, port 1 to 65535; the web UI adds `#53` when the user gives no port |
| `secure` | `proto`: one of `dns.get_info` `protos`. For `DoT`: `dot_provider` `"0"` (Cloudflare, no ID), `"1"` (NextDNS, `nextdns_id`, can be empty), `"2"` (Control D paid plan, `controld_id`), `"3"` (Control D free, `free_controld_id`, a value such as `"p0"`). For `DoH`, `DNSCrypt`, `oDoH`: `server`, one or more names from `dns.get_info` `serverlist`. |

While AdGuard Home is enabled, the web UI sends only the three flags. The web UI shows the error `-3` of `set_config` as an invalid resolver ID.

The web UI reads `proto` from `get_config` only in mode `secure`, `server` as the address list in `manual`, as `server[0]` in `proxy`, and as the server names in `secure` with `DoH`, `DNSCrypt` or `oDoH`. It reads `dot_provider` `"2"` and `"3"` as Control D and an empty `dot_provider` as Control D.

Not known: what the router returns in `server` in mode `secure` with `DoT`; if the router returns `proto` in modes other than `secure` and `auto`; if a write keeps the IDs of the other DoT providers; how the router answers an invalid write.

## Write formats from the web UI code (client v0.12.0 and v0.12.1)

These write formats and read fields come from the web UI code of firmware 4.8.3 (the views `firewallview`, `lanip`, `guestip`, `internet`, `clients` and `wireless`). They were NOT tested on a router. No write was called on 4.8.3.

| Client | Call | Params | Web UI |
|---|---|---|---|
| `Firewall.SetDMZ` | `firewall.set_dmz` | `{enabled, dest_ip?, priority}`; `priority` only when `enabled` is true | DMZ pane, `setDmz` |
| `Firewall.OrderPortForwards` | `firewall.order_port_forward` | `{id_list:[string]}`, the IDs of all port forwards in the new order | port forward card, `handleApplyPriority` |
| `LAN.SetAddress` | `lan.set_config` | `{interface, ip, netmask, ap_isolate, wan_isolate}`; `wan_isolate` (0 or 1) only for `guest` | `handleApplyLan`, `handleChangeWanIsolate` |
| `Cable.GetConfig`, `Cable.SetConfig` | `cable.get_config`, `cable.set_config` | `secondwan: true` for `SecondWANPort` | `getCableConfig`, `setCableConfig` |
| `Network.CheckWANCable` | `network.check_wan_cable` | `{secondwan: bool}` | not in the web UI code that was read; `secondwan: true` is not settled |
| `AccessControl.SetConfig` | `black_white_list.set_config` | `{mode, mac:[string]}` | access control dialog, `handleSetAccessConfig`, `handleChangeMode` |
| `AccessControl.AddAllowedMAC`, `RemoveAllowedMAC` | `black_white_list.set_single_mac` | `{mode:"white", operate:"add"\|"del", mac}` | client list switch, `blockChange` |
| `Wifi.SetConfig` | `wifi.set_config` | `chan_6g_only_psc` (bool) for the main interface of a 6G radio | `applyConfig` |
| `LAN.SetDHCP` | `lan.set_config` | with `enable` 1, always `lpr` without the empty inputs, so `[]` when there is no LPR server | `handleApplyDhcp` |
| `LAN.RemoveAllStaticBinds` | `lan.remove_static_bind` | `{mode: 1}` | address reservation card, `handleRemoveAll` |
| `Firewall.RemoveAllPortForwards` | `firewall.remove_port_forward` | `{all: true}` | port forward card, `removePortForward` |
| `Firewall.RemoveAllRules` | `firewall.remove_rule` | `{all: true}` | rule list (view `security`), `removeRule` |
| `WireGuardServer.RemoveAllPeers` | `wg-server.remove_peer` | `{all: true}` | peer card, `removePeer` |
| `WireGuardServer.AddPeer`, `SetPeer` | `wg-server.add_peer`, `set_peer` | `persistent_keepalive` and `mtu` always, `""` when empty | `submitFormData` |
| `WireGuardServer.AddRoute`, `SetRoute` | `wg-server.add_route`, `set_route` | `gateway` (also `""`), `metric` and `mtu` (0 when empty) always; no `scope` | route dialog, `handleApply` |
| `WireGuardServer.SetConfig` | `wg-server.set_config` | `{address_v4, address_v6, port, private_key}`, each also when `""`; no `ipv6_enable` | server card, `handleApply` |
| `WireGuardClient.SetPeer` | `wg-client.set_config` | `listen_port`, `persistent_keepalive` and `mtu` as `""` when the user empties the input; not sent when the user did not fill it | `submitData` |

What the router does with `lpr: []`, with `""` for a number field of a WireGuard peer, with `gateway: ""` and 0 in a route, and with the "remove all" calls is not known.

The client differs from the web UI in these cases:

- `WireGuardServer.SetConfig` does not send `private_key: ""`. The private key input of the web UI shows "Auto Generate" when it is empty, but the web UI code does not settle what the router does with `""`. The client sends `ipv6_enable` only when `IPv6Enable` is not nil and `address_v6` is not empty, as the web UI of firmware 4.7.0 did. So a configuration read from firmware 4.8.3 goes back without it.
- `WireGuardClient.SetPeer` sends `""` for each of `listen_port`, `persistent_keepalive` and `mtu` that is 0. The web UI sends `""` only for an input that the user emptied. `WireGuardClient.AddPeer` does not send them when they are 0, as the web UI does for a new peer.

Read fields:

- `firewall.get_dmz` `priority`: the web UI shows `true` with the label key `firewall.dmz_priority_higher_label` and `false` with `firewall.dmz_priority_lower_label`. The label texts are not in the web UI code, so to what the priority is compared is not settled. The web UI shows `true` when the result has no `priority`.
- `lan.get_config_list` `wan_isolate`: only for `guest`; label key `lanip.block_wan_label`. The fixture does not have it; the web UI shows 0 then.
- `lan.get_config_list` `lpr`: label key `lan_ip.lpr_serve_label` (LPR server); at most 5 IP addresses, each one time.
- `cable.get_status` `secondwan`: on a router with a second WAN port, the web UI reads `mode` and `status` as for the first port.
- `black_white_list.get_config` `white_mac`: the allow list.
- `wifi.get_config` `chan_6g_only_psc`: only on a 6G radio; true limits the channel list to the PSC channels.
- `clients.get_list` `ipv6` (list of IPv6 addresses, shown by the web UI), `last_rx` and `last_tx` (60 numbers each in the fixtures; the web UI does not use them, and their meaning is not known).

Not built, because the web UI code does not settle the shapes:

- `cable.get_ports_config`: built later from the view `netport` as `Cable.ListPorts`; see "Network mode, ports and advanced network settings (web UI code)".

- `wifi.get_mlo_config` and `wifi.set_mlo_config` (MLO, when `software_feature.mlo` is true). The web UI reads `res` with `random_bssid`, `encryptions`, `mlo_bandsupport` and `ifaces[]`; it reads `name`, `guest`, `mlo_enable`, `mlo_band` (a list of `"2g"`, `"5g"`, `"6g"`, at least 2), `ssid`, `encryption`, `key` and `hidden` of an interface. It writes `{name, mlo_enable}` for the switch, and for the form all fields of the interface object as the router gave them, plus `random_bssid` for a main interface. So the full params depend on fields of the result that the web UI code does not name.
- The provider groups (type 1) of `wg-client`. The web UI uses `wg_client.create_provider_account {group_id, username, password}` (result `expires_at`), `wg_client.get_provider_server_list` or `wg-client.get_third_config` (AzireVPN and Mullvad) with `{group_id, username, password?}`, `wg_client.generate_provider_config` or `wg-client.get_recommend_config` with `{group_id, country_name:[string]}` or `{group_id, server_info:[{country_name, cities:[{city_name, hostname:[string]}]}]}`, `wg_client.update_provider_config {group_id, update_all:true}`, `wg_client.get_provider_user_info {group_id}` (result `allocated`, `available`, `expires_at`), `wg_client.get_max_client`, `wg-client.clear_user_pass {group_id}` and `wg-client.clear_config_list {group_id, delete_key}`. The module, the form fields and the steps change with the provider name (a list in the web UI code), with `auth_type` (1: the form has a password) and with `procedure` (1: the user selects servers). The results are only partly read. These calls send account data to a third-party VPN service through the router.
- The groups of type 3 (from the GL.iNet app): no new method is needed. The web UI shows them only when they have peers or clients, and has no call that is only for them: it edits and removes their peers with `wg-client.set_config` and `remove_config`, as for a custom group (`WireGuardClient.SetPeer`, `RemovePeer`). It does not show the add button of a custom group for them.

## Not checked

- Every write method. No write was called on 4.8.3. It is not known if the `wg-client` writes still exist or work, or what the `vpn-client` writes are.
- `wg-client.get_config_list` of the custom group (type 2). The probe reads only the first group that has peers.
- Why `ipv6_enable` is not in `wg-server.get_config` and `ovpn-server.get_config`: the firmware removed it, or the router sends it only in some states.
- The item type of `wifi.get_config` `bandmode`.
- The `vpn-client` methods that the probe does not call.
- The meaning of new values, for example `vpn-client.get_status` `mode`.

## Methods not called

As on 4.7.0: all `set_*`, `add_*`, `remove_*` and other write methods. Also these `get_*` methods: `upgrade.check_firmware_online`, `tailscale.get_auth_url`, `wg-client.get_third_config` and `ovpn-client.get_third_config` (they send credentials to a third-party VPN service), `wg-client.get_recommend_config` and `ovpn-client.get_recommend_config` (they query an external service).

## Redaction

The probe output went first to a directory outside the repository. Scripts compared each string value with the same field of the 4.7.0 fixtures. They listed each new or changed value by file, JSON path, length and hash prefix, and each listed value was checked. For each value that the redactor did not replace, `cmd/probe/redact.go` got a rule with a test that uses made-up input. Then the probe ran again and the audit was done again.

The rules of the 4.7.0 report apply. Rules added for this probe:

- connection UUIDs (`uuid`): `00000000-0000-4000-8000-00000000000N`;
- the reboot schedule (an object with `hour`, `min` and `week`): `02`, `15`, `[3, 5]`, the made-up values of the 4.7.0 fixtures;
- dates and times of day in log lines (`log`), because they show the time zone: `Thu Jan  1`, `<year>/01/01` and `00:00:00`;
- names of VPN tunnels (`name` in objects that have a `tunnel_id`): `tunnel-N`, except the firmware default tunnel `last sort default policy`;
- `peer_name`: `peer-N`;
- host names in `domain` lists and in `domain_list`: `host-N.example.com`.

Kept as on 4.7.0: IDs (group, peer, tunnel, UCI section), the flags and numbers of the VPN settings, and firmware names and URLs.

## Results

"same" in the 4.7.0 column means that the result is the same as on 4.8.3.

| Method | 4.8.3 | 4.7.0 |
|---|---|---|
| `acl.get_acl_list` | skipped | same |
| `acl.get_group_list` | ok | same |
| `adguardhome.get_config` | ok | same |
| `bark.get_config` | error -32601 | same |
| `bark.get_status` | error -32601 | same |
| `black_white_list.get_config` | ok | same |
| `cable.get_config` | ok | same |
| `cable.get_status` | ok | same |
| `clients.get_list` | ok | same |
| `clients.get_status` | ok | same |
| `cloud-batch-manage.get_2b_config` | error -32601 | same |
| `cloud-batch-manage.get_batch_config` | error -32601 | same |
| `cloud.get_batch_config` | error -32601 | ok |
| `cloud.get_config` | ok | same |
| `ddns.get_config` | ok | same |
| `ddns.get_status` | ok | same |
| `dlna.get_config` | error -32601 | same |
| `dns.get_config` | ok | same |
| `dns.get_host` | ok | same |
| `dns.get_info` | ok | same |
| `edgerouter.get_config` | ok | same |
| `edgerouter.get_status` | ok | same |
| `fan.get_config` | ok | same |
| `fan.get_status` | ok | same |
| `firewall.get_acl_rule_list` | error -32601 | same |
| `firewall.get_acl_zone_list` | error -32601 | same |
| `firewall.get_dmz` | ok | same |
| `firewall.get_port_forward_list` | ok | same |
| `firewall.get_rule_list` | ok | same |
| `firewall.get_wan_access` | ok | same |
| `firewall.get_zone_list` | ok | same |
| `igmp.get_config` | ok | same |
| `ipv6.get_ipv6` | ok | same |
| `kmwan.get_config` | ok | same |
| `kmwan.get_status` | ok | same |
| `lan.get_config_list` | ok | same |
| `lan.get_static_bind_list` | ok | same |
| `lan.get_wan_info` | ok | same |
| `led.get_config` | ok | same |
| `local-access.get_config` | ok | same |
| `logread.get_config` | error -32601 | same |
| `logread.get_crash_log` | ok | same |
| `logread.get_kernel_log` | ok | same |
| `logread.get_nginx_log` | ok | same |
| `logread.get_system_log` | ok | same |
| `macclone.get_mac` | error -32601 | same |
| `mcu.get_battery_config` | error -32601 | same |
| `mcu.get_oled_config` | error -32601 | same |
| `modem.get_cell_tower` | ok | same |
| `modem.get_cells_info` | skipped | same |
| `modem.get_config` | skipped | same |
| `modem.get_debug_msg` | skipped | same |
| `modem.get_info` | ok | same |
| `modem.get_name` | ok | same |
| `modem.get_profile_list` | ok (err_code 20002001) | ok |
| `modem.get_signals` | ok | same |
| `modem.get_sim_info` | ok | same |
| `modem.get_sim_signal` | ok | same |
| `modem.get_slot_config` | skipped | same |
| `modem.get_sms_list` | ok | same |
| `modem.get_status` | ok | same |
| `modem.get_traffic_config` | skipped | same |
| `mwan3.get_config` | error -32601 | same |
| `mwan3.get_status` | error -32601 | same |
| `nas-web.get_disk_list` | ok | same |
| `nas-web.get_file_list` | ok | same |
| `nas-web.get_nas_ser` | ok | same |
| `nas-web.get_proto_config` | ok | same |
| `nas-web.get_share_list` | ok | same |
| `nas-web.get_status` | ok | same |
| `nas-web.get_user_list` | ok | same |
| `netmode.get_mode` | ok | same |
| `network.check_wan_cable` | ok | same |
| `network.get_arp_list` | ok | same |
| `network.get_dhcp_leases` | error -32601 | same |
| `network.get_hwnat_config` | error -32601 | same |
| `network.get_netnat_config` | ok | same |
| `network.routes` | error -32601 | same |
| `network.routes6` | error -32601 | same |
| `otbr.get_bbr_status` | error -32601 | same |
| `otbr.get_joiner_list` | error -32601 | same |
| `otbr.get_neighbor_list` | error -32601 | same |
| `otbr.get_network_data` | error -32601 | same |
| `otbr.get_srp_server_config` | error -32601 | same |
| `otbr.get_srp_server_service` | error -32601 | same |
| `otbr.get_status` | error -32601 | same |
| `ovpn-client.get_all_config_list` | ok | same |
| `ovpn-client.get_config_list` | ok | same |
| `ovpn-client.get_group_list` | ok | same |
| `ovpn-client.get_route_list` | skipped | same |
| `ovpn-client.get_setting` | skipped | same |
| `ovpn-client.get_status` | error -32601 | ok |
| `ovpn-server.get_config` | ok | same |
| `ovpn-server.get_route_list` | ok | same |
| `ovpn-server.get_setting` | ok | same |
| `ovpn-server.get_status` | ok | same |
| `ovpn-server.get_user_list` | ok | same |
| `parental-control.get_app_list` | ok | same |
| `parental-control.get_brief` | skipped | same |
| `parental-control.get_config` | ok | same |
| `parental-control.get_status` | ok | same |
| `plugins.get_config` | ok | same |
| `plugins.get_list` | ok | same |
| `plugins.get_package_info` | skipped | same |
| `plugins.get_repository_status` | ok | same |
| `qos.get_bandwidth_config` | ok | same |
| `qos.get_channel_bandwidth_ratio` | ok | same |
| `qos.get_client_list` | ok | same |
| `qos.get_config` | ok | same |
| `qos.get_device_group` | ok | same |
| `reboot.get_config` | error -32601 | same |
| `repeater.get_config` | ok | same |
| `repeater.get_saved_ap_list` | ok | same |
| `repeater.get_status` | ok | same |
| `rs485.get_config` | error -32601 | same |
| `rs485.get_connect_status` | error -32601 | same |
| `rs485.get_forward_config` | error -32601 | same |
| `rs485.get_forward_log` | error -32601 | same |
| `rs485.get_tcp_clients` | error -32601 | same |
| `rtty.get_config` | ok | same |
| `s2s.get_status` | ok | same |
| `samba.get_config` | error -32601 | same |
| `sms-forward.get_config` | ok | same |
| `switch-button.get_config` | ok | same |
| `switch-button.get_funcs` | ok | same |
| `system.disk_info` | ok | same |
| `system.get_httpd_mem_status` | ok | same |
| `system.get_info` | ok | same |
| `system.get_load` | ok | same |
| `system.get_network_status` | error -32601 | same |
| `system.get_security_policy` | ok | same |
| `system.get_status` | ok | same |
| `system.get_timezone_config` | ok | same |
| `system.get_timezone_list` | error -32601 | same |
| `system.get_unixtime` | ok | same |
| `tailscale.get_config` | ok | same |
| `tailscale.get_exit_node_list` | ok | same |
| `tailscale.get_status` | ok | same |
| `tethering.get_status` | ok | same |
| `timer.get_disk` | ok | same |
| `timer.get_led` | ok | same |
| `timer.get_reboot` | ok | same |
| `timer.get_wifi` | ok | same |
| `tor.get_config` | ok | same |
| `tor.get_status` | ok | same |
| `ui.get_lang` | ok | same |
| `ui.get_menu_list` | ok | same |
| `upgrade.get_config` | ok | same |
| `upgrade.get_online_upgrade_status` | ok | same |
| `vpn-client.get_all_config_list` | ok | error -32601 |
| `vpn-client.get_connection_methods` | ok | error -32601 |
| `vpn-client.get_status` | ok | error -32601 |
| `vpn-client.get_tunnel` | ok | error -32601 |
| `vpn-policy.get_domain_policy` | error -32601 | ok |
| `vpn-policy.get_global_policy` | error -32601 | ok |
| `vpn-policy.get_mac_policy` | error -32601 | ok |
| `vpn-policy.get_proxy_mode` | error -32601 | ok |
| `vpn-policy.get_vlan_policy` | error -32601 | ok |
| `wg-client.get_all_config_list` | ok | same |
| `wg-client.get_config_list` | ok | same |
| `wg-client.get_group_list` | ok | same |
| `wg-client.get_route_list` | error -32601 | ok |
| `wg-client.get_setting` | error -32601 | ok |
| `wg-client.get_status` | error -32601 | ok |
| `wg-server.get_config` | ok | same |
| `wg-server.get_peer_list` | ok | same |
| `wg-server.get_route_list` | ok | same |
| `wg-server.get_setting` | ok | same |
| `wg-server.get_status` | ok | same |
| `wifi.get_config` | ok | same |
| `wifi.get_status` | ok | same |
| `zerotier.get_config` | ok | same |
| `zerotier.get_status` | ok | same |

## Remote and local access (web UI code)

These calls come from the web UI view `security` of firmware 4.8.3 (cards "remote access" and "local access"). The reads have fixtures on 4.7.0 and 4.8.3. The writes were NOT tested on a router.

| Client | Call | Params or result | Web UI |
|---|---|---|---|
| `Firewall.GetWANAccess` | `firewall.get_wan_access` | `{enable_https, enable_ssh, enable_ping, enable_whitelist, whitelist:[{ipaddr, name}]}`; the fixtures have an empty `whitelist` | `getRemoteAccessConfig` |
| `Firewall.SetWANAccess` | `firewall.set_wan_access` | the same object, always with all fields, also `false` and `whitelist: []` | `handleApply` of the remote access card |
| `LocalAccess.GetConfig` | `local-access.get_config` | `{http_port, https_port, redirect_https, luci_http_port, luci_https_port, luci_redirect_https, session_timeout, ssh_enabled, ssh_port}`; no `ssh_enabled` on 4.7.0 | `getLocalAccessConfig` |
| `LocalAccess.SetConfig` | `local-access.set_config` | the same object, always with all fields, also the LuCI fields when LuCI is not installed; `session_timeout` in seconds | `handleApply` of the local access card |

Web UI rules:

- Whitelist item: `ipaddr` is an IPv4 or IPv6 address; `name` is optional (at most 64 characters) and is sent as `""` when empty.
- The `enable_ssh` switch shows only while `ssh_enabled` of `local-access` is true. The `enable_whitelist` switch shows only while one of `enable_https`, `enable_ssh` and `enable_ping` is true. The web UI always sends all fields.
- Ports are 1 to 65535, and the five ports are all different. The four web ports cannot be a port that browsers block (a list of 80 ports, for example 22, 25 and 6000).
- `session_timeout`: the web UI shows minutes (1 to 180) and sends the value times 60. It shows 5 minutes when the result has no value.
- The LuCI fields show only when `luci.get_status` gives `status` 2 (installed).

Side effects in the web UI code:

- After `firewall.set_wan_access`, the web UI reads `firewall.get_rule_list` again.
- After `local-access.set_config`, the web UI waits 3 s and loads the page from the new address: HTTPS and `https_port` when `redirect_https` is true or the page used HTTPS, otherwise HTTP and `http_port`.
- `local-access.set_config` errors (`err_code`): -1 HTTP port in use, -2 HTTPS port in use, -3 another port in use (the `err_msg` has "ssh", "https" for LuCI HTTPS, or else LuCI HTTP), -4 SSH needs a strong admin password (label key `security.ssh_need_strong_pwd_tips`).

Open points:

- Whether `set_wan_access` adds or changes firewall rules, and so changes the IDs of the rules and port forwards.
- The `whitelist` item shape as the router returns it (no fixture has an item), and whether the router accepts an address with a prefix length.
- What a `set_wan_access` or `set_config` without some fields does.
- What firmware 4.7.0 does with `ssh_enabled` in `local-access.set_config`.
- Whether `local-access.set_config` restarts the web server before it answers, and so whether the answer arrives when a port changes.
## Client name, icon and speed limit (web UI code)

From the web UI code of firmware 4.8.3 (view `clients`). No write was tested on a router.

| Call | Params | Client method | Notes |
|---|---|---|---|
| `clients.get_list` | `{}` | `Clients.ListClients` | read; fixture. `alias` and `class` only when set. `limit_rx` (download) and `limit_tx` (upload) in KiB/s, 0 is no limit |
| `clients.set_info` | `{mac, alias, class}` | `Clients.SetInfo` | always all three fields; `""` removes the name or the icon |
| `qos.set_speed_limit_rule` | `{mac, download, upload}` | `QoS.SetSpeedLimit` | KiB/s, 1 to 1073741824, a fraction in steps of 0.125 in the Kbps mode, 0 is no limit; the UI waits up to 30 s |
| `qos.remove_speed_limit_rule` | `{mac}` | `QoS.RemoveSpeedLimit` | the UI waits up to 30 s |

- The alias has at most 32 characters and no `,` or `"`. The web UI offers the classes `computer`, `phone`, `pad`, `camera`, `watch`, `laptop`, `printer`, `sound`, `television`, `smartappliances`, `games`, `gateway`, `nas`, `server`, `switch`.
- The web UI shows the speed limit options only in system mode 0 (router) and 6 (passthrough). It shows "set limit" only while `netnat_enabled` (network acceleration) is false, and "remove limit" also while it is true. The test router has `netnat_enabled: true`.
- The web UI has the unit modes KB/s and Kbps. Each input accepts only digits. In the KB/s mode the web UI sends the input (1 to 1073741824). In the Kbps mode it accepts 8 to 8589934592 and sends the input divided by 8, for example `1.125` for 9 Kbps. So `SpeedLimit.Download` and `Upload`, and `ClientDevice.LimitRx` and `LimitTx`, are `float64`.

Open points:

- Whether `set_info` removes `alias` and `class` for `""` or keeps an empty value in `get_list`.
- What `set_speed_limit_rule` does with `download: 0, upload: 0` (the web UI permits it), and whether the router keeps a fraction in `limit_rx` and `limit_tx` or rounds it.
- Whether `set_info` and the `qos` writes accept a MAC that is not in `get_list`, and whether they compare MACs with their letter case.
- What network acceleration does to a limit.
- `qos.get_config` (`{"enable": false, "mode": "0"}`), `get_client_list`, `get_device_group`, `get_bandwidth_config` and `get_channel_bandwidth_ratio` give empty or default results on the test router, and no view of the web UI of firmware 4.8.3 calls them or `qos.set_config`. The client has no typed method for them.
## Hardware settings (web UI code)

These calls come from the web UI code of firmware 4.8.3 (the views `overview`, `timedtask` and `btnsettings`). Only `fan.set_config` was tested on a router.

| Client | Call | Params or result | Source |
|---|---|---|---|
| `Fan.GetConfig` | `fan.get_config` | `{temperature, warn_temperature}` (integers, °C) | fixture |
| `Fan.GetStatus` | `fan.get_status` | `{speed, status}`: speed in RPM, status true while the fan turns | fixture |
| `Fan.SetTemperature` | `fan.set_config` | `{temperature}`; the web UI sends no other key and lets the user select 70 to 90 | tested on a router on 2026-11-18 (GL-AXT1800, firmware 4.8.3): `{"temperature": 60}` returned `[]`, the next `get_config` had `temperature: 60`, and the fan started (status true, speed more than 0) with the CPU at 69 °C |
| `System.GetUSB3Disabled` | `system.get_usb3_disable` | `{usb3_disable: bool}`; the web UI shows true when the key is missing | web UI code, `getUsb3` |
| `System.SetUSB3Disabled` | `system.set_usb3_disable` | `{usb3_disable: bool}`; true is USB 2.0, false is USB 3.0 | web UI code, `handleSetUsb3` |
| `Timer.GetScreenSchedule` | `timer.get_screen` | the fields of `timer.get_led` | web UI code, `getScreenTimer` |
| `Timer.SetScreenSchedule` | `timer.set_screen` | `{enable, week}`, and `turnon_hour`, `turnon_min`, `turnoff_hour`, `turnoff_min` only with `enable` true, as for `set_led` | web UI code, `setScreenTimer` |
| `SwitchButton.GetConfig` | `switch-button.get_config` | `{func, sub_func?}` | fixture |
| `SwitchButton.GetFuncs` | `switch-button.get_funcs` | `{funcs:[{func, sub_func?:[{func}]}]}`; strings on firmware 4.7.0 | fixtures |
| `SwitchButton.SetConfig` | `switch-button.set_config` | `{func, sub_func?, sync: bool}`; `func` `"none"` has no `sub_func` | web UI code, `setBtnConfig` |

The web UI shows the fan dialog when `hardware_feature.fan` is true, the USB switch when `hardware_feature.usb3` is not empty, and the screen schedule when `hardware_feature.screen` is true. The test router has a fan and a USB 3.0 port, and no screen.

The web UI sends `sync: true` only when the user accepts the dialog with label key `btnsettings.auto_sync_tips`, after `switch-button.check_sync_status {func, sub_func?}` returned `sync_status` 0. It does not call `check_sync_status` for `"none"`. In the VPN global proxy mode, it shows one item "VPN client" and sends `func` `"vpn"` with the first `sub_func` of the `vpn` function.

Not known:

- If the router accepts a fan temperature out of 70 to 90 other than 60, and if `fan.set_config` changes `warn_temperature`. The web UI only shows a warning when `temperature` is `warn_temperature` or less.
- The result of `system.get_usb3_disable` and `timer.get_screen` on a router (no fixture; the probe does not call them). The test router has no screen.
- The effect of the USB switch on the connected devices: the texts of the label keys `overview.switching_to_usb2_tips`, `overview.switching_warning_tips` and `overview.switching_to_usb3_tips` are not in the web UI code.
- If `timer.set_screen` with `enable` true needs all values, as `set_led` does.
- What `sync: true` does, and the meaning of the `sync_status` values other than 0. `check_sync_status` is not in the client.
- What `switch-button.get_config` returns when the function is `"none"`, and how the router answers a `func` that is not in `get_funcs`.
## Network mode, ports and advanced network settings (web UI code)

From the web UI views `netadvanced`, `igmp`, `bridge` (network mode) and `netport` (Ethernet ports) of firmware 4.8.3. No write was tested on a router.

| Client | Method | Params or result | Source |
|---|---|---|---|
| `Network.GetAdvancedConfig` | `network.get_advance_config` | `{nat_enable, sip_enable}`, 0 or 1 | view `netadvanced`; no fixture |
| `Network.SetAdvancedConfig` | `network.set_advance_config` | `{nat_enable, sip_enable}`, both always | view `netadvanced`; the web UI waits 2 s |
| `IGMP.GetConfig` | `igmp.get_config` | `{enable: bool, version: int}` | fixture |
| `IGMP.SetConfig` | `igmp.set_config` | `{enable, version}`, both always; version 1, 2 or 3 | view `igmp`; the web UI waits 15 s |
| `NetworkMode.GetMode` | `netmode.get_mode` | `{mode}`; `mask` and `mac` for `passthrough` | fixture (`{"mode":"router"}`); `mask`, `mac` from view `bridge` |
| `NetworkMode.SetMode` | `netmode.set_mode` | `{mode}` for `router` and `ap`; `{mode, ssid, key?, identity?}` for `relay` and `wds`; `{mode, mask, mac}` for `passthrough` | view `bridge`; result `lanip` after a change to `router` |
| `Cable.ListPorts` | `cable.get_ports_config` | `ports[]` of `{name, mode, default_mode, silk, macaddr: {mode, macaddr, defmacaddr, update, period}}` | view `netport`; no fixture |
| `Cable.SetPort` | `cable.set_port_config` | `{name, mode}`; for mode `wan` also `macaddr: {mode, macaddr, update?, period?}` | view `netport`; the web UI waits 3 s, or 15 s after a mode change |

Network mode:

- The web UI offers `router`, `ap`, `relay` (extender), `wds` and `passthrough`, each one only when the router supports it (`hidden_features`, `hardware_feature.radio` and `nowds`, `software_feature.passthrough`). In a mode other than `router` it offers only the change back to `router`. The menu list also has the mode `mesh`, which this view cannot set.
- `system.get_status` `system.mode` is a number. From the conditions of the view: 0 router, 1 WDS, 2 extender, 4 access point, 6 passthrough. Not checked on a router.
- `relay` and `wds` connect to an upstream Wi-Fi network: `key` and `identity` (802.1X) are sent only when not empty. Error codes of the result: 1 connection failed, 2 no IP address from upstream, 3 conflict with the LAN, 4 wrong password.
- A change reconfigures the whole router. In `ap`, `relay` and `wds` the router takes its address from the upstream network: the web UI tells the user to find it there. After a change to `router`, the web UI opens the address in `lanip`. The web UI waits up to 60 s for the answer and then 2 s to 40 s.
- Before a change, the web UI reads `network.check_wan_cable` (and `cable.get_status` for `passthrough`) and stops when the cable state does not fit the mode, or when `macclone_enabled` is true.

MAC clone:

- The module `macclone` gives -32601 on 4.7.0 and 4.8.3, and the web UI of 4.8.3 does not call it; the router has no view `macclone`.
- The MAC address of a wired WAN port is in `cable.get_ports_config` and `cable.set_port_config` (`macaddr`, modes `default`, `clone`, `random`). `network.check_wan_cable` `macclone_enabled` shows it. The extender mode has its own MAC setting in the `repeater` module, which this client does not have.

Open points:

- No router result for `network.get_advance_config`, `cable.get_ports_config` and `cable.get_ports_status`. The probe allowlist now has them. The types (for example of `period`), the other fields, and the names of the LAN ports are not known.
- `cable.get_ports_status` (`ports[]` of `{name, speed, duplex}`, runtime values) is not in the client.
- What the router restarts for `network.set_advance_config` and `igmp.set_config`, what `netmode.set_mode` returns for modes other than `router`, and what the router does with a mode change that the web UI does not offer.
## Multi-WAN (web UI code)

The module `kmwan` (client `MultiWAN`, view `multiwan`). The reads `get_config` and `get_status` were called on 4.7.0 and 4.8.3 (fixtures). The writes and `get_sensitivity` come from the web UI code of firmware 4.8.3 and were NOT tested on a router. `mwan3.get_config` and `mwan3.get_status` return -32601.

| Client | Call | Params or result | Web UI |
|---|---|---|---|
| `MultiWAN.GetConfig` | `kmwan.get_config` | result `{mode, interfaces:[{interface, metric, weight, enable_check, track_mode, track_proto, track_method, track_ipv4, track_ipv6, enable_ssl}]}` (fixture) | `getMwanConfig` |
| `MultiWAN.GetStatus` | `kmwan.get_status` | result `{interfaces:[{interface, status_v4, status_v6}]}` (fixture); 0 online, 1 offline, 2 error | `getMwanStatus`, each 10 s |
| `MultiWAN.GetSensitivity` | `kmwan.get_sensitivity` | result `{sensitivity:{level, val}}`; the web UI shows `medium` and 5 when `sensitivity` is missing | `getSensitiveData` |
| `MultiWAN.SetFailover` | `kmwan.set_config` | `{mode:0, interfaces:[{interface, metric}]}`, metric 1, 2, 3 in the order of the user; the hidden interfaces last | config card, `handleSetConfig` |
| `MultiWAN.SetLoadBalance` | `kmwan.set_config` | `{mode:1, interfaces:[{interface, weight}]}`, weight 0 to 10; no hidden interfaces | config card, `handleSetConfig`, `confirmSetWeight` |
| `MultiWAN.SetInterface` | `kmwan.set_interface` | `{interface, enable_check}`; with `enable_check` true also `track_proto`, `track_method`, `track_mode`, `track_ipv4` (proto 0 or 2), `track_ipv6` (proto 1 or 2), `enable_ssl` (method 1) | tracking dialog, `handleSetInterface` |
| `MultiWAN.SetSensitivity` | `kmwan.set_sensitivity` | `{sensitivity:{level, val}}`; level `low`, `medium`, `high` or `custom`; `val` 0.5 or 1 to 90 seconds, set only for `custom` | sensitivity dialog, `setSensitivity` |

Values from the web UI: mode 0 failover, 1 load balance; `track_mode` 1 normal, 0 low data (badge "passive"), 2 strict; `track_proto` 0 IPv4, 1 IPv6, 2 both (the selector shows only when IPv6 is enabled); `track_method` 0 ping, the only item of the list. The form has 4 host inputs for each protocol, the first 2 required, and sends the filled inputs. Interfaces in `system.get_info` `hidden_features` are not shown. The web UI warns when fewer than 2 interfaces have a weight that is not 0, but sends the write when the user accepts.

Side effects: a write can change the uplink that the router uses and so stop its Internet access. After `set_config` and `set_interface` the web UI waits 10 s, then reads the configuration and the system status again; after `set_sensitivity` it waits 1 s.

Open points: the meaning of `track_method` 1 (the web UI sends `enable_ssl` for it but does not list it) and of the track modes; what `val` does with a level other than `custom`; the result of `get_sensitivity` (needs a read on a router); if the writes keep the fields that they do not send (metrics in load balance mode, weights in failover mode, tracking fields when `enable_check` is false); what the router does with an interface that a `set_config` does not have, or with an unknown interface; what weight 0 does.
## Repeater and USB tethering (web UI code)

Source: the `internet` view (`internet.js`) and the Wi-Fi connect dialog in `app.js` of firmware 4.8.3. No write was tested on a router. The reads `repeater.get_config`, `repeater.get_saved_ap_list`, `repeater.get_status` and `tethering.get_status` have fixtures; `tethering.get_config` was not read on a router.

| Call | Params from the web UI | Client |
|---|---|---|
| `repeater.get_config` | `{}`; result `{auto, dfs_support, macaddr}`, `lock_band` when set; the UI ignores `status_interval` | `Repeater.GetConfig` |
| `repeater.set_config` | `{auto}` (0, 1 or 2) and `lock_band` only when it is not `""` | `Repeater.SetConfig` |
| `repeater.get_saved_ap_list` | `{}`; result `{res: [network]}`, with the Wi-Fi keys | `Repeater.ListSavedNetworks` |
| `repeater.remove_saved_ap` | `{ssid}` | `Repeater.RemoveSavedNetwork` |
| `repeater.connect` | `{ssid, protocol, disguise, manual, auto_portal, remember}`; `key` and `identity` when not empty; `macaddr: {mode, macaddr?, update?, period?}`, `ttl` and `ttl_ipv6` when `disguise` is false; `mtu` when set; `bssid` with the BSSID lock; `ip`, `netmask`, `gateway`, `dns` for `protocol` `static`. A click on a saved network sends the saved item with `remember: true`. In bridge mode only `{ssid, key?, identity?}`. UI timeout 30 s | `Repeater.Connect` |
| `repeater.disconnect` | `{}` | `Repeater.Disconnect` |
| `repeater.get_status` | read by the status poll of the UI (`repeater.status`); the UI also reads `eap`, `portal_url` and `ipv6` (`ip` is an array) | `Repeater.GetStatus` (no `ipv6`) |
| `repeater.scan` | `{all_band, refresh?}`, UI timeout 300 s; the UI shows errors -1 and -2 differently from other errors | not in the client |
| `repeater.enter_bare_mode`, `exit_bare_mode` | `{}`: the captive portal login mode | not in the client |
| `repeater.get_channel_prompt`, `set_channel_prompt` | `{chan_prompt_en?, popup_prompt_en?}`: hints of the UI | not in the client |
| `tethering.get_status` | `{}`; result `{status, devices: [{device, type, use}], ipv4?, ipv6?}`; status 0 not connected, 1 connected, 2 connecting | `Tethering.GetStatus` (no `ipv6`) |
| `tethering.get_config` | `{}`; result `{device, ttl?, ttl_ipv6?, mtu?}` | `Tethering.GetConfig` |
| `tethering.set_connect` | `{device}`, and `ttl`, `ttl_ipv6`, `mtu` when set | `Tethering.Connect` |
| `tethering.disconnect` | `{}` | `Tethering.Disconnect` |

The web UI has no call that saves or changes a saved network without a connect: "Modify" connects to the network with the new values.

Open points:

- Whether `set_config` without `lock_band` removes the band lock. The UI sends no `lock_band` for "all bands", and the test router, with all bands, returned no `lock_band`.
- Firmware 4.7.0 returns `auto` as a bool. Not known: whether it accepts a number in `set_config`.
- The meaning of `manual` in a saved network, and what `connect` with `remember: false` does to a saved network with the same SSID.
- What `remove_saved_ap` does when the repeater is connected to that SSID.
- The shape of `ipv6` in `repeater.get_status` and `tethering.get_status`, and the result of `repeater.scan`: no read on a router.
- The result of `tethering.get_config` on a router, and whether the router keeps the `set_connect` values when no phone is on USB.

A repeater write on a router whose uplink is the repeater can cut its Internet access: `connect` replaces the uplink network, and `disconnect` takes the uplink down.
## Parental control and Bark (web UI code)

`ParentalControl` calls `parental-control` and `Bark` calls `bark` (view `parentalcontrol`). The reads `get_config`, `get_status` and `get_app_list` are fixtures (no groups, only the preset rulesets `drop` and `accept`, `enable_app` false). The test router has no `bark` module (-32601). Each other call is from the web UI code of firmware 4.8.3 and the archived API reference, and was not tested on a router.

| Call | Params | Client method |
|---|---|---|
| `parental-control.set_config` | `{"enable", "drop_anonymous", "auto_update"}` | `SetConfig` |
| `parental-control.add_group` | `{"name", "default_rule", "macs"}`; result `{"id"}` | `AddGroup` |
| `parental-control.set_group` | `{"id", "name", "default_rule", "macs"}` | `SetGroup` |
| `parental-control.set_group` | `{"id", "schedules": [{"id"?, "week", "begin", "end", "rule"}]}`; `[]` removes all | `SetSchedules` |
| `parental-control.remove_group` | `{"id"}` | `RemoveGroup` |
| `parental-control.add_rule` | `{"name", "color", "apps", "manual": true, "blacklist"}` or `{..., "manual": false, "url"}`; result `{"id"}` | `AddRule` |
| `parental-control.set_rule` | as `add_rule`, with `"id"` | `SetRule` |
| `parental-control.remove_rule` | `{"id"}`; -2 default ruleset of a group, -3 temporary ruleset of a group | `RemoveRule` |
| `parental-control.get_brief` | `{"group_id"}` | `GetBrief` |
| `parental-control.set_brief` | `{"enable": true, "manual_stop", "rule_id", "group_id", "time"?}` or `{"enable": false, "group_id"}` | `StartBrief`, `StopBrief` |
| `parental-control.get_mode`, `set_mode` | `{"mode"}`: 1 parental control, 2 Bark | `GetMode`, `SetMode` |
| `bark.get_config`, `set_config` | `{"enable", "ipv6_enable"}` | `Bark.GetConfig`, `Bark.SetConfig` |
| `bark.get_status` | result `{"status", "pairing_code"}` | `Bark.GetStatus` |
| `bark.logout` | `{}` | `Bark.Logout` |

Open points:

- IDs of groups, rulesets and schedules are strings that the router makes. The reference examples look like UCI section names. Whether an ID stays the same after other items change was not tested. The reference says that ruleset names are unique; it says nothing about group names.
- Schedule `id`: the reference gives the type as number and shows strings; the client requires a string. Whether `begin` and `end` come back with seconds, and whether `"24:00"` is stored, is not known.
- The web UI sends back with `set_rule` the other fields that it read (`preset`, `url_success`); the client does not. That `get_config` returns `url` of a ruleset is an assumption.
- `add_rule` and `set_rule` error -2 also return `valid_list`, which `RPCError` does not keep.
- The fields of `get_status` groups (`id`, `rule`, `brief`, `schedule_id`) and of `get_brief` were not seen on a router.
- `get_mode` and `set_mode` are only called by the web UI on a router with Bark (`software_feature.bark`, firmware category `2c`, country `us`).
- Not built: `check_blacklist_online {"url"}` (the "detect" button of a URL ruleset; the type of `invalid_lines` is not known) and `update` (reference only; the web UI does not call it).

## Read shapes confirmed on the router (2026-11-18)

These reads were not in the probe. Each one was called one time on the GL-AXT1800 with firmware 4.8.3, and only the keys and the value types were recorded, not the values. There are no fixtures for them.

| Method | Result shape |
|---|---|
| `kmwan.get_sensitivity` | `{"sensitivity":{"level":string,"val":number}}` |
| `system.get_usb3_disable` | `{"usb3_disable":bool}` |
| `timer.get_screen` | `{"enable":bool,"turnon_hour":string,"turnon_min":string,"turnoff_hour":string,"turnoff_min":string,"week":[number]}` (7 items; the router has no screen) |
| `network.get_advance_config` | `{"nat_enable":number,"sip_enable":number}` |
| `cable.get_ports_config` | `{"ports":[{"name":string,"mode":string,"default_mode":string,"silk":string,"macaddr":{"mode":string,"macaddr":string,"defmacaddr":string}}]}` (3 ports; no `update` and no `period` in MAC mode `default`) |
| `cable.get_ports_status` | `{"ports":[{"name":string}]}` (3 ports) |
| `netmode.get_mode` | `{"mode":string}` |
| `wifi.get_mlo_config` | `{"res":{"random_bssid":bool,"encryptions":[string],"ifaces":[{"name":string,"mlo_enable":bool}]}}` (2 interfaces; the router has no Wi-Fi 7 radio, so the interfaces have no `mlo_band` and no SSID fields) |

## Writes tested on the router (2026-11-18)

- `fan.set_config {"temperature":60}` returned `[]`. The next `fan.get_config` had `temperature: 60`, and the fan started.
- `black_white_list.set_single_mac {"mode":"white","operate":"add","mac":<a made-up MAC>}` while the mode was `black`: the router accepted the write, `white_mac` then had the MAC, and the mode stayed `black`. The same call with `"operate":"del"` removed it. So the allow list can be filled before a change to allow-list mode. The opposite case (a `black` write in mode `white`) was not tested.

## Cellular modem (not built)

The client has no typed methods for the modules `modem` and `mvas`. The GL-AXT1800 has no modem: `modem.get_info` gives `{"modems":[],"offline_doc":false}`, `modem.get_status` gives `{"modems":[],"new_sms_count":0}`, and the other reads give empty results or `err_code` 20002001. No read shape of a modem is known.

What the web UI code of firmware 4.8.3 shows:

- Each call takes `bus` (and the SIM calls also `iccid`) from `modem.get_info` and `modem.get_status`. Their format is not known.
- `modem.get_sim_config {bus, iccid}` and `modem.set_sim_config {bus, iccid, config}` are the connection profile (`apn`, `auth`, `username`, `password`, `protocol`, `ip_type`, `roaming`, `ttl`, `ttl_ipv6`, `mtu`, the band selection). The write sends back each key of the read that the web UI does not use. These keys are not known, so a typed write could remove router settings.
- `modem.set_connect {bus}` and `modem.disconnect {bus}` start and stop the connection.
- Dual SIM (`get_slot_config`, `set_slot_config`), the traffic limit (`get_traffic_config`, `set_traffic_config`), APN polling, and the operator and cell tower locks have their own calls.

Before a typed API is possible, a router with a modem must show the results of `modem.get_info`, `modem.get_status` and `modem.get_sim_config`, and one `set_sim_config` call of its web UI.

