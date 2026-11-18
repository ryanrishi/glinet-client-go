package glinet

import "context"

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
// {"secondwan": false}. The web UI code of firmware 4.8.3 that was read does not call this method,
// so "secondwan": true for SecondWANPort is not settled by the web UI code and was not tested on a
// router.
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
