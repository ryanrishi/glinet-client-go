package glinet

import "context"

// NetworkService calls the network module.
type NetworkService service

// WANCableStatus is the result of network check_wan_cable for the first wired WAN port.
// CableInserted is a runtime value.
type WANCableStatus struct {
	CableEnabled    bool `json:"cable_enabled"`
	CableInserted   bool `json:"cable_inserted"`
	MACCloneEnabled bool `json:"macclone_enabled"`
}

// CheckWANCable returns the state of the cable of the first wired WAN port.
func (s *NetworkService) CheckWANCable(ctx context.Context) (*WANCableStatus, error) {
	params := struct {
		SecondWAN bool `json:"secondwan"`
	}{SecondWAN: false}
	var res WANCableStatus
	if err := s.client.Call(ctx, "network", "check_wan_cable", params, &res); err != nil {
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
