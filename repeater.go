package glinet

import (
	"context"
	"encoding/json"
)

// RepeaterService calls the repeater module: the Wi-Fi uplink ("repeater", interface wwan).
//
// A repeater write on a router whose uplink is the repeater can cut the Internet access of the
// router and its clients, and a connect to a network that the router cannot reach leaves the
// router without the repeater uplink. The reads are verified on a router (firmware 4.7.0 and
// 4.8.3). The writes come from the web UI code of firmware 4.8.3 and were not tested on a router.
type RepeaterService service

// RepeaterSwitchMode is the "auto" value of the repeater configuration: what the router does with
// the other saved networks. The names come from the label keys of the web UI of firmware 4.8.3
// (repeater.no_switching, repeater.auto_switching, repeater.auto_switching_without_network).
type RepeaterSwitchMode int

const (
	RepeaterNoSwitching                 RepeaterSwitchMode = 0
	RepeaterAutoSwitching               RepeaterSwitchMode = 1
	RepeaterAutoSwitchingWithoutNetwork RepeaterSwitchMode = 2
)

// UnmarshalJSON decodes a number, as firmware 4.8.3 returns it, or a bool, as firmware 4.7.0
// returns it. The client reads true as RepeaterAutoSwitching and false as RepeaterNoSwitching;
// this mapping is an assumption.
func (m *RepeaterSwitchMode) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "true":
		*m = RepeaterAutoSwitching
		return nil
	case "false":
		*m = RepeaterNoSwitching
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*m = RepeaterSwitchMode(n)
	return nil
}

// RepeaterConfig is the result of repeater get_config.
//
// LockBand is the band that the repeater uses, for example "2g" or "5g", or "" for all bands.
// The test router, with all bands, did not return lock_band. DFSSupport and MACAddr are read-only
// values; MACAddr is the MAC address of the repeater interface. The web UI also ignores
// status_interval of the result, which the test router did not return.
type RepeaterConfig struct {
	Auto       RepeaterSwitchMode `json:"auto"`
	LockBand   string             `json:"lock_band,omitempty"`
	DFSSupport bool               `json:"dfs_support"`
	MACAddr    string             `json:"macaddr"`
}

// SetRepeaterConfigRequest is the params of repeater set_config. The web UI shows LockBand only
// on a router with two or more bands, and does not send it when it is "" (all bands).
type SetRepeaterConfigRequest struct {
	Auto     RepeaterSwitchMode `json:"auto"`
	LockBand string             `json:"lock_band,omitempty"`
}

// RepeaterNetwork is a Wi-Fi network of the repeater: an item of repeater get_saved_ap_list, the
// config of repeater get_status, and the params of repeater connect.
//
//   - Key is the Wi-Fi key, or the password of an enterprise (EAP) network. It is a secret, and
//     get_saved_ap_list and get_status return it. Identity is the user of an enterprise network.
//     The web UI sends Key and Identity only when they are not empty.
//   - BSSID locks the repeater to one access point of the network.
//   - Protocol is "dhcp" or "static". For "static", IP, Netmask, Gateway and DNS (one or two
//     servers) are the static address. Some saved networks of the test router had no protocol.
//   - Disguise is the disguise option of the web UI (label key repeater.enable_disguise_label).
//     With Disguise true, the web UI sends no MACAddr, TTL and TTLIPv6.
//   - MACAddr is the MAC address that the repeater uses on this network.
//   - TTL and TTLIPv6 (1-255) and MTU (576-9200) are optional; 0 means not set.
//   - Manual: the web UI sends the value that it read, and false for a new network. Its meaning
//     is not known.
//   - AutoPortal is the captive portal option of the web UI (label key
//     repeater.allow_auto_portal_label).
//   - Remember saves the network in the list of saved networks. The web UI always sends true.
//     get_status returns it in config; get_saved_ap_list does not return it.
type RepeaterNetwork struct {
	SSID       string       `json:"ssid"`
	Key        string       `json:"key,omitempty"`
	Identity   string       `json:"identity,omitempty"`
	BSSID      string       `json:"bssid,omitempty"`
	Protocol   string       `json:"protocol,omitempty"`
	IP         string       `json:"ip,omitempty"`
	Netmask    string       `json:"netmask,omitempty"`
	Gateway    string       `json:"gateway,omitempty"`
	DNS        []string     `json:"dns,omitempty"`
	Disguise   bool         `json:"disguise"`
	MACAddr    *RepeaterMAC `json:"macaddr,omitempty"`
	TTL        int          `json:"ttl,omitempty"`
	TTLIPv6    int          `json:"ttl_ipv6,omitempty"`
	MTU        int          `json:"mtu,omitempty"`
	Manual     bool         `json:"manual"`
	AutoPortal bool         `json:"auto_portal"`
	Remember   bool         `json:"remember"`
}

