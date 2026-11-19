package glinet

import (
	"context"
	"encoding/json"
)

// WifiService calls the wifi module.
type WifiService service

// WifiRadioStatus is an item of the result of wifi get_status: the state of a radio, for example
// Name "radio0" and State "ready". The values are runtime values. The router gives no Channel for
// some radios; then Channel is 0. On the test router, radio1 had no channel.
type WifiRadioStatus struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Channel int    `json:"channel,omitempty"`
}

// GetStatus returns the state of the radios.
func (s *WifiService) GetStatus(ctx context.Context) ([]WifiRadioStatus, error) {
	var res struct {
		Res []WifiRadioStatus `json:"res"`
	}
	if err := s.client.Call(ctx, "wifi", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return res.Res, nil
}

// WifiConfig is the result of wifi get_config: the configuration of the radios and their Wi-Fi
// interfaces. DFSSupport is true when the router can use DFS channels.
type WifiConfig struct {
	DFSSupport bool        `json:"dfs_support"`
	Radios     []WifiRadio `json:"res"`
}

// WifiRadio is the configuration of a radio, for example Device "radio0" with Band "5G" or
// Device "radio1" with Band "2G". Band is read only.
//
// Channel 0 is automatic. Channels is the channels that the radio can use. HWMode is one of
// HWModes, for example "11a/n/ac/ax". HTMode is the channel width in MHz as a string, for example
// "40", or "auto"; HTModes gives the widths that each hardware mode allows. TxPower is "Max",
// "High", "Medium" or "Low". Encryptions is the encryptions that the interfaces can use.
//
// Chan6GOnlyPSC is only for a radio with Band "6G": true limits the channels to the preferred
// scanning channels (WifiChannel.PSC). It is from the web UI code of firmware 4.8.3 and was not
// seen on a router. The web UI shows it as a switch with the label key wireless.enable_psc_label
// for the main interface of a 6G radio, and uses it to filter the channel list: when it is set to
// true and the channel is not a PSC, the web UI sets Channel to 0 (automatic).
type WifiRadio struct {
	Device        string        `json:"device"`
	Band          string        `json:"band"`
	Channel       int           `json:"channel"`
	Channels      []WifiChannel `json:"channels"`
	HWMode        string        `json:"hwmode"`
	HWModes       []string      `json:"hwmodes"`
	HTMode        string        `json:"htmode"`
	HTModes       WifiHTModes   `json:"htmodes"`
	TxPower       string        `json:"txpower"`
	RandomBSSID   bool          `json:"random_bssid"`
	Chan6GOnlyPSC *bool         `json:"chan_6g_only_psc,omitempty"`
	Encryptions   []string      `json:"encryptions"`
	Ifaces        []WifiIface   `json:"ifaces"`
}

// WifiChannel is a channel that a radio can use. DFS is true for a DFS channel. PSC is true for a
// 6 GHz preferred scanning channel.
type WifiChannel struct {
	Channel int  `json:"channel"`
	DFS     bool `json:"dfs"`
	PSC     bool `json:"psc"`
}

// WifiHTModes is the htmodes object of a radio: MaxWidth maps each hardware mode to its largest
// channel width in MHz, for example "11a/n/ac/ax" to 80. The web UI offers the widths 20, 40, ...
// up to MaxWidth[HWMode], and "auto" when Auto is true.
type WifiHTModes struct {
	Auto     bool
	MaxWidth map[string]int
}

// UnmarshalJSON decodes the htmodes object of the router.
func (m *WifiHTModes) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*m = WifiHTModes{MaxWidth: map[string]int{}}
	for k, v := range raw {
		var err error
		if k == "auto" {
			err = json.Unmarshal(v, &m.Auto)
		} else {
			var width int
			err = json.Unmarshal(v, &width)
			m.MaxWidth[k] = width
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON encodes m as the router does.
func (m WifiHTModes) MarshalJSON() ([]byte, error) {
	obj := map[string]any{"auto": m.Auto}
	for k, v := range m.MaxWidth {
		obj[k] = v
	}
	return json.Marshal(obj)
}

// WifiIface is the configuration of a Wi-Fi interface, for example Name "default_radio0" (the
// main Wi-Fi of radio0) or "guest5g" (Guest true). Key is the Wi-Fi password, a secret; the router
// returns it. Encryption is one of the Encryptions of the radio: "none", "psk2", "psk-mixed",
// "sae" or "sae-mixed" on the test router.
type WifiIface struct {
	Name       string `json:"name"`
	SSID       string `json:"ssid"`
	Key        string `json:"key"`
	Encryption string `json:"encryption"`
	Enabled    bool   `json:"enabled"`
	Hidden     bool   `json:"hidden"`
	Guest      bool   `json:"guest"`
}

// FindIface returns the interface with the name name and its radio. It returns nil, nil when
// there is no such interface.
func (c *WifiConfig) FindIface(name string) (*WifiRadio, *WifiIface) {
	for i := range c.Radios {
		r := &c.Radios[i]
		for j := range r.Ifaces {
			if r.Ifaces[j].Name == name {
				return r, &r.Ifaces[j]
			}
		}
	}
	return nil, nil
}

// GetConfig returns the configuration of the radios and their Wi-Fi interfaces. It has the Wi-Fi
// passwords.
func (s *WifiService) GetConfig(ctx context.Context) (*WifiConfig, error) {
	var res WifiConfig
	if err := s.client.Call(ctx, "wifi", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetWifiConfigRequest is the params of wifi set_config. The write is for the interface
// IfaceName. A field with no value is not sent. The web UI sends these params:
//   - IfaceName and Enabled to switch an interface on or off;
//   - for a guest interface: IfaceName, SSID, Encryption, Hidden, and Key unless Encryption is
//     "none";
//   - for a main interface: the guest fields and the radio fields Device, HWMode, TxPower and
//     RandomBSSID, Channel and HTMode unless the repeater uses the radio, and Chan6GOnlyPSC for a
//     radio with Band "6G" (firmware 4.8.3).
//
// Key is the Wi-Fi password, a secret: 8-63 printable ASCII characters or 64 hex digits, at most
// 63 characters for "sae" and "sae-mixed". SSID has 1-32 bytes. Channel 0 is automatic.
type SetWifiConfigRequest struct {
	IfaceName     string `json:"iface_name"`
	Enabled       *bool  `json:"enabled,omitempty"`
	SSID          string `json:"ssid,omitempty"`
	Encryption    string `json:"encryption,omitempty"`
	Hidden        *bool  `json:"hidden,omitempty"`
	Key           string `json:"key,omitempty"`
	Device        string `json:"device,omitempty"`
	HWMode        string `json:"hwmode,omitempty"`
	Channel       *int   `json:"channel,omitempty"`
	HTMode        string `json:"htmode,omitempty"`
	TxPower       string `json:"txpower,omitempty"`
	RandomBSSID   *bool  `json:"random_bssid,omitempty"`
	Chan6GOnlyPSC *bool  `json:"chan_6g_only_psc,omitempty"`
}

// SetConfig sets the configuration of a Wi-Fi interface and, for a main interface, of its radio.
// The web UI waits 5 s after the write and then reads the configuration again.
//
// Not verified on a router: the parameters come from the web UI code. A write to radio0 of the
// test router can drop the repeater uplink, which uses that radio.
func (s *WifiService) SetConfig(ctx context.Context, config SetWifiConfigRequest) error {
	return s.client.Call(ctx, "wifi", "set_config", config, nil)
}

// SetTxPower sets the transmit power of the radio device, for example "radio1", to txpower:
// "Max", "High", "Medium" or "Low". The web UI sends it at once when the select changes and waits
// 5 s.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WifiService) SetTxPower(ctx context.Context, device, txpower string) error {
	params := struct {
		TxPower string `json:"txpower"`
		Device  string `json:"device"`
	}{TxPower: txpower, Device: device}
	return s.client.Call(ctx, "wifi", "set_txpower", params, nil)
}
