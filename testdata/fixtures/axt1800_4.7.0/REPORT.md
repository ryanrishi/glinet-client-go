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

Verified on the router only with `enable: false`, so that no reboot could be scheduled. The schedule was `{"enable":false,"hour":"03","min":"00","week":[1,4]}` before and after the test. The fixtures `timer.*_reboot*.json` with request IDs 12000-12010 are one sequence (`glinettest/timer_test.go` replays them).

| Call | Params | Result |
|---|---|---|
| `get_reboot` | none | `{"enable":bool, "hour":string, "min":string, "week":[int]}` |
| `set_reboot` | `{enable, hour?, min?, week?}` | `[]` |

Formats: as in "timer LED schedule". `hour` and `min` are strings with two digits (`"03"`, `"00"`) in the router's local time. The web UI labels the days 0 Sunday to 6 Saturday, and sends `week` sorted.

Behavior:
- A write with `enable: false`, without `enable`, or `{}` sets `enable` to `false` and ignores `hour`, `min` and `week` (`{"enable":false,"hour":"05","min":"30","week":[2]}` changed nothing else).
- `enable: null` and `week` as a string (`"1,2"`) return `-32602 Invalid params` (no `err`) and change nothing.
- Each write returned in about 4 ms.

Not verified (no write with `enable: true` was sent): the enabled write and its errors, how `hour`, `min` and `week` are stored, and the day numbering of the router. A write with `enable: true` can only set the values; after it the router keeps them, so a test of it cannot restore the initial values `03:00` on days `[1,4]` without a schedule that can run in less than 6 days. The web UI sends `{enable: true, hour, min, week}` and only `{enable: false}` to disable, as for the LED schedule, so the enabled write probably works like `timer.set_led` (all fields necessary, else `err_code -1`). `glinettest` emulates it like that.

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

Verified on the router in mode `auto`, with a WireGuard client tunnel that supplies DNS (`wg 192.168.5.1` in `server`). The config was `{"mode":"auto","force_dns":false,"rebind_protection":false,"override_vpn":true,"dot_provider":"0","nextdns_id":"","controld_id":"","free_controld_id":"","server":["wwan 192.168.4.253","wg 192.168.5.1"]}` before and after the test, and `server` did not change during it. Only `rebind_protection` and `force_dns` were changed on purpose, one at a time. The fixtures `dns.*_config*.json` with request IDs 14000-14315 are one sequence (`glinettest/dns_test.go` replays them).

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
- `get_config`: no `proto`. Extra: `controld_id`, `dot_provider` (string, `"0"`), `free_controld_id`, `override_vpn`. `server` items are strings like `"wwan 192.168.4.253"`.
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
