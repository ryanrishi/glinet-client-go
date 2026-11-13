package main

type results = map[string]map[string]any

type callSpec struct {
	module string
	method string
	params func(results) (map[string]any, bool)
}

func (c callSpec) name() string { return c.module + "." + c.method }

func none(results) (map[string]any, bool) { return map[string]any{}, true }

func fixed(p map[string]any) func(results) (map[string]any, bool) {
	return func(results) (map[string]any, bool) { return p, true }
}

func plain(module string, methods ...string) []callSpec {
	specs := make([]callSpec, len(methods))
	for i, m := range methods {
		specs[i] = callSpec{module, m, none}
	}
	return specs
}

func dep(module, method string, params func(results) (map[string]any, bool)) []callSpec {
	return []callSpec{{module, method, params}}
}

func resultOf(res results, name string) map[string]any {
	m, _ := res[name]["result"].(map[string]any)
	return m
}

func firstItem(res results, name, list string) (any, bool) {
	arr, _ := resultOf(res, name)[list].([]any)
	if len(arr) == 0 {
		return nil, false
	}
	return arr[0], true
}

func firstField(res results, name, list, field string) (any, bool) {
	item, ok := firstItem(res, name, list)
	if !ok {
		return nil, false
	}
	obj, ok := item.(map[string]any)
	if !ok {
		return nil, false
	}
	v, ok := obj[field]
	return v, ok
}

func aclGroup(res results) (map[string]any, bool) {
	item, ok := firstItem(res, "acl.get_group_list", "groups")
	if !ok {
		return nil, false
	}
	if obj, isObj := item.(map[string]any); isObj {
		item, ok = obj["group"]
		if !ok {
			item, ok = obj["name"]
		}
	}
	return map[string]any{"group": item}, ok
}

func nonEmptyGroupID(res results, module, countField string) (any, bool) {
	groups, _ := resultOf(res, module+".get_group_list")["groups"].([]any)
	for _, g := range groups {
		obj, _ := g.(map[string]any)
		if n, _ := obj[countField].(float64); n > 0 {
			return obj["group_id"], true
		}
	}
	return firstField(res, module+".get_group_list", "groups", "group_id")
}

func groupID(module, countField string) func(results) (map[string]any, bool) {
	return func(res results) (map[string]any, bool) {
		id, ok := nonEmptyGroupID(res, module, countField)
		return map[string]any{"group_id": id}, ok
	}
}

func groupAndItem(module, countField, list, idField string) func(results) (map[string]any, bool) {
	return func(res results) (map[string]any, bool) {
		gid, ok := nonEmptyGroupID(res, module, countField)
		if !ok {
			return nil, false
		}
		id, ok := firstField(res, module+".get_config_list", list, idField)
		return map[string]any{"group_id": gid, idField: id}, ok
	}
}

func parentalGroup(res results) (map[string]any, bool) {
	id, ok := firstField(res, "parental-control.get_config", "groups", "id")
	return map[string]any{"group_id": id}, ok
}

func pluginName(res results) (map[string]any, bool) {
	name, ok := firstField(res, "plugins.get_list", "packages", "name")
	return map[string]any{"name": name}, ok
}

func modemBus(res results) (map[string]any, bool) {
	bus, ok := firstField(res, "modem.get_info", "modems", "bus")
	return map[string]any{"bus": bus}, ok
}

