# Probe report: GL.iNet AXT1800, firmware 4.7.0

- Model: `axt1800` (GL-AXT1800, "Slate AX"), board "GL Technologies, Inc. AXT1800"
- Firmware: `4.7.0`, type `release8`, date 2025-03-19, OpenWrt 23.05-SNAPSHOT, kernel 5.4.164
- Probe date: 2026-11-12. Program: `cmd/probe` (read-only allowlist in `cmd/probe/calls.go`).
- Result: 173 read methods in the allowlist. 126 ok, 37 error, 10 skipped (no ID from a previous list response).

## Login

`challenge` returns `{"alg":1,"salt":...,"nonce":...}`. There is no `hash_method`, so the hash is MD5. `alg` 1 = MD5-crypt. `login` returns `{"sid":...,"username":"root"}`. See `challenge.json` and `login.json`.

## Error format

All errors in the probe were a JSON-RPC `error` object. No probe response had `err_code` inside `result`. Some write methods return `{"err_code":...,"err_msg":...}` as the result (see "timer LED schedule").

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

Many lists were empty on this router, so their item fields could not be compared: `firewall.get_port_forward_list.res` (later verified, see "firewall port forward"), `firewall.get_rule_list.res` (later verified, see "firewall rule"), `firewall.get_wan_access.whitelist`, `wg-server.get_peer_list.peers`, `wg-server.get_route_list.*_route_rules`, `wg-server.get_status.peers`. DMZ is disabled, so `firewall.get_dmz` has only `enabled` (no `dest_ip`; later verified, see "firewall DMZ"). Later phases must get fixtures with data for these (or use the reference shapes and mark them as not verified).

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
- `get_dmz`, `get_rule_list`, `get_wan_access`: no data to compare (see above). Top-level fields match.

### firewall port forward (write test, 2026-11-14)

Verified on the router with names `tf-test-N`, external ports `50001`-`50034`, and destination IPs `192.168.8.250` and `192.168.8.251` (no known client). The list was empty before and after the test. The fixtures `firewall.*port_forward*.json` are one sequence; sorted by request `id`, they replay the state changes in order (`glinettest/firewall_test.go` does this). A fixture with `"response": {"http_status": 500}` is an HTTP 500 response with an nginx HTML page and no JSON-RPC body.

| Call | Params | Result |
|---|---|---|
| `get_port_forward_list` | none | `{"res":[{id, name?, proto?, src?, src_dport?, dest?, dest_ip?, dest_port?, enabled}]}` |
| `add_port_forward` | `{name?, proto?, src, src_dport, dest, dest_ip, dest_port?, enabled?}` | `{"id":"cfg233837"}` |
| `set_port_forward` | `{id, ...}` | `[]` |
| `remove_port_forward` | `{id}` | `[]` |

Item fields:
- `id`, `name`, `proto`, `src`, `src_dport`, `dest`, `dest_ip`, `dest_port` are strings. Ports are strings also when they were written as numbers (`50002` gives `"50002"`), and can be a range (`"50004-50005"`).
- `proto` values used: `tcp`, `udp`, `tcp udp`.
- `enabled` is a JSON bool (not 0/1). On write, `true`/`false` and the numbers `0`/`1` work; the string `"false"` gives `true`. No `enabled` on add gives `true`.
- A field without a value has no key. `name`, `proto` and `dest_port` written as `""` have no key. Other fields were not written as `""`.

ID:
- The ID is the UCI name of an anonymous section: `cfg` + the section index in the firewall configuration (2 hex digits) + `3837` (from the section type). On this router the first port forward is `cfg233837`.
- The ID is not stable. After `remove` of a forward, each forward after it gets the ID of the forward before it (`firewall.get_port_forward_list.after_remove.json`). `add` appends at the end and does not change other IDs. `set` keeps the position and the ID.
- After all forwards are removed, the next `add` gets `cfg233837` again.

Behavior:
- `add` and `set` do not check values. Accepted: `proto: "foo"`, zones `"nope"`, no zones, `dest_ip: "not-an-ip"` and `"192.168.8.256"`, `src_dport: "abc"` and `"70000"`, `dest_port: "70000"`, no `src_dport`, no `dest_ip`.
- Duplicates are accepted: an `add` equal to an existing forward, and two forwards with the same external port and protocol, both get their own ID.
- `set` writes each field that the params have. A field that is not in the params keeps its value, except `name`: no `name` removes the name. `""` removes `name`, `proto` and `dest_port`.
- `set` with an unknown ID returns `[]` and changes nothing.
- `remove` with an unknown ID, or with no `id`, returns HTTP 500 and changes nothing.
- `remove` of a forward that has no `dest_port` returns HTTP 500 and changes nothing. After `set` of a `dest_port`, `remove` works. A forward without `proto` can be removed.
- Each write returned in about 20 ms. The next read worked about 5-10 ms later. No write made the router or the link slow or unreachable.

Errors:

| Response | Cause |
|---|---|
| `-32602 Invalid params`, `err: "id: string required"` | `set_port_forward` without `id` |
| HTTP 500 | `remove_port_forward` with an unknown ID, without `id`, or of a forward that has no `dest_port` |

Differences from the 2023 reference:
- `src_dport` and `dest_port` are strings in the list (reference: `dest_port` number). Both accept numbers on write.
- `add_port_forward` returns `{"id":...}` as the reference says; `set` and `remove` return `[]`, not `null`.
- `remove_port_forward` with an unknown ID, and of a forward without `dest_port`, is an HTTP 500, not a JSON-RPC error.
- The reference does not say that IDs change after a remove, that `set` without `name` removes the name, that values are not checked, or that duplicates are accepted.
- The reference marks `src`, `src_dport`, `dest`, `dest_ip`, `dest_port` as required. The router accepts `add` without them.

Not verified: `remove_port_forward` with `all` (remove all; not called); whether a write reloads the firewall (no API shows it, and no traffic was sent through a forward); whether the firewall uses a forward with invalid values; IDs after changes to zones (not written), which can change the index part of the ID (for rules see "firewall rule"); `""` for `src`, `src_dport`, `dest`, `dest_ip`; `enabled: null` and other value types; other `proto` values such as `icmp`.

### firewall rule (write test, 2026-11-14)

Verified on the router with names `tf-test-N`, source zone `wan` in each rule, destination IP `192.168.8.250` (no known client) and ports `50001`-`50024`. Each rule was disabled, or enabled only when it matched only traffic from `wan` to `192.168.8.250`. The rule list and the port forward list were empty before and after the test. The fixtures `firewall.*rule*.json` and the `firewall.*port_forward*.json` fixtures with request IDs 3000-3999 are one sequence (`glinettest/firewall_test.go` replays them, with a new fake router for each block of 1000 request IDs). `firewall.remove_*.cleanup_N.json` were written from the log of the test program, which did not save them as files.

| Call | Params | Result |
|---|---|---|
| `get_rule_list` | none | `{"res":[{id, name?, src?, src_ip?, src_mac?, src_port?, proto?, dest?, dest_ip?, dest_port?, target?, enabled}]}` |
| `add_rule` | `{name?, src?, src_ip?, src_mac?, src_port?, proto?, dest?, dest_ip?, dest_port?, target?, enabled?}` | `{"id":"cfg2392bd"}` |
| `set_rule` | `{id, ...}` | `[]` |
| `remove_rule` | `{id}` | `[]` |

Item fields:
- `id`, `name`, `src`, `src_ip`, `src_mac`, `src_port`, `proto`, `dest`, `dest_ip`, `dest_port`, `target` are strings. Ports are strings also when they were written as numbers (`50002` gives `"50002"`), and can be a range (`"50004-50005"`).
- Values used: `proto` `tcp`, `udp`, `tcp udp`; `target` `ACCEPT`, `REJECT`, `DROP`; `src` and `dest` `wan` and `lan`. A rule without `dest` (traffic to the router itself) and a rule without `proto` are accepted.
- `enabled` is a JSON bool. On write, `true`/`false` and the numbers `0`/`1` work; the string `"false"` gives `true`. No `enabled` on add gives `true`.
- A field without a value has no key. The router adds no default value: a rule added without `target` or `proto` has no `target` or `proto` key.

ID:
- The ID is the UCI name of an anonymous section: `cfg` + the section index in the firewall configuration (2 hex digits) + `92bd` (from the section type; port forwards have `3837`).
- Rules and port forwards are in the same index. On this router the first rule was `cfg2392bd`, at the index of the first port forward (before the first `set_dmz`; see "firewall DMZ"). A port forward added after a rule gets `cfg243837`. After `remove` of a rule, each rule and port forward after it gets the index before it, and after `remove` of a port forward, the same is true for the rules after it (`firewall.get_*_list.after_rule_remove.json`, `firewall.get_rule_list.after_forward_remove.json`).
- `add` appends at the end and does not change other IDs. `set` keeps the position and the ID.

Behavior:
- `add` and `set` do not check values. Accepted: `target: "FOO"`, `proto: "foo"`, `dest_port: "70000"`, `name: ""` (gives no `name` key), no `target`.
- Duplicates are accepted: an `add` equal to an existing rule gets its own ID.
- `set` writes each field that the params have. A field that is not in the params keeps its value, except `name`: no `name` removes the name. No `enabled` keeps `enabled`. `""` removes the field (verified for `name`, `src_ip`, `src_mac`, `src_port`, `proto`, `dest`, `dest_ip`, `dest_port`).
- `set` with an unknown ID returns `[]` and changes nothing.
- `remove` with an unknown ID, or with no `id`, returns HTTP 500 and changes nothing.
- `remove` of a rule that has no `dest_port` returns HTTP 500 and changes nothing. After `set` of a `dest_port`, `remove` works. A rule without `proto` can be removed.
- The methods do not check the section type of the ID: `remove_rule` with the ID of a port forward removes that port forward, `set_rule` with the ID of a port forward writes its `name`, and `remove_port_forward` with the ID of a rule removes that rule. Because the ID suffix comes from the type, an ID of a rule can only be the ID of a rule or of no section.
- There is no method to move a rule. Its position is the order of the add calls.
- Each write returned in about 15-25 ms. The next read worked about 5 ms later. No write made the router or the link slow or unreachable.

Errors:

| Response | Cause |
|---|---|
| `-32602 Invalid params`, `err: "id: string required"` | `set_rule` without `id` |
| HTTP 500 | `remove_rule` with an unknown ID, without `id`, or of a rule that has no `dest_port` |

