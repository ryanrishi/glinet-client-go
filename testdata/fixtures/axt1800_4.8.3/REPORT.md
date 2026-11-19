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
| `wg-server.get_config` | no `ipv6_enable` | `WireGuardServerConfig.IPv6Enable` is false; the cause is not known |
| `ovpn-server.get_config` | no `ipv6_enable` | not in the client |
| `wifi.get_config` | new top-level key `bandmode`, an empty array | not in `WifiConfig`, because the item type is not known |
| `repeater.get_config` | `auto` is a number (1), not a bool | not in the client |
| `switch-button.get_funcs` | `funcs` items are objects `{func, sub_func?:[{func}]}`, not strings; `vpn` with the tunnel IDs as `sub_func` replaces `wireguard` and `openvpn` | not in the client |
| `switch-button.get_config` | new key `sub_func` (a tunnel ID as a string when `func` is `vpn`) | not in the client |
| `local-access.get_config` | new key `ssh_enabled` (bool) | not in the client |
| `cloud.get_config` | new key `aw_api_url` (string) | not in the client |
| `modem.get_profile_list` | result is `{"err_code":20002001,"err_msg":"Parameter not found"}`, not a profile list (the router has no modem) | not in the client |
| `challenge` | new key `hash-method` | supported since `v0.10.2` |

All other ok methods have the same fields and types as on 4.7.0.

These differences come from the router state, not from the firmware:

- `firewall.get_dmz` has `priority` (bool). On 4.7.0 it appeared after the first `set_dmz` of the write test. `DMZ.Priority` (client `v0.12.0`) has it.
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

- `WireGuardServer.SetConfig` does not send `private_key: ""`. The private key input of the web UI shows "Auto Generate" when it is empty, but the web UI code does not settle what the router does with `""`. The client sends `ipv6_enable` with a non-empty `address_v6`, as the web UI of firmware 4.7.0 did.
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

- `cable.get_ports_config` (view `internet`, `getPortsConfig`, params `{}`). The web UI reads only `ports[]`, keeps the items whose `name` is `"wan"` or `"secondwan"`, and shows `silk` (the port label on the case). Other fields, other `name` values and the types are not in that code, and no fixture has the result.

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