var calls = concat(
	plain("system", "get_info", "get_status", "disk_info", "get_timezone_config", "get_unixtime",
		"get_httpd_mem_status", "get_security_policy", "get_load", "get_network_status", "get_timezone_list"),
	plain("acl", "get_group_list"),
	dep("acl", "get_acl_list", aclGroup),
	plain("adguardhome", "get_config"),
	plain("bark", "get_status", "get_config"),
	plain("black_white_list", "get_config"),
	plain("cable", "get_status", "get_config"),
	plain("clients", "get_list", "get_status"),
	plain("cloud", "get_config", "get_batch_config"),
	plain("cloud-batch-manage", "get_batch_config", "get_2b_config"),
	plain("ddns", "get_config", "get_status"),
	plain("dlna", "get_config"),
	plain("dns", "get_host", "get_config", "get_info"),
	plain("edgerouter", "get_config", "get_status"),
	plain("fan", "get_status", "get_config"),
	plain("firewall", "get_dmz", "get_rule_list", "get_port_forward_list", "get_wan_access", "get_zone_list",
		"get_acl_rule_list", "get_acl_zone_list"),
	plain("igmp", "get_config"),
	plain("ipv6", "get_ipv6"),
	plain("kmwan", "get_config", "get_status"),
	plain("lan", "get_static_bind_list", "get_config_list", "get_wan_info"),
	plain("led", "get_config"),
	plain("local-access", "get_config"),
	dep("logread", "get_system_log", fixed(map[string]any{"lines": 20})),
	plain("logread", "get_kernel_log"),
	dep("logread", "get_crash_log", fixed(map[string]any{"mode": 0})),
	plain("logread", "get_config"),
	dep("logread", "get_nginx_log", fixed(map[string]any{"type": "error"})),
	plain("macclone", "get_mac"),
	plain("mcu", "get_battery_config", "get_oled_config"),
	plain("modem", "get_info", "get_status", "get_sms_list", "get_sim_info", "get_profile_list",
		"get_sim_signal", "get_name", "get_signals", "get_cell_tower"),
	dep("modem", "get_config", modemBus),
	dep("modem", "get_cells_info", modemBus),
	dep("modem", "get_traffic_config", modemBus),
	dep("modem", "get_debug_msg", modemBus),
	dep("modem", "get_slot_config", modemBus),
	plain("mwan3", "get_config", "get_status"),
	plain("nas-web", "get_proto_config", "get_user_list", "get_disk_list", "get_file_list", "get_share_list",
		"get_nas_ser", "get_status"),
	plain("netmode", "get_mode"),
	plain("network", "routes", "routes6", "get_arp_list", "get_dhcp_leases", "get_hwnat_config", "get_netnat_config"),
	dep("network", "check_wan_cable", fixed(map[string]any{"secondwan": false})),
	plain("otbr", "get_status", "get_joiner_list", "get_network_data", "get_neighbor_list", "get_bbr_status",
		"get_srp_server_service", "get_srp_server_config"),
	plain("ovpn-client", "get_status", "get_group_list", "get_all_config_list"),
	dep("ovpn-client", "get_config_list", groupID("ovpn-client", "client_count")),
	dep("ovpn-client", "get_route_list", groupAndItem("ovpn-client", "client_count", "clients", "client_id")),
	dep("ovpn-client", "get_setting", groupAndItem("ovpn-client", "client_count", "clients", "client_id")),
	plain("ovpn-server", "get_config", "get_status", "get_user_list", "get_route_list", "get_setting"),
	plain("parental-control", "get_app_list", "get_config", "get_status"),
	dep("parental-control", "get_brief", parentalGroup),
	plain("plugins", "get_repository_status", "get_config"),
	dep("plugins", "get_list", fixed(map[string]any{"status": 1})),
	dep("plugins", "get_package_info", pluginName),
	plain("qos", "get_device_group", "get_config", "get_client_list", "get_channel_bandwidth_ratio", "get_bandwidth_config"),
	plain("reboot", "get_config"),
	plain("repeater", "get_config", "get_status", "get_saved_ap_list"),
	plain("rs485", "get_forward_log", "get_config", "get_forward_config", "get_connect_status", "get_tcp_clients"),
	plain("rtty", "get_config"),
	plain("s2s", "get_status"),
	plain("samba", "get_config"),
	plain("sms-forward", "get_config"),
	plain("switch-button", "get_funcs", "get_config"),
	plain("tailscale", "get_status", "get_config", "get_exit_node_list"),
	plain("tethering", "get_status"),
	plain("timer", "get_led", "get_reboot", "get_wifi", "get_disk"),
	plain("tor", "get_config", "get_status"),
	plain("ui", "get_lang", "get_menu_list"),
	plain("upgrade", "get_config", "get_online_upgrade_status"),
	plain("vpn-client", "get_all_config_list", "get_status", "get_tunnel", "get_connection_methods"),
	plain("vpn-policy", "get_proxy_mode", "get_global_policy", "get_domain_policy", "get_mac_policy", "get_vlan_policy"),
	plain("wg-client", "get_status", "get_group_list", "get_all_config_list"),
	dep("wg-client", "get_config_list", groupID("wg-client", "peer_count")),
	dep("wg-client", "get_route_list", groupAndItem("wg-client", "peer_count", "peers", "peer_id")),
	dep("wg-client", "get_setting", groupAndItem("wg-client", "peer_count", "peers", "peer_id")),
	plain("wg-server", "get_peer_list", "get_config", "get_status", "get_route_list", "get_setting"),
	plain("wifi", "get_config", "get_status"),
	plain("zerotier", "get_config", "get_status"),
)

func concat(lists ...[]callSpec) []callSpec {
	var out []callSpec
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}
