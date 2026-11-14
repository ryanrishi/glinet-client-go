package glinet

import "context"

// AccessControlService calls the black_white_list module: the MAC access control of the client
// devices. In mode "black" the router blocks the MACs in the block list.
//
// The router keeps each MAC as written and compares MACs with their letter case: "02:00:00:00:00:a1"
// and "02:00:00:00:00:A1" are two entries, and a remove with the other case removes nothing.
// The web UI compares MACs without letter case and writes them in upper case.
type AccessControlService service

// AccessControlConfig is the result of black_white_list get_config. Mode is "black" (block list)
// or "white" (allow list). BlackMAC is the block list; the router leaves it out when it is empty.
// WhiteMAC is the allow list; it is from the web UI code and was not seen on a router.
type AccessControlConfig struct {
	Mode     string   `json:"mode"`
	BlackMAC []string `json:"black_mac,omitempty"`
	WhiteMAC []string `json:"white_mac,omitempty"`
}

type singleMACRequest struct {
	Mode    string `json:"mode"`
	Operate string `json:"operate"`
	MAC     string `json:"mac"`
}

// GetConfig returns the access control configuration.
func (s *AccessControlService) GetConfig(ctx context.Context) (*AccessControlConfig, error) {
	var res AccessControlConfig
	if err := s.client.Call(ctx, "black_white_list", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// AddBlockedMAC adds mac to the block list. The next GetConfig has it.
//
// The router does not check mac: it keeps any string, also "not-a-mac", but ignores a MAC with "-"
// separators, such as "02-00-00-00-00-a1". It does not add a MAC that the list has with the same
// letter case.
// The call sends mode "black"; what it does when the router is in mode "white" was not tested.
func (s *AccessControlService) AddBlockedMAC(ctx context.Context, mac string) error {
	return s.client.Call(ctx, "black_white_list", "set_single_mac", singleMACRequest{Mode: "black", Operate: "add", MAC: mac}, nil)
}

// RemoveBlockedMAC removes mac from the block list. The router returns no error when the list
// does not have mac with the same letter case.
func (s *AccessControlService) RemoveBlockedMAC(ctx context.Context, mac string) error {
	return s.client.Call(ctx, "black_white_list", "set_single_mac", singleMACRequest{Mode: "black", Operate: "del", MAC: mac}, nil)
}
