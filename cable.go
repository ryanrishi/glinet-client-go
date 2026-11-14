package glinet

import "context"

// CableService calls the cable module.
type CableService service

// CableStatus is the result of cable get_status: the state of the wired WAN port.
// The values are runtime values.
//
// The 2023 API reference gives Status 0 as failed, 1 as connected, 2 as connecting and 3 as not
// connected physically. The test router, without a WAN cable, gave Mode 0 and Status 3.
type CableStatus struct {
	Mode      int             `json:"mode"`
	Status    int             `json:"status"`
	SecondWAN SecondWANStatus `json:"secondwan"`
}

// SecondWANStatus is the state of a second wired WAN port. A router without one gives ErrCode -4
// and ErrMsg "Unsupported features!" in a result without an error, so GetStatus returns no error.
type SecondWANStatus struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

// GetStatus returns the state of the wired WAN port.
func (s *CableService) GetStatus(ctx context.Context) (*CableStatus, error) {
	var res CableStatus
	if err := s.client.Call(ctx, "cable", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
