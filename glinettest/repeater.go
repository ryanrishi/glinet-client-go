package glinettest

import (
	"encoding/json"
	"maps"
)

// The default results of the repeater module: the shapes of the test router (firmware 4.8.3), with
// made-up values. The repeater is connected to the saved network "example".
const (
	repeaterConfig = `{"auto":1,"dfs_support":true,"macaddr":"02:00:00:00:00:38"}`

	repeaterSavedNetworks = `{"res":[{
		"auto_portal": false,
		"disguise": false,
		"key": "example-key",
		"macaddr": {"mode": "default", "update": "none"},
		"manual": false,
		"protocol": "dhcp",
		"ssid": "example"
	}]}`

	repeaterStatus = `{
		"band": "5g",
		"bare_mode": false,
		"bssid": "02:00:00:00:00:36",
		"channel": 48,
		"config": {
			"auto_portal": false,
			"disguise": false,
			"key": "example-key",
			"macaddr": {"mode": "default"},
			"manual": false,
			"protocol": "dhcp",
			"remember": true,
			"ssid": "example"
		},
		"connected": "40m,13s",
		"device": "radio0",
		"dfs": false,
		"fail_type": "",
		"htmode": "HT40",
		"ipv4": {"dns": ["192.0.2.1"], "gateway": "192.0.2.1", "ip": "192.0.2.20/24"},
		"macaddr": "02:00:00:00:00:38",
		"network": "wwan",
		"portal": false,
		"running": true,
		"signal": -53,
		"ssid": "example",
		"state": 2,
		"state_s": "connected",
		"wpa_status": {
			"address": "02:00:00:00:00:38",
			"bssid": "02:00:00:00:00:36",
			"freq": "5240",
			"key_mgmt": "WPA2-PSK",
			"mode": "station",
			"ssid": "example",
			"wpa_state": "COMPLETED"
		}
	}`
)

// Repeater serves the repeater module. The reads return made-up values in the shapes of the test
// router until a test or a write sets them.
//
// The write behaviour is an assumption, because the writes were not tested on a router. The fake
// does not check the values:
//   - set_config sets auto, and sets lock_band, or removes it when the params have no lock_band.
//   - connect sets get_status to state 2, state_s "connected", ssid and config (the params).
//     With remember true, it also saves the params without remember: it replaces the saved
//     network with the same ssid, or appends one.
//   - disconnect sets get_status to {"state":0}.
//   - remove_saved_ap removes each saved network with the ssid of the params.
type Repeater struct {
	config readResult
	saved  readResult
	status readResult
}

func newRepeater() *Repeater {
	m := &Repeater{}
	m.config.set(json.RawMessage(repeaterConfig))
	m.saved.set(json.RawMessage(repeaterSavedNetworks))
	m.status.set(json.RawMessage(repeaterStatus))
	return m
}

// Config returns the result of get_config.
func (m *Repeater) Config() json.RawMessage {
	return m.config.get()
}

// SetConfig sets the result of get_config.
func (m *Repeater) SetConfig(result json.RawMessage) {
	m.config.set(result)
}

// SavedNetworks returns the result of get_saved_ap_list.
func (m *Repeater) SavedNetworks() json.RawMessage {
	return m.saved.get()
}

// SetSavedNetworks sets the result of get_saved_ap_list.
func (m *Repeater) SetSavedNetworks(result json.RawMessage) {
	m.saved.set(result)
}

// Status returns the result of get_status.
func (m *Repeater) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *Repeater) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

func (m *Repeater) register(s *Server) {
	s.Handle("repeater", "get_config", m.config.handle)
	s.Handle("repeater", "get_saved_ap_list", m.saved.handle)
	s.Handle("repeater", "get_status", m.status.handle)
	s.Handle("repeater", "set_config", m.setConfig)
	s.Handle("repeater", "connect", m.connect)
	s.Handle("repeater", "disconnect", m.disconnect)
	s.Handle("repeater", "remove_saved_ap", m.removeSaved)
}

func (m *Repeater) setConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	if err := m.config.update(func(v any) (any, error) {
		c, _ := v.(map[string]any)
		if c == nil {
			c = map[string]any{}
		}
		if auto, ok := p["auto"]; ok {
			c["auto"] = auto
		}
		if band, ok := p["lock_band"]; ok {
			c["lock_band"] = band
		} else {
			delete(c, "lock_band")
		}
		return c, nil
	}); err != nil {
		return nil, err
	}
	return writeResult, nil
}

func (m *Repeater) connect(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	ssid, _ := p["ssid"].(string)
	if p["remember"] == true {
		network := maps.Clone(p)
		delete(network, "remember")
		if err := m.saved.update(func(v any) (any, error) {
			res := savedNetworks(v)
			for i, n := range res {
				if n["ssid"] == ssid {
					res[i] = network
					return map[string]any{"res": res}, nil
				}
			}
			return map[string]any{"res": append(res, network)}, nil
		}); err != nil {
			return nil, err
		}
	}
	if err := m.status.update(func(v any) (any, error) {
		st, _ := v.(map[string]any)
		if st == nil {
			st = map[string]any{}
		}
		st["state"] = 2
		st["state_s"] = "connected"
		st["ssid"] = ssid
		st["config"] = p
		return st, nil
	}); err != nil {
		return nil, err
	}
	return writeResult, nil
}

func (m *Repeater) disconnect(json.RawMessage) (any, error) {
	m.status.set(json.RawMessage(`{"state":0}`))
	return writeResult, nil
}

func (m *Repeater) removeSaved(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	if err := m.saved.update(func(v any) (any, error) {
		res := savedNetworks(v)
		kept := []map[string]any{}
		for _, n := range res {
			if n["ssid"] != p["ssid"] {
				kept = append(kept, n)
			}
		}
		return map[string]any{"res": kept}, nil
	}); err != nil {
		return nil, err
	}
	return writeResult, nil
}

// savedNetworks returns the items of a get_saved_ap_list result.
func savedNetworks(v any) []map[string]any {
	obj, _ := v.(map[string]any)
	items, _ := obj["res"].([]any)
	res := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if n, ok := item.(map[string]any); ok {
			res = append(res, n)
		}
	}
	return res
}
