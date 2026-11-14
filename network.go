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
