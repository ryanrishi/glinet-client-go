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

// System serves the system module. The results of get_info, get_status and disk_info are those of
// the test router until a test sets them. Timestamp is the router clock: set_timezone_config sets it
// from localtime and ignores autotimezone_enabled, like the router. It also sets TZOffset to the
// UTC offset of the POSIX TZ string timezone at that clock, with daylight saving time.
// A timezone that is not a valid POSIX TZ string gives "+0000".
//
// get_usb3_disable returns {"usb3_disable": USB3Disabled}, and set_usb3_disable sets it when the
// params have usb3_disable as a bool; another type gives ErrInvalidParams. These methods were not
// seen on a router: the shapes are from the web UI code of firmware 4.8.3, and the fake serves
// them also when SystemInfo has no USB 3.0 port.
type System struct {
	mu           sync.Mutex
	timezone     TimezoneConfig
	usb3Disabled bool

	info, status, diskInfo readResult
}

func newSystem() *System {
	m := &System{
		timezone: TimezoneConfig{
			Zonename:            "UTC",
			Timezone:            "UTC0",
			TZOffset:            "+0000",
			AutoTimezoneEnabled: true,
			Localtime:           1700000000,
			Timestamp:           1700000000,
		},
	}
	m.info.set(json.RawMessage(systemInfo))
	m.status.set(json.RawMessage(systemStatus))
	m.diskInfo.set(json.RawMessage(systemDiskInfo))
	return m
}

// Info returns the result of get_info.
func (m *System) Info() json.RawMessage {
	return m.info.get()
}

// SetInfo sets the result of get_info.
func (m *System) SetInfo(result json.RawMessage) {
	m.info.set(result)
}

// Status returns the result of get_status.
func (m *System) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *System) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

// DiskInfo returns the result of disk_info.
func (m *System) DiskInfo() json.RawMessage {
	return m.diskInfo.get()
}

// SetDiskInfo sets the result of disk_info.
func (m *System) SetDiskInfo(result json.RawMessage) {
	m.diskInfo.set(result)
}

// USB3Disabled returns true when the USB port works as USB 2.0.
func (m *System) USB3Disabled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usb3Disabled
}

// SetUSB3Disabled sets the USB port to USB 2.0 (true) or USB 3.0 (false).
func (m *System) SetUSB3Disabled(disabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.usb3Disabled = disabled
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
	s.Handle("system", "get_info", m.info.handle)
	s.Handle("system", "get_status", m.status.handle)
	s.Handle("system", "disk_info", m.diskInfo.handle)
	s.Handle("system", "get_timezone_config", func(json.RawMessage) (any, error) {
		return m.TimezoneConfig(), nil
	})
	s.Handle("system", "set_timezone_config", m.setTimezoneConfig)
	s.Handle("system", "get_usb3_disable", func(json.RawMessage) (any, error) {
		return map[string]bool{"usb3_disable": m.USB3Disabled()}, nil
	})
	s.Handle("system", "set_usb3_disable", m.setUSB3Disable)
}

func (m *System) setUSB3Disable(params json.RawMessage) (any, error) {
	var p struct {
		USB3Disable *bool `json:"usb3_disable"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}
	if p.USB3Disable != nil {
		m.SetUSB3Disabled(*p.USB3Disable)
	}
	return writeResult, nil
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
