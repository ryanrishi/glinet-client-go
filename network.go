package glinet

import (
	"context"
	"encoding/json"
)

// NetworkService calls the network module.
type NetworkService service

// WANCableStatus is the result of network check_wan_cable for a wired WAN port.
// CableInserted is a runtime value.
type WANCableStatus struct {
	CableEnabled    bool `json:"cable_enabled"`
	CableInserted   bool `json:"cable_inserted"`
	MACCloneEnabled bool `json:"macclone_enabled"`
}

// CheckWANCable returns the state of the cable of the wired WAN port port. It sends
// {"secondwan": bool}, as the cable methods do. The test router gave the fixture result for
// {"secondwan": false}. The web UI script netport.js of firmware 4.8.3 calls this method.
// "secondwan": true for SecondWANPort is not settled and was not tested on a router.
func (s *NetworkService) CheckWANCable(ctx context.Context, port WANPort) (*WANCableStatus, error) {
	var res WANCableStatus
	if err := s.client.Call(ctx, "network", "check_wan_cable", port.params(), &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ARPEntry is an entry of the ARP table of the router. The values are runtime values.
type ARPEntry struct {
	IP     string `json:"ip"`
	MAC    string `json:"mac"`
	Device string `json:"device"`
}

// ListARPEntries returns the ARP table of the router.
func (s *NetworkService) ListARPEntries(ctx context.Context) ([]ARPEntry, error) {
	var res struct {
		Entries []ARPEntry `json:"entries"`
	}
	if err := s.client.Call(ctx, "network", "get_arp_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Entries, nil
}

// NetworkAdvancedConfig is the result of network get_advance_config and the params of
// network set_advance_config (web UI view netadvanced). The router value of each field is 1
// (true) or 0 (false); the client decodes and encodes these numbers:
//   - FullConeNATEnabled (nat_enable) enables Full Cone NAT.
//   - SIPALGEnabled (sip_enable) enables the SIP ALG.
//
// The fields and the values are from the web UI code of firmware 4.8.3. The test router did not
// give a get_advance_config result: the web UI reads a missing value as 0, and compares the values
// with the numbers 1 and 0, so the router is expected to give numbers. Other fields of the result
// are not known.
type NetworkAdvancedConfig struct {
	FullConeNATEnabled bool `json:"nat_enable"`
	SIPALGEnabled      bool `json:"sip_enable"`
}

type networkAdvancedConfigJSON struct {
	FullConeNATEnabled flag `json:"nat_enable"`
	SIPALGEnabled      flag `json:"sip_enable"`
}

// MarshalJSON encodes c as the router does, with 0 or 1 for each field.
func (c NetworkAdvancedConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(networkAdvancedConfigJSON{flag(c.FullConeNATEnabled), flag(c.SIPALGEnabled)})
}

// UnmarshalJSON decodes the result of network get_advance_config. Each field accepts 0, 1, false
// and true.
func (c *NetworkAdvancedConfig) UnmarshalJSON(data []byte) error {
	var v networkAdvancedConfigJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*c = NetworkAdvancedConfig{bool(v.FullConeNATEnabled), bool(v.SIPALGEnabled)}
	return nil
}

// GetAdvancedConfig returns the advanced network settings.
func (s *NetworkService) GetAdvancedConfig(ctx context.Context) (*NetworkAdvancedConfig, error) {
	var res NetworkAdvancedConfig
	if err := s.client.Call(ctx, "network", "get_advance_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetAdvancedConfig sets the advanced network settings. It sends
// {"nat_enable": 0|1, "sip_enable": 0|1}, always with both values, as the web UI does.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI waits 2 s after the
// write and then reads the settings again. What the router restarts for the write is not known.
func (s *NetworkService) SetAdvancedConfig(ctx context.Context, config NetworkAdvancedConfig) error {
	return s.client.Call(ctx, "network", "set_advance_config", config, nil)
}