Differences from the 2023 reference:
- `src_port` and `dest_port` are strings in the list (reference: numbers). Both accept numbers on write. A port can be a range.
- `target` is optional on add (reference: required).
- `set` and `remove` return `[]`, not `null`.
- `remove_rule` with an unknown ID, and of a rule without `dest_port`, is an HTTP 500, not a JSON-RPC error.
- The reference does not say that IDs change after a remove of a rule or a port forward, that `set` without `name` removes the name, that values are not checked, that duplicates are accepted, or that the methods accept the ID of another section type.

Not verified: `remove_rule` with `all` (remove all; not called); `""` for `src` and `target`; invalid values in `src`, `src_ip` and `src_mac` (not sent, so that no rule could match more traffic); other `proto` values such as `icmp` or `all`; `set_port_forward` with the ID of a rule; whether the firewall uses a rule with invalid values; whether a write reloads the firewall (no traffic was sent).

### firewall DMZ (write test, 2026-11-14)

Verified on the router with the DMZ host `192.168.8.250` (no known client), in the same session as "firewall rule". DMZ was disabled before and after the test, and was enabled for less than 1 s two times. The fixtures `firewall.*dmz*.json` are in the same sequence as the rule fixtures (request IDs 3069-3100); the rule and port forward fixtures after them show the effect on the IDs.

| Call | Params | Result |
|---|---|---|
| `get_dmz` | none | `{"enabled":false}` before the first `set_dmz`; after it `{"enabled":bool, "dest_ip"?:string, "priority":false}` |
| `set_dmz` | `{enabled?, dest_ip?}` | `[]` |

Behavior:
- `enabled` is a JSON bool. `set_dmz` without `enabled` enables the DMZ (`{"dest_ip":"192.168.8.250"}` gave `enabled: true`).
- `set_dmz` without `dest_ip` removes `dest_ip`, also when it disables the DMZ. `{"enabled":false,"dest_ip":...}` keeps the IP in a disabled DMZ.
- `dest_ip` is not checked: `"not-an-ip"` was accepted (with `enabled: false`).
- `priority` appears after the first `set_dmz` and stays `false`. The test did not write it.
- `set_dmz` does not add an item to the rule list or the port forward list, and does not change their IDs at once.

DMZ section and IDs:
- The first `set_dmz` adds a section without an ID at the end of the firewall configuration. It is not in the rule or port forward list, but it has an index. It stays when the DMZ is disabled. The test could not remove it, so it is still on the router: the next rule or port forward on an empty list gets index `0x24` (`cfg2492bd`, `firewall.add_rule.after_cleanup.json`), not `0x23` as before.
- `add_port_forward` moves the DMZ section to the end of the configuration after it gives the ID. So when the DMZ section was before the new forward, the ID that `add_port_forward` returns is one more than the ID of the forward in the next list (`firewall.add_port_forward.after_dmz.json`: returned `cfg2f3837`, listed `cfg2e3837`), and the rules and forwards that were after the DMZ section get the index before it (`firewall.add_port_forward.after_dmz_second.json`: rule `cfg3092bd` became `cfg2f92bd`).
- `add_rule` does not move the DMZ section (`firewall.add_rule.after_dmz_second.json`: the rule kept `cfg3092bd` for 17 s). A rule added when the DMZ section is not at the end gets the correct ID; `firewall.add_rule.after_dmz.json` returned `cfg2e92bd` because the DMZ section was at `0x2d` before the rule, and the rule moved to `cfg2d92bd` when the next `add_port_forward` moved the DMZ section.
- The move was seen on the next read, 5 ms after the write.

Differences from the 2023 reference:
- `enabled` is optional on `set_dmz` (reference: required) and its default is `true`.
- `get_dmz` has the extra field `priority`, after the first `set_dmz`.
- `set_dmz` returns `[]`, not `null`.
- The reference does not say that `dest_ip` is removed when it is not sent, or that the DMZ adds a firewall section that changes IDs.

Not verified: `set_dmz` with `enabled: true` and no `dest_ip`, or with an invalid `dest_ip` (not sent, so that the DMZ could not take traffic for an unknown host); `priority` on write and what it does; whether `set_port_forward`, `remove_port_forward`, `set_rule` or `remove_rule` move the DMZ section (in the test the DMZ section was at the end at each of these calls); whether the firewall forwards traffic to the DMZ host (no traffic was sent).

### led (write test, 2026-11-14)

