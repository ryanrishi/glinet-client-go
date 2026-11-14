package glinettest

import (
	"encoding/json"
	"slices"
	"sync"
)

// readResult is the result of a read-only method. A test can replace it.
type readResult struct {
	mu  sync.Mutex
	raw json.RawMessage
}

func (r *readResult) get() json.RawMessage {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.raw)
}

func (r *readResult) set(raw json.RawMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.raw = slices.Clone(raw)
}

func (r *readResult) handle(json.RawMessage) (any, error) {
	return r.get(), nil
}

// The default results of the read-only methods. They are the results of the real router in
// testdata/fixtures/axt1800_4.7.0, with the secrets and names replaced by made-up values.
const (
	systemInfo = `{
		"board_info": {
			"architecture": "ARMv8 Processor rev 4",
			"hostname": "hostname-1",
			"kernel_version": "5.4.164",
			"model": "GL Technologies, Inc. AXT1800",
			"openwrt_version": "OpenWrt 23.05-SNAPSHOT r23485+111-e92cf0c46f"
		},
		"country_code": "US",
		"cpu_num": 4,
		"ddns": "redacted-ddns-1",
		"device_type": 1,
		"firmware_date": "2025-03-19 10:28:59",
		"firmware_type": "release8",
		"firmware_version": "4.7.0",
		"hardware_feature": {
			"bluetooth": false,
			"build_in_modem": "",
			"fan": true,
			"gps": false,
			"hwnat": true,
			"lan": "eth1 eth2",
			"mcu": false,
			"microsd": "",
			"modem_reset": 2,
			"nand": false,
			"noled": false,
			"nowds": false,
			"radio": "radio0 radio1",
			"reset_button": "gpio-50",
			"rs485": false,
			"slot": "single",
			"submodel": "",
			"switch_button": "gpio9",
			"usb": "1-1,2-1",
			"usb3": "2-1",
			"usb_reset": "usb_power,1",
			"wan": "eth0"
		},
		"hardware_version": "",
		"hidden_features": [],
		"mac": "02:00:00:00:00:01",
		"model": "axt1800",
		"sn": "redacted-sn-1",
		"sn_bak": "redacted-sn_bak-1",
		"software_feature": {
			"adguard": true,
			"bark": false,
			"cellular_upgrade": false,
			"ids_ips": false,
			"ipv6": true,
			"ksmbd": false,
			"mlo": false,
			"nas": true,
			"passthrough": false,
			"repeater_eap": true,
			"secondwan": false,
			"sms_forward": true,
			"tor": true,
			"vpn": true
		},
		"vendor": "GL.iNet"
	}`

	systemStatus = `{
		"client": [
			{
				"cable_total": 1,
				"wireless_total": 0
			}
		],
		"network": [
			{
				"interface": "wwan6",
				"online": false,
				"up": false
			},
			{
				"interface": "tethering6",
				"online": false,
				"up": false
			},
			{
				"interface": "wwan",
				"online": true,
				"up": true
			},
			{
				"interface": "tethering",
				"online": false,
				"up": false
			},
			{
				"interface": "wan",
				"online": false,
				"up": false
			},
			{
				"interface": "wan6",
				"online": false,
				"up": false
			}
		],
		"service": [
			{
				"group_id": 2082,
				"name": "wgclient",
				"peer_id": 2004,
				"status": 1
			},
			{
				"name": "wgserver",
				"status": 0
			},
			{
				"name": "ovpnclient",
				"status": 0
			},
			{
				"name": "ovpnserver",
				"status": 0
			},
			{
				"name": "adguard",
				"status": 0
			},
			{
				"name": "tor",
				"status": 0
			},
			{
				"name": "parental_control",
				"status": 0
			},
			{
				"name": "tailscale",
				"status": 0
			},
			{
				"name": "zerotier",
				"status": 0
			},
			{
				"name": "mptun",
				"status": 0
			}
		],
		"system": {
			"cpu": {
				"temperature": 69
			},
			"ddns_enabled": false,
			"flash_app": 1929216,
			"flash_free": 38596608,
			"flash_total": 134217728,
			"guest_ip": "192.168.9.1",
			"guest_netmask": "255.255.255.0",
			"ipv6_enabled": false,
			"lan_ip": "192.168.8.1",
			"lan_netmask": "255.255.255.0",
			"load_average": [
				0,
				0.05,
				0.07
			],
			"memory_buff_cache": 104353792,
			"memory_free": 155758592,
			"memory_total": 424235008,
			"mode": 0,
			"netnat_enabled": true,
			"timestamp": 1794517485,
			"tzoffset": "-0500",
			"uptime": 835.31
		},
		"wifi": [
			{
				"band": "5G",
				"channel": 153,
				"encryption": "psk2",
				"guest": false,
				"hidden": false,
				"mld": false,
				"name": "default_radio0",
				"passwd": "redacted-passwd-1",
				"ssid": "ssid-1",
				"up": true
			},
			{
				"band": "2G",
				"channel": 6,
				"encryption": "psk2",
				"guest": false,
				"hidden": false,
				"mld": false,
				"name": "default_radio1",
				"passwd": "redacted-passwd-2",
				"ssid": "ssid-1",
				"up": false
			},
			{
				"band": "5G",
				"channel": 153,
				"encryption": "psk2",
				"guest": true,
				"hidden": false,
				"mld": false,
				"name": "guest5g",
				"passwd": "redacted-passwd-2",
				"ssid": "ssid-2",
				"up": false
			},
			{
				"band": "2G",
				"channel": 6,
				"encryption": "psk2",
				"guest": true,
				"hidden": false,
				"mld": false,
				"name": "guest2g",
				"passwd": "redacted-passwd-2",
				"ssid": "ssid-3",
				"up": false
			}
		]
	}`

	systemDiskInfo = `{
		"root": {
			"free": 38596608,
			"total": 42745856,
			"used": 1929216
		},
		"tmp": {
			"free": 210915328,
			"total": 212115456,
			"used": 1200128
		}
	}`

	// clientsList has 3 of the 51 clients of the test router.
	clientsList = `{
		"clients": [
			{
				"blocked": false,
				"iface": "cable",
				"ip": "192.168.8.193",
				"ipv6": [],
				"last_rx": [84772603, 84779577, 84789148, 84796893, 84802552, 84806682, 84808647, 84837314, 84842948, 86311695, 114848575, 147629463, 159134558, 163594078, 163616946, 163618344, 163623506, 163628305, 163636779, 163645079, 163652721, 165547641, 172853498, 172863127, 172951556, 172955667, 172962378, 172970007, 172991193, 173011674, 173014828, 173145888, 173246857, 173870234, 173892953, 173905613, 173918232, 173922885, 173931813, 173940134, 173941934, 173945043, 173966466, 173969956, 174003175, 174073287, 174164123, 174170421, 174173872, 174202221, 174210352, 174218206, 174232415, 174239461, 174252120, 174253077, 174259227, 174266924, 174269688, 174272028],
				"last_tx": [5582204, 5584867, 5596119, 5621512, 5678351, 5703063, 5709641, 5776042, 5782902, 5871234, 6265926, 6669330, 6813866, 6940208, 6947022, 6949566, 6953391, 6979719, 6985605, 7005709, 7048481, 7078211, 7192777, 7207484, 7242698, 7245090, 7280479, 7315835, 7361196, 7404352, 7458577, 7485706, 7494056, 7608800, 7676431, 7719164, 7727643, 7756437, 7760723, 7872569, 7874018, 7913669, 7931981, 7936523, 7950852, 7976143, 7981241, 8007976, 8009358, 8032562, 8038490, 8064665, 8116879, 8120699, 8156254, 8156975, 8158903, 8206460, 8207618, 8236413],
				"last_update_rate": 1794515224,
				"limit_rx": 0,
				"limit_tx": 0,
				"mac": "02:00:00:00:00:16",
				"name": "hostname-3",
				"online": true,
				"online_time": 1794517483,
				"remote": true,
				"rx": 497219,
				"total_rx": 174272028,
				"total_rx_init": 0,
				"total_tx": 8236413,
				"total_tx_init": 96880,
				"tx": 14745,
				"type": 2
			},
			{
				"alias": "hostname-11",
				"blocked": false,
				"class": "laptop",
				"iface": "5G",
				"ip": "192.168.8.174",
				"ipv6": [],
				"last_rx": [1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722, 1020070722],
				"last_tx": [212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725, 212475725],
				"last_update_rate": 1794515224,
				"limit_rx": 0,
				"limit_tx": 0,
				"mac": "02:00:00:00:00:27",
				"name": "hostname-12",
				"online": false,
				"online_time": 1713557437,
				"remote": false,
				"rx": 0,
				"total_rx": 1020070722,
				"total_rx_init": 1020070722,
				"total_tx": 212475725,
				"total_tx_init": 212475725,
				"tx": 0,
				"type": 1
			},
			{
				"alias": "hostname-7",
				"blocked": false,
				"iface": "5G",
				"ip": "192.168.8.236",
				"ipv6": [],
				"last_rx": [354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875, 354843683875],
				"last_tx": [70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349, 70437565349],
				"last_update_rate": 1794515224,
				"limit_rx": 0,
				"limit_tx": 0,
				"mac": "02:00:00:00:00:0E",
				"name": "hostname-6",
				"online": false,
				"online_time": 1713557437,
				"remote": false,
				"rx": 0,
				"total_rx": 354843683875,
				"total_rx_init": 354843683875,
				"total_tx": 70437565349,
				"total_tx_init": 70437565349,
				"tx": 0,
				"type": 1
			}
		]
	}`
)