// RepeaterMAC is the MAC address that the repeater uses on a network.
//
// Mode is "default" (the MAC address of the router), "clone" (MAC, the MAC address of a client)
// or "random" (MAC, a random MAC address). For mode "random", Update is when the router changes
// the MAC address: "none", "time" (every Period hours; the web UI sends whole days) or "reboot".
// The web UI sends Period with Update "time" and "reboot". The test router returned Update "none"
// with mode "default".
type RepeaterMAC struct {
	Mode   string `json:"mode"`
	MAC    string `json:"macaddr,omitempty"`
	Update string `json:"update,omitempty"`
	Period int    `json:"period,omitempty"`
}

// RepeaterStatus is the result of repeater get_status. The values are runtime values.
//
// State is 0 when the repeater is not connected (web UI code) and 2 when it is connected, with
// StateS "connected". The test router also gave StateS "failed" and "connecting". FailType is
// "key" (wrong key) or "not-found" (network not found) in the web UI code. Connected is the
// connection time, for example "40m,13s". Config is the network of the repeater; it has the Wi-Fi
// key. EAP and PortalURL come from the web UI code and were not seen on a router. BareMode is
// the captive portal mode of the web UI. WPAStatus is the wpa_supplicant status, with string
// values.
type RepeaterStatus struct {
	State     int               `json:"state"`
	StateS    string            `json:"state_s"`
	Running   bool              `json:"running"`
	FailType  string            `json:"fail_type"`
	SSID      string            `json:"ssid"`
	BSSID     string            `json:"bssid"`
	Band      string            `json:"band"`
	Channel   int               `json:"channel"`
	HTMode    string            `json:"htmode"`
	DFS       bool              `json:"dfs"`
	Signal    int               `json:"signal"`
	Device    string            `json:"device"`
	Network   string            `json:"network"`
	MACAddr   string            `json:"macaddr"`
	Connected string            `json:"connected"`
	EAP       bool              `json:"eap,omitempty"`
	Portal    bool              `json:"portal"`
	PortalURL string            `json:"portal_url,omitempty"`
	BareMode  bool              `json:"bare_mode"`
	Config    *RepeaterNetwork  `json:"config,omitempty"`
	IPv4      *UplinkIPv4       `json:"ipv4,omitempty"`
	WPAStatus map[string]string `json:"wpa_status,omitempty"`
}

// UplinkIPv4 is the IPv4 address of an uplink. IP has a prefix length, for example
// "192.0.2.10/24".
type UplinkIPv4 struct {
	IP      string   `json:"ip"`
	Gateway string   `json:"gateway"`
	DNS     []string `json:"dns"`
}

// GetConfig returns the repeater configuration.
func (s *RepeaterService) GetConfig(ctx context.Context) (*RepeaterConfig, error) {
	var res RepeaterConfig
	if err := s.client.Call(ctx, "repeater", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the repeater configuration.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not known: whether a write
// without lock_band (LockBand "") removes the band lock, and whether firmware 4.7.0, which
// returns auto as a bool, accepts a number.
func (s *RepeaterService) SetConfig(ctx context.Context, config SetRepeaterConfigRequest) error {
	return s.client.Call(ctx, "repeater", "set_config", config, nil)
}

// GetStatus returns the repeater status.
func (s *RepeaterService) GetStatus(ctx context.Context) (*RepeaterStatus, error) {
	var res RepeaterStatus
	if err := s.client.Call(ctx, "repeater", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ListSavedNetworks returns the saved networks of the repeater. Each item can have the Wi-Fi key.
func (s *RepeaterService) ListSavedNetworks(ctx context.Context) ([]RepeaterNetwork, error) {
	var res struct {
		Res []RepeaterNetwork `json:"res"`
	}
	if err := s.client.Call(ctx, "repeater", "get_saved_ap_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Res, nil
}

// RemoveSavedNetwork removes the saved network ssid, with repeater remove_saved_ap {"ssid": ssid}.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not known: what the router does
// when the repeater is connected to ssid.
func (s *RepeaterService) RemoveSavedNetwork(ctx context.Context, ssid string) error {
	params := struct {
		SSID string `json:"ssid"`
	}{SSID: ssid}
	return s.client.Call(ctx, "repeater", "remove_saved_ap", params, nil)
}

// Connect connects the repeater to network, with repeater connect. With network.Remember true,
// the router also saves the network. The web UI has no other call that changes a saved network:
// to change one, it connects to it with the new values.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI waits up to 30 s for
// the result. The connect replaces the current uplink network of the repeater: when the router
// cannot connect to network, the repeater uplink stays down. In bridge mode, the web UI sends
// only ssid, key and identity; Connect also sends the fields that have no omitempty.
func (s *RepeaterService) Connect(ctx context.Context, network RepeaterNetwork) error {
	return s.client.Call(ctx, "repeater", "connect", network, nil)
}

// Disconnect disconnects the repeater, with repeater disconnect {}. The repeater uplink goes
// down.
//
// From the web UI code of firmware 4.8.3, not tested on a router.
func (s *RepeaterService) Disconnect(ctx context.Context) error {
	return s.client.Call(ctx, "repeater", "disconnect", nil, nil)
}