Verified on the router. `led_enable` was `true` before and after the test. The fixtures `led.*.json` with request IDs 10000-10106 are one sequence (`glinettest/led_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_config` | none | `{"led_enable":bool}` |
| `set_config` | `{led_enable}` | `[]` |

Behavior:
- `led_enable` is a JSON bool. The router reads the value as a flag: no `led_enable` (also `{}`) gives `false`; `false` gives `false`; each other value gives `true`, also `0`, `1`, `"true"` and `"false"`.
- `led_enable: null` returns `-32602 Invalid params` with no `err` and changes nothing.
- Other keys are ignored.
- Each write returned in about 130-180 ms. The next read, 3 ms later, had the new value.
- A write of the LED schedule (`timer.set_led`, also enabled) did not change `led_enable`.

Differences from the 2023 reference: `set_config` returns `[]`, not `null`. The reference does not say that no `led_enable` disables the LEDs.

Not verified: the physical LEDs (no one looked at the router).

### timer LED schedule (write test, 2026-11-14)

Verified on the router. The schedule was `{"enable":false,"turnon_hour":"07","turnon_min":"00","turnoff_hour":"22","turnoff_min":"00","week":[0,1,2,3,4,5,6]}` before and after the test. The fixtures `timer.*_led*.json` are one sequence (request IDs 11000-11219; `glinettest/timer_test.go` replays them). The fixtures named `*.disabled_*` are the writes with `enable: false`.

| Call | Params | Result |
|---|---|---|
| `get_led` | none | `{"enable":bool, "turnon_hour":string, "turnon_min":string, "turnoff_hour":string, "turnoff_min":string, "week":[int]}` |
| `set_led` | `{enable, turnon_hour, turnon_min, turnoff_hour, turnoff_min, week}` | `[]`, or `{"err_code":-1,"err_msg":"parameter missing"}` |

Formats:
- Times are strings: hour and minute, in the router's local time (the time zone of `system.get_timezone_config`). The router has `"07"`, `"00"` (two digits); the web UI splits `"HH:MM"` and writes the parts. The router does not change a written value: `9` (number) and `"9"` give `"9"`, `"0"` gives `"0"`.
- `week` is an array of numbers. The web UI labels the days 0 Sunday, 1 Monday, ..., 6 Saturday. The router keeps the array as written: `[5,1,1]` stays `[5,1,1]`, and `[7]` stays `[7]`. Strings with numbers are written as numbers (`["1","2"]` gives `[1,2]`). An empty array gives `[0,1,2,3,4,5,6]`.

Behavior:
- `enable` is read as a flag: no `enable` and `false` are false; each other value is true, also `0` and `"false"`.
- When the write does not enable the schedule, the router sets `enable` to `false` and ignores all other fields, also when they are valid (`{"enable":false,"turnon_hour":"08",...}` changed nothing else). The web UI also sends `week` in this case.
- When the write enables the schedule, all five other fields are necessary. If one is missing, or a time is `""`, the result is `{"err_code":-1,"err_msg":"parameter missing"}` (a successful JSON-RPC response) and nothing changes. `{"enable":true}` alone also gives this error.
- No value is checked: `"25"`, `"60"`, `"ab"`, `[7]`, and the same on and off time were accepted and stored.
- `null` for any field returns `-32602 Invalid params` (no `err`) and changes nothing. `week` as a string (`"1,2"`) returns `-32602`, also when `enable` is `false`.
- Each write returned in about 5-20 ms. The next read, 4 ms later, had the new values.
- The schedule does not change `led.get_config`.

Differences from the 2023 reference:
- All fields are required only when `enable` is true. A write that disables the schedule keeps the times and days.
- The "parameter missing" error is in the result, not a JSON-RPC error.
- `set_led` returns `[]`, not `null`.

Not verified: whether the LEDs turn off and on at the set times (enabled for less than 1 s); the day numbering of the router itself (only the web UI labels); times that are not two digits, which the web UI does not write.

### timer reboot schedule (write test, 2026-11-14)

Verified on the router only with `enable: false`, so that no reboot could be scheduled. The schedule was `{"enable":false,"hour":"02","min":"15","week":[3,5]}` before and after the test. The fixtures `timer.*_reboot*.json` with request IDs 12000-12010 are one sequence (`glinettest/timer_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_reboot` | none | `{"enable":bool, "hour":string, "min":string, "week":[int]}` |
| `set_reboot` | `{enable, hour?, min?, week?}` | `[]` |

Formats: as in "timer LED schedule". `hour` and `min` are strings with two digits (`"02"`, `"15"`) in the router's local time. The web UI labels the days 0 Sunday to 6 Saturday, and sends `week` sorted.

Behavior:
- A write with `enable: false`, without `enable`, or `{}` sets `enable` to `false` and ignores `hour`, `min` and `week` (`{"enable":false,"hour":"05","min":"30","week":[2]}` changed nothing else).
- `enable: null` and `week` as a string (`"1,2"`) return `-32602 Invalid params` (no `err`) and change nothing.
- Each write returned in about 4 ms.

Not verified (no write with `enable: true` was sent): the enabled write and its errors, how `hour`, `min` and `week` are stored, and the day numbering of the router. A write with `enable: true` can only set the values; after it the router keeps them, so a test of it cannot restore the initial values without an enabled schedule that can reboot the router before the restore. The web UI sends `{enable: true, hour, min, week}` and only `{enable: false}` to disable, as for the LED schedule, so the enabled write probably works like `timer.set_led` (all fields necessary, else `err_code -1`). `glinettest` emulates it like that.

Differences from the 2023 reference: `hour`, `min` and `week` are not required (and are ignored) when `enable` is `false`. `set_reboot` returns `[]`, not `null`.

### dns hosts file (write test, 2026-11-14)

Verified on the router with names under `.invalid` and single-label names, and the IPs `192.0.2.1`, `192.0.2.2` and `192.168.8.250` (no known client). The hosts file was `"127.0.0.1 localhost\n\n::1     localhost ip6-localhost ip6-loopback\nff02::1 ip6-allnodes\nff02::2 ip6-allrouters\n"` before and after the test. The fixtures `dns.*_host*.json` with request IDs 13000-13405 are one sequence (`glinettest/dns_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_host` | none | `{"content":string}` |
| `set_host` | `{content}` | `[]` |

Behavior:
- `set_host` replaces the whole file with `content`. A write without the localhost lines removes them.
- The router keeps the content byte for byte: comments (`# ...`, also after a name), blank lines, tabs, more than one space, `\r\n`, two trailing newlines, IPv6 lines. It adds `\n` when the content does not end with `\n`; `""` gives `"\n"`.
- No line is checked: `not-an-ip name`, a line with one field and `192.0.2.256 name` were accepted. (The web UI checks that each non-empty line is an IP and at least one name, and does not allow comments.)
- `{}` and `content` as a number return HTTP 500 (no JSON-RPC body) and change nothing. `content: null` returns `-32602 Invalid params` (no `err`).
- Each write returned in about 7-25 ms. The next read, 3 ms later, had the new content.
- The DNS server of the router (`dig @192.168.8.1`) did not answer with a new name: `tf-test.invalid`, `tftesthost`, `tftesthost.lan` (192.0.2.x) and `tftesthost2` (192.168.8.250) were NXDOMAIN for at least 30 s after the write, and also after a `dns.set_config` write. `ip6-allnodes` from the initial file was answered. So `set_host` does not reload the DNS server, and it is not known when the new names take effect (probably at the next start of dnsmasq).

Differences from the 2023 reference: `set_host` returns `[]`, not `null`. The reference does not say that the content is not checked, that a newline is added, or that the DNS server is not reloaded.

Timing test (2026-11-14, request IDs 17000-17004, fixtures `dns.*_host.timing_*.json`): one line `192.0.2.1 tf-test-host.invalid tf-test-host.lan` was added to the initial file. `dig @192.168.8.1` for `tf-test-host.invalid` and `tf-test-host.lan` (A record) was NXDOMAIN at each query, every 15 s from 0 s to 300 s after the write. So the router does not answer with a name from a new hosts file within 5 minutes, for a `.invalid` name and for a `.lan` name. No service was restarted and no DNS setting was changed. The exact initial string was written back and read back after the test.

Not verified: whether a reboot, a restart of dnsmasq or another write makes the DNS server use the new file; the maximum size.

What the web UI does when the user saves the hosts file (from the UI code, not tested). Source: `/views/gl-sdk4-ui-dnsview.common.js.gz` of the router web server (`/views/gl-sdk4-ui-dnsview.common.js` redirects with 302), 77187 bytes unpacked, read on 2026-11-14.
- The API object of the DNS view has only these methods: `dns.get_info {}`, `dns.get_config {}`, `dns.set_config`, `dns.get_host {}` and `setDnsHost(t)`, which is `$rpcRequest("call", ["sid", "dns", "set_host", t], {errTips: {keep: true}})`.
- `handleApplyHost` checks each line with `validateHost`. If all lines are valid, it calls `setDnsHost({content: this.hostStr})`. On success it closes the dialog and calls `getDnsHost()`, which is `dns.get_host {}` and puts `content` in the text area. On failure it only closes the message.
- No other call follows `set_host`: no `set_config`, no reload, restart or apply method. `app.js` has no `set_host`. The shared `$rpcRequest` adds no call after a request; it only rejects a result with a top-level `err_code` or `err_msg`.
- So the UI does nothing more than the client `SetHosts`. It does not explain why the new names did not resolve.

### dns config (write test, 2026-11-14)

Verified on the router in mode `auto`, with a WireGuard client tunnel that supplies DNS (`wg 10.66.0.1` in `server`). The config was `{"mode":"auto","force_dns":false,"rebind_protection":false,"override_vpn":true,"dot_provider":"0","nextdns_id":"","controld_id":"","free_controld_id":"","server":["wwan 192.168.100.53","wg 10.66.0.1"]}` before and after the test, and `server` did not change during it. Only `rebind_protection` and `force_dns` were changed on purpose, one at a time. The fixtures `dns.*_config*.json` with request IDs 14000-14315 are one sequence (`glinettest/dns_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_config` | none | `{"mode":string, "force_dns":bool, "rebind_protection":bool, "override_vpn":bool, "server":[string], "dot_provider":string, "nextdns_id":string, "controld_id":string, "free_controld_id":string}` |
| `set_config` | `{mode?, force_dns, rebind_protection, override_vpn}` | `[]` |

Formats:
- In mode `auto`, `server` is the upstream DNS servers that the router learned, one string `"<interface> <ip>"` for each. It is runtime data.
- `dot_provider` is a string (`"0"`). The web UI reads `"1"` as NextDNS and `"2"`/`"3"` as ControlD plans.

Behavior:
- `force_dns`, `rebind_protection` and `override_vpn` are read as flags: no value gives `false`, `false` gives `false`, each other value gives `true` (`rebind_protection: 0` gave `true`). The first test, `{"rebind_protection":true}` alone, also set `override_vpn` from `true` to `false`; it was restored 12 s later. `server` and the DNS answers did not change in that time.
- No `mode` keeps the mode (`{"force_dns":false,"rebind_protection":false,"override_vpn":true}` kept `auto`). This is what the web UI sends when AdGuard Home is enabled; otherwise it sends `mode` too, and the fields of that mode (`server` for `manual` and `proxy`; `proto`, `dot_provider` and an ID for `secure`).
- `null` for a field returns `-32602 Invalid params` (no `err`) and changes nothing.
- Each write returned in about 0.6-1.1 s. The next read, 7 ms later, had the new values. DNS queries to the router (every 250 ms) were answered during each write; the WireGuard client stayed connected.
- A `set_config` while AdGuard Home was enabled returned `[]` (no "DNS is occupied" error).

Differences from the 2023 reference:
- No `proto` in mode `auto`. Extra fields: `override_vpn`, `dot_provider`, `controld_id`, `free_controld_id`.
- `mode` is optional. The booleans are not optional: a missing boolean is written as `false`.
- `set_config` returns `[]`, not `null`.

Not verified: modes other than `auto` (the active tunnel supplies DNS, so no mode change was tested); `server`, `proto`, `dot_provider` and the IDs on write; an invalid `mode`; the errors `-1` (DNS is occupied) and `-2`/`-3` of the reference and the web UI; a write of `override_vpn` on purpose and what it does with the tunnel DNS.

### dns
- `get_config`: no `proto`. Extra: `controld_id`, `dot_provider` (string, `"0"`), `free_controld_id`, `override_vpn`. `server` items are strings like `"wwan 192.168.100.53"`.
- `get_host`, `get_info`: match.

### adguardhome
- `get_config`: extra `dns_enabled` (reference has it only as `set_config` input).

### adguardhome (write test, 2026-11-14)

Verified on the router. The config was `{"enabled":false,"dns_enabled":true}` before and after the test. AdGuard Home was enabled two times, for 13 s and 38 s. The fixtures `adguardhome.*.json` with request IDs 15000-15501 are one sequence (`glinettest/adguardhome_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_config` | none | `{"enabled":bool, "dns_enabled":bool}` |
| `set_config` | `{enabled, dns_enabled?}` | `[]` |

Behavior:
- No `enabled` disables AdGuard Home: `{"dns_enabled":true}` while it was enabled gave `enabled: false`.
- No `dns_enabled` keeps `dns_enabled` (`{}` and `{"enabled":false}` kept `true`). `dns_enabled` can be changed while AdGuard Home is disabled. The web UI sends `{enabled: true, dns_enabled}` to enable and only `{enabled: false}` to disable.
- `enabled: null` returns `-32602 Invalid params` (no `err`) and changes nothing. Other value types were not sent.
- A write that enables AdGuard Home returned in about 270 ms, and `get_config` had the new values at once. The service started later: `system.get_status` showed `adguard` status `0` 1 s after the write and `1` 38 s after it. The DNS server of the router did not answer (`dig` timeout) from 1 s after `{"enabled":true,"dns_enabled":false}` until 3 s after the next disable (16 s); in the second test, after `{"enabled":true,"dns_enabled":true}`, one query 8 s after the write timed out and the others were answered. The web UI waits 6 s after it enables and 3 s after it disables.
- A write that disables it returned in about 20-50 ms. DNS queries were answered after it.
- No write returned an error. The "other DNS not turned off" error (`err_code 1` in the result, from the web UI and the reference) did not occur in DNS mode `auto`.

Differences from the 2023 reference: `dns_enabled` is in the `get_config` result. `enabled` is not required (no value disables), `dns_enabled` is not required (no value keeps it). `set_config` returns `[]`, not `null`.

Not verified: the error when another DNS feature is active (that needs a DNS mode change); whether AdGuard Home answers queries after it started (it was disabled after 38 s); what `dns_enabled: false` does while AdGuard Home runs.

### wifi
- `get_config`: extra `res[].random_bssid`. `res[].channels` items are objects `{channel, dfs, psc}`. `res[].htmodes` is an object (mode → max bandwidth, plus `auto`), as the reference says.
- `get_status`: matches.

### clients
- `get_list`: extra per client: `ipv6`, `last_rx`, `last_tx`, `last_update_rate`, `total_rx_init`, `total_tx_init`.
- `get_status`: matches.

### black_white_list block list (write test, 2026-11-14)

Verified on the router in mode `black` only, with made-up MACs `02:00:00:00:00:a1` to `a9` (no real client). Mode `white` was never written. The config was `{"mode":"black"}` (empty list) before and after the test, and the mode stayed `black` after each write. The fixtures `black_white_list.*.json` with request IDs 16000-16060 are one sequence (`glinettest/accesscontrol_test.go` replays them; the `clients.block_client` calls 16045-16058 and their reads are not in the fixtures, see below).

| Call | Params | Result |
|---|---|---|
| `get_config` | none | `{"mode":string, "black_mac":[string]}`; no `black_mac` when the list is empty |
| `set_single_mac` | `{mode, operate, mac}` | `[]` |
| `set_config` | `{mode, mac:[string]}` | `[]`, or `{"err_code":-1,"err_msg":"mac required"}` |

Formats:
- `black_mac` is the block list, an array of strings in the order of the writes. The web UI code also reads `white_mac` (the allow list); it was not seen, because mode `white` was not tested.
- The router keeps each MAC as written, also its letter case (`"02:00:00:00:00:A2"` stays upper case).

Behavior of `set_single_mac` (the web UI uses it for the block switch of one client, with `mode` = the current mode):
- `operate: "add"` appends `mac`. A MAC that is in the list with the same letter case is not added again; the same MAC in the other letter case is added as a second entry.
- `operate: "del"` removes `mac` only with the same letter case (`del 02:00:00:00:00:a2` did not remove `02:00:00:00:00:A2`). A MAC that is not in the list gives `[]` and no change.
- `mac` is not checked: `"not-a-mac"` was added. A MAC with `-` separators (`"02-00-00-00-00-a3"`) was not added, and `[]` was returned.
- No `mac`, an unknown `operate` (`"toggle"`) or no `operate` gives `[]` and no change. `mac: null` gives `-32602 Invalid params` (no `err`).
- Each write returned in about 10-50 ms. The next read, 4 ms later, had the new list.

Behavior of `set_config` (the web UI uses it for the access control dialog, with the MACs in upper case):
- It replaces the whole list with `mac`, as written: duplicates (`[a4, a4, A4]`), `"not-a-mac"` and `-` separators were kept. `mac: []` empties the list (`get_config` then has no `black_mac`).
- No `mac` or `mac` as a string gives the result `{"err_code":-1,"err_msg":"mac required"}` and no change. `mac: null` gives `-32602`.
- Each write returned in about 10 ms. The next read had the new list.

`clients.block_client {mac, block}` (not in the web UI; named in the 2023 reference and `gli4py`):
- It exists on 4.7.0 and returns `[]` after about 850-890 ms, also for `"not-a-mac"` and without `block`.
- It changes the same list (`black_mac`), in upper case (`a6` was stored as `"02:00:00:00:00:A6"`), but not reliably and not at once: in one test the MAC was in the list 6 ms after the call; in another, `block: true` for `a7` (two times) was not in the list after the call, but `"02:00:00:00:00:A7"` and `"NOT-A-MAC"` appeared some seconds later, after a `set_config` had emptied the list; in a third test, `block: true` for `a7` did not change the list within 15 s. `block: false` with either letter case did not remove `"02:00:00:00:00:A6"`. The list was emptied with `set_config` and then stayed empty for 2.5 minutes, also after `block: false` for each test MAC.
- So the client does not use `clients.block_client`.

Same list or two features: one list. `clients.block_client` writes the `black_mac` list of `black_white_list`, and the web UI shows a client as blocked when its MAC is in `black_mac` (compared without letter case) and does not read the `blocked` field of `clients.get_list`. Not verified: whether `clients.get_list[].blocked` is computed from `black_mac` (that needs the MAC of a real client in the list). The made-up MACs did not appear in `clients.get_list`, and the `blocked` flags of the real clients stayed `false`.

Not verified: that the router really blocks a MAC (no real client was blocked), how long the firewall takes to apply a write, mode `white` and `white_mac`, what `set_single_mac` and `set_config` do with `mode: "white"` or no `mode` (never sent, because mode `white` cuts the access of all clients that are not in the list), the maximum list size (the web UI allows 1000).

### ddns (write test, 2026-11-14)

The owner of the router allowed DDNS to be enabled for a short time. DDNS was enabled for 27 s (14:37:50-14:38:17). No access over the WAN was enabled. The config was `{"enable_ddns":false, "device_id":<redacted>}` before and after the test. The fixtures `ddns.*.json` with request IDs 21000-21010 are one sequence (`glinettest/ddns_test.go` replays the config fixtures).

| Call | Params | Result |
|---|---|---|
| `get_config` | none | `{"enable_ddns":bool, "device_id":string}` |
| `get_status` | none | `{"ddns"?:string, "ips":[{"interface":string, "ip":[string]}], "status":int}` |
| `set_config` | `{enable_ddns}` | `[]`, or `{"err_code":-1,"err_msg":"missing parameters"}` |

- `device_id` is the device ID for the GL.iNet DDNS name (the same value as `system.get_info.ddns`). It is a secret and is redacted in the fixtures.
- `set_config {"enable_ddns":true}` and `{"enable_ddns":false}` each returned `[]` in about 30 ms. The next `get_config`, 15 ms later, had the new value, and `system.get_status.system.ddns_enabled` had it at once.
- `set_config {}` (sent while DDNS was enabled) returned `{"err_code":-1,"err_msg":"missing parameters"}` and changed nothing.
- `get_status` takes about 0.6 s. `status` was 2 before, during and after the test (meaning not known). About 5 s after the enable, the result had a new key `ddns`: an IPv4 address (a public one, redacted in the fixtures). It was gone 2 s after the disable.
- While DDNS was enabled, these did not change: firewall rule list, port forward list, DMZ, zone list, WAN access (`get_wan_access`), DNS config, static binds, block list, hosts, LED, timers, AdGuard Home, time zone, the services in `system.get_status` and the WireGuard client (connected, same peer).

From the UI code, not tested (`/views/gl-sdk4-ui-dynamicdns.common.js.gz`, read on 2026-11-14):
- The view calls `ddns.get_config {}`, `ddns.get_status {}` (the "DDNS test"), `ddns.set_config` and `firewall.get_wan_access {}`.
- The write is `ddns.set_config {"enable_ddns": bool}`, nothing else. The UI enables it only after the user accepts the terms. It reads `get_config` again 3 s after the write. A result with `err_msg` is shown as an error (`err_code` + `err_msg`).
- The UI reads `ddns`, `ddnsv6`, `status` and `ips` from `get_status`. The access over the WAN (`enable_ssh_access`, `enable_http_access`, `enable_https_access` of the 2023 reference) is not in the DDNS write; the view shows `firewall.get_wan_access` (`enable_https`, `enable_ssh`) instead.

Not verified: non-boolean values of `enable_ddns`, `ddnsv6` in `get_status`, the meaning of `status`, that the DDNS name resolves (no DNS lookup was made), DDNS on a router with a public WAN address.

### tailscale (write test, 2026-11-14)

First test (request IDs 19000-19039): only `enabled: false`. Second test (request IDs 22000-22011, with the permission of the owner of the router): Tailscale was enabled for 38 s (14:38:50-14:39:28). `get_auth_url`, `logout` and every login method were not called, so the router was never logged in to a tailnet. The config was `{"enabled":false,"lan_enabled":false,"wan_enabled":false,"lan_ip":"192.168.8.0/24"}` before and after each test, and `system.get_status` had the `tailscale` service at status `0` after each test. The fixtures `tailscale.*_config*.json` of each block are one sequence (`glinettest/tailscale_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_config` | none | `{"enabled":bool, "lan_enabled":bool, "wan_enabled":bool, "lan_ip":string, "exit_node_ip"?:string}` |
| `set_config` | `{enabled, lan_enabled, wan_enabled, exit_node_ip?}` | `[]` |
| `get_status` | none | `{"dns":[string]}` while disabled (takes 2 s); `{"dns":[string], "status":1}` while enabled and not logged in (takes 45 ms) |
| `get_exit_node_list` | none | `[]` while disabled and while enabled and not logged in |

Behavior with `enabled: false`:
- The router stores `lan_enabled`, `wan_enabled` and `exit_node_ip` also while Tailscale is disabled; the next read, 5 ms later, had them.
- `enabled`, `lan_enabled` and `wan_enabled` are set to the sent values: a missing one is written as `false` (`{"enabled":false}` and `{}` set `lan_enabled` from `true` to `false`). `lan_enabled: 1` gave `true`. `lan_enabled: null` returned `-32602 Invalid params` (no `err`) and changed nothing.
- `exit_node_ip` (`"100.64.0.1"`) was stored and is then in `get_config`. `exit_node_ip: ""` and a write without `exit_node_ip` remove it (the key is then not in `get_config`).
- `lan_ip` is not changed by a write.
- Each write returned in about 5-16 ms.

Behavior with `enabled: true` (the writes of the web UI):
- `{"enabled":true,"lan_enabled":false,"wan_enabled":false,"exit_node_ip":""}` returned `[]` in 17 ms. The next `get_config`, 30 ms later, had `enabled: true`. The first `get_status`, sent 50 ms after the write (it took 1.7 s), had `status: 1` ("login needed" in the UI code), and it stayed 1. `system.get_status` had the `tailscale` service at status `2`.
- A second enabled write with `lan_enabled: true` returned `[]` in 14 ms and was in `get_config` at once.
- The disable `{"enabled":false}` (sent with `lan_enabled: true` stored) returned `[]` in 13 ms. `get_config` then had `lan_enabled: false`: a disable does not keep the other fields. One `get_status` 50 ms after the disable still had `status: 1`; the next, 1 s later, had no `status`. The service status in `system.get_status` was `0` at once.
- While Tailscale was enabled (also with `lan_enabled: true`), these did not change: firewall zone list, rule list, port forward list, DMZ, WAN access, DNS config (`server` list), `get_status.dns`, static binds, block list, hosts, LED, timers, AdGuard Home, time zone, `system.get_status.system` flags and the WireGuard client (connected, same peer). After the disable, all guards were as before the test.

From the UI code, not tested (`/views/gl-sdk4-ui-tailscaleview.common.js.gz`, read on 2026-11-14):
- To disable, the UI sends only `{"enabled":false}`. To enable, it sends `{"enabled":true, "lan_enabled", "wan_enabled", "exit_node_ip":""}`. To set an exit node, it sends `{"enabled", "lan_enabled", "wan_enabled", "exit_node_ip"}`. It reads `get_config` again 5 s after a write.
- When enabled, the UI reads `status` (1 login needed, 4 connecting; 3 is shown as running in the service list), `login_name`, `address_v4` and `dns` from `get_status`, and `exit_node_list` from `get_exit_node_list`. It calls `get_auth_url` (timeout 30 s, result `auth_url`, `err_code -1` = no URL) for the login, and `logout`.

Not verified: the login, `status` values other than 1, `login_name` and `address_v4`, the exit node list items, what `lan_enabled` and `wan_enabled` do in the firewall when the router is in a tailnet (the zone list did not change, but the firewall sections that the API does not show were not read), `wan_enabled: true` while enabled.

### zerotier (write test, 2026-11-14)

First test (request IDs 20000-20017): only `enabled: false`, no network ID. Second test (request IDs 23000-23021, with the permission of the owner of the router): ZeroTier was enabled with the made-up network ID `0123456789abcdef` (it joins no real network) for 58 s (14:40:10-14:41:08) and again for 2.4 s (14:41:46-14:41:49) to restore `lan_enabled`. The config was `{"enabled":false,"lan_enabled":false,"wan_enabled":false}` before the second test and is `{"enabled":false,"id":"0123456789abcdef","lan_enabled":false,"wan_enabled":false}` after it: no write removes the ID (see below). `system.get_status` had the `zerotier` service at status `0` after each test. The fixtures `zerotier.*_config*.json` of each block are one sequence (`glinettest/zerotier_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_config` | none | `{"enabled":bool, "id"?:string, "lan_enabled":bool, "wan_enabled":bool}` (no `id` before the first write of an ID) |
| `set_config` | `{enabled, id, lan_enabled, wan_enabled}` | `[]`, `{"err_code":1,"err_msg":"parameter error"}`, or HTTP 500 |
| `get_status` | none | `[]` while disabled; `{"status":int}` while enabled |

Behavior with `enabled: false`:
- A write that does not enable ZeroTier only disables it. It ignores `id`, `lan_enabled` and `wan_enabled`: `{"enabled":false,"lan_enabled":true,...}`, `{"enabled":false,...,"wan_enabled":true}`, `lan_enabled: 1`, `{"enabled":false}`, `{}` and `{"enabled":false,"id":"","lan_enabled":false,"wan_enabled":false}` did not change them, also 15 s later. (This is unlike `tailscale`, which stores them.)
- The disable `{"enabled":false}` keeps `id`, `lan_enabled` and `wan_enabled`: after the disable, `get_config` had `{"enabled":false,"id":"0123456789abcdef","lan_enabled":true,"wan_enabled":false}`.
- `lan_enabled: null` returned `-32602 Invalid params` (no `err`).
- Each write returned in about 0.95-1 s, a disable of a running ZeroTier in 1.6-1.8 s (a write that fails with `-32602` in 4 ms).

Behavior with `enabled: true`:
- `{"enabled":true,"id":"0123456789abcdef","lan_enabled":false,"wan_enabled":false}` returned `[]` after 10.7 s (the first join). Later enabled writes returned after 2.4 s. The next `get_config` had the sent values; `id` is the string as written.
- An enabled write stores `lan_enabled` and `wan_enabled`: `lan_enabled: true` was in `get_config` at once.
- `get_status` was `[]` 20 ms after the first enabled write, and `{"status":1}` ("connecting" in the UI code) from 1 s later until the disable (47 s). Status 3 ("network not found") did not appear. `system.get_status` had the `zerotier` service at status `2` while enabled. After the disable, `get_status` was `[]` and the service status `0` at once.
- `id: ""` returned `{"err_code":1,"err_msg":"parameter error"}` in 6 ms and changed nothing. A write without `id` returned HTTP status 500 (nginx error page) and changed nothing, also with an ID stored. `id: null` returned `-32602`. So a stored ID cannot be removed with `set_config`.
- While ZeroTier was enabled (also with `lan_enabled: true`), these did not change: firewall zone list, rule list, port forward list, DMZ, WAN access, DNS config, static binds, block list, hosts, LED, timers, AdGuard Home, time zone, `system.get_status.system` flags and the WireGuard client (connected, same peer). After the test, all guards were as before, except the stored `id` in `zerotier.get_config`.

From the UI code, not tested (`/views/gl-sdk4-ui-zerotierview.common.js.gz`, read on 2026-11-14):
- To disable, the UI sends only `{"enabled":false}`. To enable, it sends `{"enabled":true, "id", "lan_enabled", "wan_enabled"}`, with a timeout of 30 s, and reads `get_config` again 5 s after the write. The UI accepts `id` only as 16 hex digits.
- The UI reads `id`, `enabled`, `lan_enabled` and `wan_enabled` from `get_config` (each may be missing), and `status` (0 connected, 1 connecting, 2 not authorized, 3 network not found), `zerotier_ip`, `lan_ip`, `wan_ip`, `wwan_ip` and `secondwan_ip` from `get_status`.

Not verified: a real ZeroTier network (join, status 0, 2 or 3, `zerotier_ip` and the other addresses of `get_status`), an `id` that is not 16 hex digits, `wan_enabled: true` while enabled, what `lan_enabled` and `wan_enabled` do in the firewall when ZeroTier is connected (the firewall sections that the API does not show were not read), how to remove a stored ID (only `set_config` was tried).

### wg-server
- `get_status`: server fields are nested in `server{initialization,log,rx_bytes,status,tunnel_ip,tx_bytes}`. The reference has them at top level. `peers` is a top-level list (empty here).
- `get_config`, `get_setting`: match.
- `get_peer_list`, `get_route_list`: empty lists; item fields not verified.

### network
- `routes`, `routes6`, `get_dhcp_leases`, `get_hwnat_config`: -32601 Method not found.
- `check_wan_cable`: returns `cable_enabled`, `cable_inserted`, `macclone_enabled`.
- `get_netnat_config`: extra `conflict_features`, `qos_enabled`, `wifi_reload`.
- `get_arp_list`: matches.

## Network configuration

### Write test of 2026-11-14 (stopped after the first write)

The owner approved a supervised write test of `lan` (guest network only), `cable` (protocol `dhcp` only), `ipv6` (`{"enable":false}` only) and `wifi` (radio1 only), in this order. The test tool had an allow-list in code that refused every other write before it was sent, and it ran a health check before and after each write: `repeater.get_status`, `lan.get_wan_info`, `system.get_status`, `wg-client.get_status`, `dns.get_config`, `wifi.get_status` and the `lan` item of `lan.get_config_list`, every 2 s.

Only one write was sent: `lan.set_config` with `{"interface":"guest","enable":0}`, while the guest DHCP server was already off (fixtures `lan.*.guest_dhcp_disable_noop.json`, IDs 24002-24005):

- It returned `[]` in 54 ms. The result had no `api_extra`.
- `get_config_list` right after the write and 3 min later was equal to the read before: both items, with `leasetime` `"12h"`, `lpr` `[]` and `transfer_enable` `0`.
- At 2.3 s the health check was unchanged. At 9.4 s `wifi.get_status` had `radio0` on channel 153 (before: 48, the channel of the AP of the repeater uplink), `wwan` was offline in `system.get_status`, `lan.get_wan_info` was empty and `dns.get_config` `server` had lost its `wwan` entry. From 11.6 s the repeater had no SSID, BSSID and address, and from 13.7 s its `state_s` was `failed` or `connecting`. `wifi.get_status` showed a different value for `radio1` in two samples (15.8 s and 22.3 s). So the router restarted the Wi-Fi about 9 s after the write.
- `wifi.get_config` after the write (fixture `wifi.get_config.after_guest_dhcp_write.json`) had `radio0` with `channel` 153 and `htmode` `"20"`; before the write it was 48 and `"40"`, the values of the repeater uplink. The interface settings were unchanged.
- The WireGuard client went from status 1 to status 2 (connecting) between 3 and 5 min after the write, because it had no uplink.
- The `lan` item, the management connection on Ethernet and the system uptime were not affected.
- The repeater uplink did not connect again by itself: 16 min after the write the repeater was still `failed` or `connecting`, `radio0` was on channel 153, and the WireGuard client had no uplink. The test sent no repeater or Wi-Fi write to repair it.

The uplink was not back after 3 minutes, so the test stopped by its rule: no other write was sent. The `cable`, `ipv6` and `wifi` writes, the other `lan` writes and the error `-8` (not sent in any case: the UI code does not show that the router checks the overlap before it applies the write) remain not tested.

### Network configuration from the UI code

Except for the one write above, the write methods in this section were never called on a router. The parameters, the call sequence and the validation rules come from the JavaScript of the web UI of the test router (firmware 4.7.0), which the router serves as static files: `/views/gl-sdk4-ui-<view>.common.js.gz`. The views are `internet` (WAN), `ipv6`, `lanip` and `guestip` (LAN and guest network, one shared component with `netInterface` `lan` or `guest`), and `wireless`. The reads are verified by the fixtures `cable.get_config.json`, `ipv6.get_ipv6.json`, `lan.get_config_list.json` and `wifi.get_config.json` (probe of 2026-11-12). Each write test proposal below needs the owner at the router.

### cable WAN config (from the UI code, not tested)

Read (verified): `cable.get_config` with `{}` returns `{"protocol":"dhcp"}` on the test router. The wired WAN port has no cable; the uplink is the Wi-Fi repeater (`wwan`).

The web UI reads with `{secondwan: false}` and uses these fields of the result (not verified): `protocol, ipv6, ipv4, username, password, vlanid, ttl, ttl_ipv6, mtu`. For `dhcp` it shows the runtime addresses from `cable.status`, not from `get_config`.

Write (view `internet`, component `CableSettingDialog`): one `cable.set_config` with the full configuration of the protocol, then the UI waits 4 s:

```js
r={protocol:t};if(e&&(r.vlanid=e),a&&(r.ttl=a),n&&(r.ttl_ipv6=n),o&&(r.mtu=o),"pppoe"===t)r.username=i,r.password=s;else if("static"===t){...r.ipv4={...i,dns:s},this.ipv6Enable&&this.ipv6Required){...r.ipv6=s}}
```

The page adds `secondwan: "secondwan"===this.currentWan` to the params. So the params are:

| Protocol | Params |
|---|---|
| `dhcp` | `{protocol, vlanid?, ttl?, ttl_ipv6?, mtu?, secondwan}` |
| `static` | `{protocol, ipv4{ip, netmask, gateway, dns[1-2]}, ipv6{ip, gateway, dns[1-2]}?, vlanid?, ttl?, ttl_ipv6?, mtu?, secondwan}` |
| `pppoe` | `{protocol, username, password, vlanid?, ttl?, ttl_ipv6?, mtu?, secondwan}` |

Validation in the UI: `protocol` is one of `dhcp`, `static`, `pppoe`. `vlanid` is a number 1-4094 (`e<1||e>4094`). `ttl` and `ttl_ipv6` are 1-255. `mtu` is 576-9200 (regex `^(57[6-9]|...|9200)$`); the UI warns when IPv6 is enabled and `mtu` is less than 1280. Static: `ipv4.ip` and `ipv4.gateway` pass `checkStaticIp`, `ipv4.netmask` passes `isMask`, `dns1` is required and `dns2` is optional. `ipv6` is sent only while IPv6 is enabled (`system.ipv6_enabled`) and one of its fields is set; then `ipv6.ip` is a CIDR and gateway and `dns1` are IPv6 addresses. PPPoE `username` has at most 256 characters; `username` and `password` have no other check (the username field says "optional"). The error `-1` is shown as `cable.vlan_id_conflict_tips` (VLAN ID conflict). The write button is shown only in router mode (`system.mode` 0).

MAC clone is not part of this write (it is in another view; `macclone.get_mac` is -32601 on 4.7.0).

Unknown: the types and presence of the optional fields in the `get_config` result (the UI treats `vlanid`, `ttl`, `mtu` as numbers); whether `{}` and `{"secondwan":false}` give the same read; whether a write keeps fields that it does not send (the UI always sends the full configuration, so the client does too); how long the router takes to apply it; the errors other than `-1`.

Proposal for a supervised write test: write `{"protocol":"dhcp","secondwan":false}` (the current value) and read back; then `{"protocol":"dhcp","mtu":1500,"secondwan":false}` and read back to learn the type of `mtu`; restore with `{"protocol":"dhcp","secondwan":false}` and check that `mtu` is gone. Do not test `static` or `pppoe` unless no cable is in the WAN port. Risks: a write probably restarts the `wan` interface; the WAN has no cable and is not the uplink, so the uplink (`wwan`) should not drop, but `kmwan` gives `wan` metric 10 before `wwan` 20, so a wrong static gateway could matter when a cable is plugged in later. Watch `lan.get_wan_info` and `system.get_status` `network[]` for `wwan` during the test.

Not tested: the write test of 2026-11-14 stopped before the `cable` writes. Still not verified: every `cable.set_config` call (also the no-change `dhcp` write), the read form of `mtu` and `ttl`, whether a write replaces the configuration, the errors, and the effect on the `wwan` uplink.

### ipv6 (from the UI code, not tested)

Read (verified): `ipv6.get_ipv6` returns `{"enable":false,"lan_dns_mode":true,"lan_mode":"nat6"}`. The web UI also reads `lan_dns1`, `lan_dns2` and `lan_ip` (with the default `"fd05:0e25:8607::1/64"` when there is none); the router did not return them while IPv6 was disabled.

Write (view `ipv6`): one `ipv6.set_ipv6`, then the page stays locked and reads `get_ipv6` again after 10 s:

```js
t.enable=this.ipv6.enable,this.ipv6.enable&&(t.lan_mode=this.ipv6.lan_mode,t.lan_dns_mode=this.ipv6.lan_dns_mode,"static"===this.ipv6.lan_mode&&(t.lan_ip=this.ipv6.lan_ip),this.ipv6.lan_dns_mode||(t.lan_dns1=this.ipv6.lan_dns1,this.ipv6.lan_dns2&&(t.lan_dns2=this.ipv6.lan_dns2)))
```

So a disable sends only `{enable: false}`. The page has no WAN setting for IPv6: only "LAN" mode and DNS. The WAN IPv6 address is a static field of `cable.set_config` (see above).

Validation in the UI: `lan_mode` is one of `native` ("Native"), `relay` ("Passthrough"), `nat6` ("NAT6"), `static` ("Static IPv6"). A change to `relay` sets `lan_dns_mode` to true. `lan_ip` (static) is an IPv6 address with a prefix length 48-64 (regex ends with `\/((?:4[8-9]|5[0-9]|6[0-4]))$`). `lan_dns1` is required and `lan_dns2` is optional when `lan_dns_mode` is false; both pass `checkIpv6`. A result with `err_msg` is an error.

The page shows the tip `ipv6.use_tip` with the names of the firewall page, VPN and Cloud as `features`, and `ipv6.note` when the router supports VPN. The texts are loaded at run time with `ui.load_locales`, which this run did not call, so the exact words are not known; the tip says that these features and IPv6 have a limit together.

Unknown: which fields a disable keeps; the read shape while IPv6 is enabled; the errors; whether the write restarts the network or the firewall, and how long it takes.

Proposal for a supervised write test: write `{"enable":false}` (the current value) and read back. Do not enable IPv6 without the owner: the uplink (`wwan`) has no IPv6 (`wwan6` is down), and a WireGuard client runs, which the tip says may not work together with IPv6. If the owner agrees to an enable test, use `{"enable":true,"lan_mode":"nat6","lan_dns_mode":true}` for at most one minute, check that the WireGuard client and the LAN still work, and restore with `{"enable":false}`. Risk: a network or firewall restart drops the management connection for some seconds.

Not tested: the write test of 2026-11-14 stopped before the `ipv6` write. Still not verified: every `ipv6.set_ipv6` call (also `{"enable":false}`), its result, and whether it keeps `lan_mode` and `lan_dns_mode`.

### lan config and guest network (one write tested, the rest from the UI code)

Read (verified): `lan.get_config_list` returns `interfaces[]` with the items `lan` and `guest` (see "lan" above). Flags are 0/1 numbers; `leasetime` is `"12h"`; `transfer_enable` is only in `guest`.

Write (views `lanip` and `guestip`, one component `NetMain` with `netInterface` `lan` or `guest`): the page has two cards, and each sends its own partial `lan.set_config`:

1. Address card: `{interface, ip, netmask, ap_isolate}`.

   ```js
   const t={interface:this.netInterface,ip:this.lan.ip,netmask:this.lan.netmask,ap_isolate:this.lan.ap_isolate};this.setLanConfig(t,"lan"===this.netInterface&&this.oldLan.ip!==t.ip)
   ```

2. DHCP card: `{interface, enable, leasetime, gateway, start, end, dns, lpr?}`, or `{interface, enable: 0}` to disable (after a confirm dialog `lan_ip.dhcp_disable_warning_tips`).

   ```js
   c={interface:this.netInterface,enable:t};t&&(c.leasetime=r+"m",c.gateway=a,c.start=s,c.end=o,c.dns=[],e&&(c.dns=[e]),i&&(c.dns=[...c.dns,i]),n.length&&(c.lpr=n))
   ```

   The UI keeps `lpr` as `[""]` when the read had `[]`, so it sends `"lpr":[""]` in that case. The client sends `lpr` only when it has items.

After the write: when the `lan` IP changed, the page waits 13.5 s, shows "changed to <ip>" and after 1.5 s more loads `http://<new ip>/`. Else it polls `http://<ip>/ping.svg` until the router answers, with a first wait of `api_extra + 10` s, where `api_extra` is an optional number in the result (seconds), and then reads `get_config_list` again.

Errors that the UI maps: `-7` IP conflict with the uplink (`internet.lan_ip_tips` / `internet.guest_ip_tips`), `-8` conflict with the LAN, `-9` conflict with the guest network.

Validation in the UI:
- `ip`: a private IPv4 address that can be a gateway (`checkPrivateIpGateway`); for `guest`, and for `lan` while the guest Wi-Fi is up, the new network must not contain the IP of the other network.
- `netmask`: the select offers `255.255.255.0` and `255.255.0.0` and also accepts typed values; the prefix length is 8-30, at least 16 for `192.*` and at least 12 for `172.*`.
- `ap_isolate`: switch with values 1 and 0; shown only on a router with Wi-Fi.
- DHCP `start` and `end`: private IPv4 addresses inside the current network, `start` < `end`.
- `leasetime`: minutes, 2 to 71582788, sent as `"<n>m"`; the presets are 2, 30, 120, 720, 1440 and 2880 minutes. The UI reads `h`, `m` and `s` units and converts them to minutes.
- `gateway`, `dns1`, `dns2`: optional IPv4 addresses (`isIp`). `dns` has at most two items.
- `lpr`: optional strings without duplicates; the UI does not check their format. The meaning is not known (label `lan_ip.lpr_serve_label`).
- The guest DHCP card is shown only while a guest Wi-Fi is up. `transfer_enable` is not in the UI.

Unknown: whether the router keeps the fields that a partial write does not send (the UI depends on it, so it probably does); the effect and meaning of `lpr` and `transfer_enable`; whether `"12h"` is accepted in a write; the duration of the restart.

Verified (write test of 2026-11-14, see above): `{"interface":"guest","enable":0}` with the guest DHCP server off returns `[]` without `api_extra`, changes no field of either item, and about 9 s later restarts the Wi-Fi: `radio0` changes channel and the repeater uplink drops; `radio1` showed a change in two status samples.

Still not verified:
- a DHCP write with `enable` 1 and the read form of `leasetime` after a write of `"120m"` or `"720m"`;
- whether a DHCP write keeps the address fields and an address write keeps the DHCP fields;
- an address write (also a no-change one) and a change of the guest subnet;
- the handling of `lpr` (`[""]` and no `lpr`) and of `transfer_enable`;
- the errors `-7`, `-8`, `-9`;
- whether every `lan.set_config` restarts the Wi-Fi, or only this one.

Proposal for a supervised write test (guest only; the guest network is off and has no clients): send `{interface:"guest", enable:0}` (the current value) and read back; then `{interface:"guest", enable:1, start:"192.168.9.100", end:"192.168.9.200", leasetime:"120m", gateway:"", dns:[]}`, read back, and restore with the same write with `end:"192.168.9.249"`, `leasetime:"720m"`, and then `{interface:"guest", enable:0}`; check whether the restored `leasetime` reads `"12h"` or `"720m"`. Then `{interface:"guest", ip:"192.168.9.1", netmask:"255.255.255.0", ap_isolate:1}` (no change). Do not write the `lan` item: a change of its IP, netmask or DHCP range cuts the management connection of the Mac (`192.168.8.1` on Ethernet), and a DHCP disable stops new leases. Risks also for `guest`: the write probably restarts the network and the DHCP server, so the LAN may drop for some seconds; a guest IP inside `192.168.8.0/24` gives a conflict.

### wifi config (from the UI code, not tested)

Read (verified): `wifi.get_config` returns `{dfs_support, res[]}` with one item per radio (see "wifi" above). The test router has `radio0` (5G: interfaces `default_radio0`, `guest5g`) and `radio1` (2G: `default_radio1`, `guest2g`). A fresh read on 2026-11-14 had the same shape and values as the fixture (only the four get methods of this section were called), and each typed read in the client keeps all values of the real result (cable 1, ipv6 3, lan 19, wifi 179 leaf values). `key` is the Wi-Fi password in clear text; the fixture has made-up values.

Write (view `wireless`): each radio has a tab per interface. The page sends:

1. The enable switch of an interface sends at once `{iface_name, enabled}`:

   ```js
   changeWifiEnable(t,e,i){this.$emit("apply",{iface_name:e,enabled:i}),this.is5g&&!t&&this.$emit("updateStatus")}
   ```

2. "Apply" sends one `wifi.set_config` for the interface; a main interface also carries the radio settings:

   ```js
   i.iface_name=t.name,i.ssid=t.ssid,i.encryption=t.encryption,i.hidden=t.hidden,"none"!==t.encryption&&(i.key=t.key),!t.guest){...i.device=t,i.hwmode=e,this.repeaterIsUsing||(i.channel=a,i.htmode=n),i.txpower=s,i.random_bssid=r,this.is6g&&(i.chan_6g_only_psc=o)}
   ```

   So `enabled` is not in this write. While the repeater uses the radio (`repeater.status` `device` is the radio), the page disables the channel and width selects and does not send `channel` and `htmode`.

3. The TX power select sends at once `wifi.set_txpower` with `{txpower, device}`.

After each write the page stays locked for 5 s and then reads `get_config` again. After a write to a main 5G interface it polls `wifi.get_status` every 2 s; a radio `state` of `starting` or `cac` shows "DFS starting" and `conflict` shows a DFS conflict. A result with `err_msg` is an error. The UI shows no "Wi-Fi will restart" warning, but the 5 s lock and the status poll suggest that the radio restarts; whether other radios restart too is not known. MLO (`get_mlo_config` / `set_mlo_config`) is used only when `system.get_info` has `software_feature.mlo` true.

Validation in the UI:
- `ssid`: 1-32 bytes (UTF-8).
- `key` (only when `encryption` is not `none`): 8-63 printable ASCII characters (`^[\x20-\x7F]{8,63}$`) or 64 hex digits; at most 63 characters for `sae` and `sae-mixed`.
- `encryption`: one of the `encryptions` of the radio. Labels: `none` OPEN, `psk2` WPA2-PSK, `psk-mixed` WPA/WPA2-PSK, `sae` WPA3-SAE, `sae-mixed` WPA2-PSK/WPA3-SAE.
- `hidden`: `false` (shown) or `true` (hidden).
- `hwmode`: one of `hwmodes`.
- `htmode`: "20", "40", ... up to `htmodes[hwmode]`, and "auto" when `htmodes.auto` is true. A change of `hwmode` that removes the current width selects the largest width.
- `channel`: 0 (Auto) or a channel of `channels` that fits the width (5 GHz blocks of 2/4/8 channels from 36 for 40/80/160 MHz); DFS channels only when `dfs_support`. A change to a DFS channel or to 160 MHz shows a confirm dialog (`wireless.enable_dfs_tips`): the radio waits 60 s (10 min on some channels in Germany) before it sends.
- `txpower`: `Max`, `High`, `Medium`, `Low`.
- The guest tab is shown only in router mode (`system.mode` 0) and without the hidden feature `guestip`. The guest enable switch is disabled when the LAN and guest networks overlap.

Unknown: whether a write without `enabled` keeps the state (the UI depends on it); whether a partial write such as `{iface_name, ssid}` keeps the other fields; whether `channel` and `htmode` are kept when they are not sent; whether `device` is required for the radio fields; whether a write restarts only the radio of the interface or all radios; the errors.

Proposal for a supervised write test (only `radio1` interfaces, which are off): send `{iface_name:"guest2g", enabled:false}` (no change) and read back; then `{iface_name:"guest2g", ssid:<current>, encryption:"psk2", hidden:true, key:<current>}`, read back, and restore with `hidden:false`; then `wifi.set_txpower` `{device:"radio1", txpower:"High"}` and restore `"Max"`. Watch `wifi.get_status` for `radio0` and the uplink (`lan.get_wan_info`, `wwan`) during each step to learn whether a `radio1` write restarts `radio0`. Do not write `radio0` interfaces or radio fields: `radio0` carries the repeater uplink and the main Wi-Fi (`default_radio0`), so a restart drops the uplink and the Wi-Fi clients, and a channel or width change can make the repeater lose its AP. Never enable `guest2g` or `default_radio1` with a test key: that opens a new Wi-Fi network. The read returns all Wi-Fi keys: keep the raw results out of logs and commits.

Not tested: the write test of 2026-11-14 stopped before the `wifi` writes. Still not verified: every `wifi.set_config` and `wifi.set_txpower` call, their results, the partial-write behaviour, and whether a write to a `radio1` interface restarts `radio0`. The `lan.set_config` write of that test restarted the Wi-Fi and dropped the repeater uplink on `radio0`; expect the same risk from a `wifi` write.

## WireGuard and VPN policy

The owner's WireGuard client tunnel runs on the test router (group type 2, status 1). Nobody was at the router, so no write was sent. The reads are verified; all write methods come only from the web UI code.

### Read probe of 2026-11-17

A separate program sent only these read methods. It checked each method name against an allow-list in code before the call: `wg-client.get_status`, `get_group_list`, `get_all_config_list`, `get_config_list` (2 groups with peers), `get_setting` and `get_route_list` (each of the 3 peers); `wg-server.get_config`, `get_status`, `get_peer_list`, `get_setting`, `get_route_list`; `vpn-policy.get_proxy_mode`, `get_global_policy`, `get_domain_policy`, `get_mac_policy`, `get_vlan_policy`. 21 calls, request IDs 30001-30021, all ok. The tunnel had status 1 before and after.

The program redacted the results in memory before it printed or wrote anything, and it stopped without a write if a real string was still in the output. The fixtures replace the `wg-client`, `wg-server` and `vpn-policy` fixtures of 2026-11-12, which had the real VPN names and tunnel addresses (they are still in the git history). Made-up values in these fixtures:
- keys (`private_key`, `public_key`, `preshared_key`): base64 of SHA-256 of a counter; empty values stay empty;
- names: `group-N` for groups of type 2 and 3, `peer-N` for peers (the names of the provider groups, type 1, are presets of the firmware and are kept);
- endpoints: `vpn-N.example.com:51820`, and `port` 51820 in `get_status` and `wg-server.get_config`;
- addresses in `address_v4`, `address_v6`, `allowed_ips`, `dns`, `ipv4`: `10.66.x.y` and `fd00:66::x` with the real prefix lengths (`0.0.0.0/0` and `::/0` are kept), the same real address always gives the same made-up address;
- `listen_port`: 0 and the default 51820 are kept, other ports are made up.

Real: IDs, flags, numbers (MTU, keepalive, byte counts) and the field structure.

Findings of the reads:
- `group_id` and `peer_id` are numbers that the router gives (for example 4 digits). They were the same as in the probe of 5 days before. They are not positions in a list.
- Groups: `group_type` 1 provider (7 presets with `username`, `procedure` and mostly `password`), 2 custom (only `auth_type`, `group_id`, `group_name`, `group_type`, `peer_count`), 3 "from app".
- `get_config_list` and `get_all_config_list` return the same peers with a different set of empty fields: `get_config_list` has `address_v6` `""` and `preshared_key` `""`; `get_all_config_list` leaves them out, but has `listen_port` 0 and `mtu` 0 where `get_config_list` leaves them out. `ipv6_enable` is only in the peers of the custom group.
- `get_setting` of a peer has `default_metric`, `local_access`, `masq`, and `mtu` for 2 of the 3 peers.
- `get_status` while connected: `status` 1, `group_id`, `peer_id`, `name`, `domain`, `port`, `ipv4`, `proxy`, `log`, `rx_bytes`, `tx_bytes`. A stopped status was not read.
- The WireGuard server is not initialized (`initialization` false, empty keys) and has no peers. The route lists, the domain list and the MAC list are empty. So the items of `wg-server.get_peer_list`, `wg-server.get_status.peers`, the route rules, `domain_list` and `url` are not verified.

### Methods from the UI code

Source: the views `wgclient`, `wgserver` and `vpndashboard` (names from `ui.get_menu_list`), `/views/gl-sdk4-ui-<view>.common.js.gz`, and the error table `coreErrTips` of the main bundle.

`wg-client` (client: `WireGuardClientService`):

| Method | Params | UI use |
|---|---|---|
| `get_status` | `{}` | status poll (verified) |
| `get_group_list` | `{}` | group list (verified) |
| `get_all_config_list` | `{}` | dashboard list (verified) |
| `get_config_list` | `{group_id}` | peers of a group (verified) |
| `get_setting` | `{group_id, peer_id}` | options of the peer of the tunnel (verified) |
| `get_route_list` | `{group_id, peer_id}` | route rules (verified) |
| `add_group` | `{group_name}` | the UI sends a default name, reads the list, takes the last group and opens the rename field; the result is not used |
| `set_group` | `{group_id, group_name}` | rename; only type 2 |
| `remove_group` | `{group_id}` | only type 2, not the group of the tunnel |
| `add_config` | peer fields + `group_id` | the result is not used; the UI reads the list again |
| `set_config` | peer fields + `group_id`, `peer_id` | not for the peer of the tunnel (read only in the UI) and not for provider groups |
| `remove_config` | `{group_id, peer_id}` | not for the peer of the tunnel |
| `clear_config_list` | `{group_id, delete_key?}` | `delete_key` only for provider groups; not for the group of the tunnel |
| `start` | `{group_id, peer_id}` | then status 2 (connecting), poll after 5 s |
| `stop` | `{}` | |
| `set_setting` | `{group_id, peer_id, default_metric, local_access, masq, mtu}` | only for the peer of the tunnel; the UI sends the read fields back, `mtu` 0 when empty (68-65535) |
| `add_route`, `set_route` | `{group_id, peer_id, rule_id?, dest, gateway?, metric?, mtu?, scope?, route_flag}` | `route_flag` 4 or 6; only non-empty fields |
| `remove_route` | `{group_id, peer_id, rule_id, route_flag}` | |
| `check_config`, `confirm_config` | `{group_id, filename}`, `{group_id}` | import of uploaded files (`/tmp/wg_upload`); not in the client |
| `clear_user_pass`, `get_third_config`, `get_recommend_config` | `{group_id, ...}` | provider accounts; not in the client |

Peer fields of `add_config` and `set_config`: `name`, `address_v4`, `address_v6`, `ipv6_enable`, `private_key`, `public_key`, `allowed_ips` (one string, `", "` between items), `end_point`, `presharedkey_enable`, `preshared_key` (only when enabled), and `dns`, `listen_port`, `persistent_keepalive`, `mtu` only when set. The edit dialog sends the fields of the read peer back with the changes. Validation in the UI: `name` 1-64 characters, not only spaces, no leading `.`, no `~`; `address_v4` required, IPv4 or IPv4 CIDR; `address_v6` IPv6 or CIDR when `ipv6_enable`; `private_key`, `public_key` required and `preshared_key` 44 characters ending in `=`; `end_point` required, `host:port` with port 0-65535 (the UI checks the host with its `isDomain` test or as IPv6); `allowed_ips` IPv4 or IPv6 CIDRs; `dns` IPv4 or IPv6 addresses, comma separated; `listen_port` and `persistent_keepalive` 1-65535; `mtu` 576-65535. For a custom group the UI calls `wg_client.gen_key` `{private_key}` (module with an underscore) to show the public key of the router; with an empty key this makes a new key pair. The client does not call it.

`wg-server` (client: `WireGuardServerService`):

| Method | Params | UI use |
|---|---|---|
| `get_config`, `get_status`, `get_peer_list`, `get_setting`, `get_route_list` | `{}` | verified |
| `generate_key` | `{}` | "initialize" button, only while `initialization` is false; makes the server keys |
| `set_config` | `{address_v4, port, private_key?, ipv6_enable?, address_v6?}` | only while the server is stopped; IPv6 fields only while IPv6 of the router is enabled; error 20001024 is an invalid private key |
| `generate_publickey` | `{private_key}` | shows the public key of a typed private key; not in the client |
| `add_peer` | `{name, presharedkey_enable, presharedkey, allowed_ips?, dns?, persistent_keepalive, mtu}` | returns `{peer_id}`; then the UI calls `generate_peer`. The UI sends `""` for an empty `persistent_keepalive` or `mtu`; the client leaves them out |
| `set_peer` | the fields of `add_peer` + `peer_id` (and the other fields of the read peer) | then `generate_peer` |
| `remove_peer` | `{peer_id}` | |
| `generate_peer` | `{peer_id}` | the peer configuration file and QR code: `address`, `allowed_ips`, `dns`, `domain_end_point`, `ip_end_point`, `listen_port`, `persistent_keepalive`, `private_key`, `public_key`, `mtu`, `presharedkey` |
| `start`, `stop` | `{}` | |
| `set_setting` | `{local_access, masq, mtu, client_to_client}` | |
| `add_route`, `set_route`, `remove_route` | as `wg-client` without `group_id`, `peer_id` | |

Peer list fields that the UI reads: `peer_id`, `name`, `client_ip`, `deprecated` (1: struck through, can only be removed), `allowed_ips` (at most 5 in the UI), `dns`, `presharedkey_enable`, `presharedkey` (`"Auto Generate"` placeholder), `persistent_keepalive`, `mtu`.

`vpn-policy` (client: `VPNPolicyService`):

| Method | Params | UI use |
|---|---|---|
| `get_proxy_mode` / `set_proxy_mode` | `{mode}` | 0 global, 1 auto detect, 2 custom route, 3 domain/IP, 4 MAC, 5 VLAN; timeout 20 s; modes 0 and 3-5 not offered while the WireGuard and OpenVPN clients both run |
| `get_global_policy` / `set_global_policy` | client dialog `{kill_switch, wan_access, service_policy}`, server dialog `{vpn_server_policy}` | timeout 50 s; `kill_switch` 1 sets `wan_access` 0 |
| `get_domain_policy` / `set_domain_policy` | `{default_policy, manual, domain_list}` (one string, a line per item) or `{default_policy, manual: false, url}` | timeout 30 s; `default_policy` 0 "use VPN" for the list, 1 "do not use VPN"; with 0 the list must not be empty |
| `check_domain_online` | `{url}` | checks a list URL; the router fetches it; not in the client |
| `get_mac_policy` / `set_mac_policy` | `{default_policy, mac_list}` | full list; with 0 the list must not be empty; MACs compared without case |
| `get_vlan_policy` / `set_vlan_policy` | `{vlans: [{id, vpn}]}` | full list; id 1 "Private", 2 "Guest" |

The read of `domain_list` can be a list or a string in the UI code.

Errors: a result with `err_code` is an error (an `*RPCError` in the client). The UI maps these VPN codes to texts: 20001013 parameter has an injection risk, 20001014/20001015 configuration file or configuration not read, 20001017 target interface already runs, 20001020/20001021 key file missing, 20001025/20001026 local IPv4/IPv6 network not read, 20001027/20001028 conflict with a local IPv4/IPv6 address, 20001029 client interface conflict, 20001033 port in use, 20001100-20001107 provider login and download errors, 20001201 upload not unzipped, 20001203-20001215 provider account errors (20001205: the provider key must be revoked, 20001206 invalid private WireGuard key). `wg-server.set_config` maps 20001024 to an invalid private key.

### Data for a Terraform resource

- WireGuard client group: identity `group_id` (from the group list after `add_group`); managed: `group_name` of a type 2 group.
- WireGuard client peer: identity `group_id` + `peer_id` (from the peer list after `add_config`); managed: the peer fields above. `private_key` and `preshared_key` are secrets that the router returns in clear text in each read.
- Tunnel: which peer runs (`get_status`), with `start` / `stop`; peer setting and route rules (`rule_id`) per peer.
- WireGuard server: a singleton configuration (one-way initialization), peers with identity `peer_id` (returned by `add_peer`), a setting and route rules.
- VPN policy: singletons (proxy mode, global policy, domain list, MAC list, VLAN list); the lists are written in full.

### Open questions

- The results of `add_group` and `add_config` (does the router return the new ID?) and of all other writes.
- Whether `set_config` keeps fields that it does not get, and which fields the router fills in (for example `listen_port`, `mtu`, `ipv6_enable`).
- Whether a write to a peer, its setting or its routes restarts the running tunnel, also for a peer that does not run; how long `start` takes; the status after `stop`.
- Whether the router refuses a write to the running peer or the removal of its group (the UI only hides the buttons).
- Whether `set_proxy_mode`, `set_global_policy` or a policy list write reloads the firewall or the tunnel, and for how long the traffic drops. The UI timeouts (20-50 s) suggest long writes.
- The meanings of `proxy`, `default_metric`, `procedure`, `auth_type`, `wan_access` and `service_policy` (the label texts are loaded at run time).
- The item shapes that the probe could not read (see above).
- Firmware 4.8: whether `wg-client` is still there. On 4.9.0 `vpn-client` replaces it.

### Proposal for a supervised write test

Only with the owner at the router. Before and during each step, read every 2 s: `wg-client.get_status` (`status` 1 and growing `rx_bytes`), `system.get_status` and `repeater.get_status`. Stop at the first change of the tunnel or the uplink. Do not write the custom group of the tunnel, its peers, `start` or `stop`.

1. `wg-client.add_group` `{"group_name":"tf-test"}`, then `get_group_list`. Learn the result and the new `group_id`. Risk: low (a new, unused group). Restore: `remove_group`.
2. `set_group` `{"group_id":<new>,"group_name":"tf-test-2"}`. Risk: low.
3. `add_config` in the new group with made-up keys, `end_point` `192.0.2.1:51820` (a documentation address), `address_v4` `10.255.0.2/32`, `allowed_ips` `10.255.0.0/24`. Then `get_config_list`. Learn the result, the new `peer_id` and the fields that the router adds. Risk: low while the peer does not run. Restore: `remove_config`.
4. `set_config` of that peer without `dns` and `mtu`: learn whether the router keeps them. Risk: low; watch the tunnel for a restart.
5. `get_setting`, then `set_setting` with the same values, and `add_route` / `set_route` / `remove_route` on that peer. Risk: a restart of the running tunnel if the router reloads all WireGuard settings. Restore: write the read values back.
6. `remove_config`, `clear_config_list` (empty group), `remove_group`; check that the list equals the list before the test.
7. Optional, with the owner's consent: `vpn-policy.set_vlan_policy` and `set_mac_policy` with the values just read (no change). Risk: a firewall reload can drop the tunnel traffic for some seconds. Do not test `set_proxy_mode`, `set_global_policy` or `set_domain_policy`: `kill_switch` and the policy modes can cut the traffic of all clients.
8. Optional: `wg-server.generate_key` cannot be undone (the server stays initialized). Then `add_peer`, `generate_peer`, `set_peer`, `remove_peer` with the server stopped. Do not `start` the server: it opens a UDP port on the WAN.

## Methods not called

Not read-only or not sure: all `set_*`, `add_*`, `remove_*`, etc. Also these `get_*` methods: `upgrade.check_firmware_online`, `tailscale.get_auth_url`, `wg-client.get_third_config`, `ovpn-client.get_third_config` (send credentials to a third-party VPN service), `wg-client.get_recommend_config`, `ovpn-client.get_recommend_config` (query an external service).

## Redaction

Replaced with stable fake values (same real value → same fake value, type kept): password, cipher password, sid, salt, nonce, login hash, Wi-Fi keys, private keys, preshared keys, tokens, `sn`, `sn_bak`, `ddns` / `device_id`, MAC addresses (`02:00:00:00:xx:xx`), public IPv4 (`203.0.113.x`), global/ULA/link-local IPv6 (`2001:db8::x`, `fd00::x`, `fe80::x`), WireGuard endpoint hostnames (`host-N.example.com`), SSIDs (`ssid-N`, also the SSIDs of other networks in `repeater.*`), hostnames and client names (`hostname`, `alias`, and `name` in objects that have a `mac`: `hostname-N`), WireGuard public keys (base64 of 32 bytes from SHA-256 of a counter; empty values stay empty), and:
- private and shared (`100.64.0.0/10`) IPv4 addresses outside the default LAN (`192.168.8.0/24`) and guest (`192.168.9.0/24`) networks: each real /24 network gets one made-up network `10.x.y`, and the host part and the prefix length stay;
- names of VPN groups (`group_name`: `group-N`), except the built-in provider groups `AzireVPN`, `FromApp`, `Hideme`, `IPVanish`, `Mullvad`, `NordVPN`, `PIA` and `Surfshark`;
- names of VPN peers and configurations (`name` in objects that have a `peer_id` or a `client_id`: `peer-N`), except the service names `wgclient`, `wgserver`, `ovpnclient` and `ovpnserver`;
- names of firewall rules and port forwards (`name` in objects that have a `src` and a `dest`: `rule-N`);
- VPN usernames (`username` other than `root`: `user-N`);
- `listen_port` other than 0 and the default 51820: 40000 + N;
- the time zone: `zonename` `UTC`, `timezone` `UTC0`, `tzoffset` `+0000`, and `localtime` gets the value of `timestamp`.

A fake value replaces the real value also inside other strings, such as log lines, when the real value has 4 or more characters. The number N depends on the order of the probe run, so it can change between runs.

Kept: loopback, link-local, multicast and broadcast IPv4 addresses, the addresses of the default LAN and guest networks, well-known public DNS resolvers, the built-in VPN provider group names, all other `name` fields (firmware lists, for example DNS servers, apps, Wi-Fi interfaces and services), settings, flags, IDs, counters, and the times in log lines.

The fixtures in this directory come from an earlier version of `cmd/probe`. After the probe, the values from the owner's network and settings were replaced by hand with made-up values: the uplink network and its DNS server (`192.168.100.x`), the connection UUID of the uplink, a link-local client address, a WireGuard `listen_port`, the time zone (`UTC`) and the reboot schedule. The `wg-client`, `wg-server` and `vpn-policy` fixtures have the made-up values of "WireGuard and VPN policy".

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
