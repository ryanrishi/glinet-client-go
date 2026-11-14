package glinettest

import (
	"encoding/json"
	"slices"
)

// Wifi serves the get_status, get_config, set_config and set_txpower methods of the wifi module.
// The results are those of the test router, with made-up SSIDs and keys, until a test or a write
// sets them.
//
// The write behaviour is an assumption, because the writes were not verified on a router:
// set_config finds the interface iface_name in get_config, sets each interface field that it gets
// (enabled, ssid, encryption, hidden, key) in the interface and each radio field (hwmode, channel,
// htmode, txpower, random_bssid) in its radio, and keeps the other fields. It ignores device.
// set_txpower sets the txpower of the radio device. Not known: which fields the router keeps when
// a write does not send them. It does not check the values. An unknown interface or radio gives
// ErrInvalidParams.
type Wifi struct {
	status readResult
	config readResult
}

var (
	wifiIfaceFields = []string{"enabled", "ssid", "encryption", "hidden", "key"}
	wifiRadioFields = []string{"hwmode", "channel", "htmode", "txpower", "random_bssid"}
)

func newWifi() *Wifi {
	m := &Wifi{}
	m.status.set(json.RawMessage(wifiStatus))
	m.config.set(json.RawMessage(wifiConfig))
	return m
}

// Status returns the result of get_status.
func (m *Wifi) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *Wifi) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

// Config returns the result of get_config.
func (m *Wifi) Config() json.RawMessage {
	return m.config.get()
}

// SetConfig sets the result of get_config.
func (m *Wifi) SetConfig(result json.RawMessage) {
	m.config.set(result)
}

func (m *Wifi) register(s *Server) {
	s.Handle("wifi", "get_status", m.status.handle)
	s.Handle("wifi", "get_config", m.config.handle)
	s.Handle("wifi", "set_config", m.setConfig)
	s.Handle("wifi", "set_txpower", m.setTxPower)
}

// radios returns the radio objects of a get_config result.
func radios(config any) []map[string]any {
	obj, _ := config.(map[string]any)
	items, _ := obj["res"].([]any)
	var res []map[string]any
	for _, item := range items {
		if radio, ok := item.(map[string]any); ok {
			res = append(res, radio)
		}
	}
	return res
}

func copyFields(dst, src map[string]any, fields []string) {
	for k, v := range src {
		if slices.Contains(fields, k) {
			dst[k] = v
		}
	}
}

func (m *Wifi) setConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}

	err = m.config.update(func(config any) (any, error) {
		for _, radio := range radios(config) {
			ifaces, _ := radio["ifaces"].([]any)
			for _, item := range ifaces {
				if iface, ok := item.(map[string]any); ok && iface["name"] == p["iface_name"] {
					copyFields(iface, p, wifiIfaceFields)
					copyFields(radio, p, wifiRadioFields)
					return config, nil
				}
			}
		}
		return nil, ErrInvalidParams
	})
	if err != nil {
		return nil, err
	}
	return writeResult, nil
}

func (m *Wifi) setTxPower(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}

	err = m.config.update(func(config any) (any, error) {
		for _, radio := range radios(config) {
			if radio["device"] == p["device"] {
				copyFields(radio, p, []string{"txpower"})
				return config, nil
			}
		}
		return nil, ErrInvalidParams
	})
	if err != nil {
		return nil, err
	}
	return writeResult, nil
}
