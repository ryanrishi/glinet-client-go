package glinettest

import (
	"encoding/binary"
	"encoding/json"
	"sync"
	"time"
)

// TimezoneConfig is the time zone configuration of the system module.
type TimezoneConfig struct {
	Zonename            string `json:"zonename"`
	Timezone            string `json:"timezone"`
	TZOffset            string `json:"tzoffset"`
	AutoTimezoneEnabled bool   `json:"autotimezone_enabled"`
	Localtime           int64  `json:"localtime"`
	Timestamp           int64  `json:"timestamp"`
}

// System serves the system module. Timestamp is the router clock: set_timezone_config sets it
// from localtime and ignores autotimezone_enabled, like the router. It also sets TZOffset to the
// UTC offset of the POSIX TZ string timezone at that clock, with daylight saving time.
// A timezone that is not a valid POSIX TZ string gives "+0000".
type System struct {
	mu       sync.Mutex
	timezone TimezoneConfig
}

const systemStatus = `{
	"client": [{"cable_total": 1, "wireless_total": 2}],
	"network": [
		{"interface": "wan", "online": true, "up": true},
		{"interface": "wwan", "online": false, "up": false}
	],
	"service": [
		{"name": "wgclient", "status": 1, "group_id": 1001, "peer_id": 2001},
		{"name": "adguard", "status": 0}
	],
	"system": {
		"cpu": {"temperature": 45},
		"ddns_enabled": false,
		"flash_app": 1000000,
		"flash_free": 30000000,
		"flash_total": 134217728,
		"guest_ip": "192.168.9.1",
		"guest_netmask": "255.255.255.0",
		"ipv6_enabled": false,
		"lan_ip": "192.168.8.1",
		"lan_netmask": "255.255.255.0",
		"load_average": [0.1, 0.05, 0.01],
		"memory_buff_cache": 100000000,
		"memory_free": 200000000,
		"memory_total": 400000000,
		"mode": 0,
		"netnat_enabled": true,
		"timestamp": 1700000000,
		"tzoffset": "+0000",
		"uptime": 1234.56
	},
	"wifi": [
		{"band": "5G", "channel": 36, "encryption": "psk2", "guest": false, "hidden": false, "mld": false, "name": "default_radio0", "passwd": "wifi-password", "ssid": "test-5g", "up": true},
		{"band": "2G", "channel": 6, "encryption": "psk2", "guest": true, "hidden": false, "mld": false, "name": "guest2g", "passwd": "guest-password", "ssid": "test-guest", "up": false}
	]
}`

func newSystem() *System {
	return &System{
		timezone: TimezoneConfig{
			Zonename:            "UTC",
			Timezone:            "UTC0",
			TZOffset:            "+0000",
			AutoTimezoneEnabled: true,
			Localtime:           1700000000,
			Timestamp:           1700000000,
		},
	}
}

// TimezoneConfig returns the time zone configuration.
func (m *System) TimezoneConfig() TimezoneConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.timezone
}

// SetTimezoneConfig sets the time zone configuration.
func (m *System) SetTimezoneConfig(c TimezoneConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timezone = c
}

func (m *System) register(s *Server) {
	s.Handle("system", "get_status", func(json.RawMessage) (any, error) {
		return json.RawMessage(systemStatus), nil
	})
	s.Handle("system", "get_timezone_config", func(json.RawMessage) (any, error) {
		return m.TimezoneConfig(), nil
	})
	s.Handle("system", "set_timezone_config", m.setTimezoneConfig)
}

func (m *System) setTimezoneConfig(params json.RawMessage) (any, error) {
	var p struct {
		Zonename  *string `json:"zonename"`
		Timezone  *string `json:"timezone"`
		Localtime *int64  `json:"localtime"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Zonename != nil {
		m.timezone.Zonename = *p.Zonename
	}
	if p.Timezone != nil {
		m.timezone.Timezone = *p.Timezone
	}
	if p.Localtime != nil {
		m.timezone.Timestamp = *p.Localtime
	}
	m.timezone.TZOffset = tzOffset(m.timezone.Timezone, m.timezone.Timestamp)

	return json.RawMessage(`[]`), nil
}

// tzOffset returns the UTC offset, such as "+0530", of the POSIX TZ string tz at the Unix time t.
// The time package applies a POSIX TZ string only as the footer of TZif data (RFC 8536),
// so tzOffset makes TZif data with one zone, no transitions, and tz as the footer.
func tzOffset(tz string, t int64) string {
	header := append([]byte("TZif2"), make([]byte, 15)...)
	for _, n := range []uint32{0, 0, 0, 0, 1, 1} { // isutcnt, isstdcnt, leapcnt, timecnt, typecnt, charcnt
		header = binary.BigEndian.AppendUint32(header, n)
	}
	zone := make([]byte, 7) // one ttinfo with UTC offset 0 and an empty abbreviation

	var data []byte
	for range 2 { // the version 1 block and the version 2 block
		data = append(data, header...)
		data = append(data, zone...)
	}
	data = append(data, "\n"+tz+"\n"...)

	loc, err := time.LoadLocationFromTZData("", data)
	if err != nil {
		return "+0000"
	}
	return time.Unix(t, 0).In(loc).Format("-0700")
}
