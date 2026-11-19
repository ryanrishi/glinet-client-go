package glinet

import "context"

// AccessControlService calls the black_white_list module: the MAC access control of the client
// devices. In mode "black" the router blocks the MACs in the block list. In mode "white" the
// allow list applies.
//
// The router keeps each MAC as written and compares MACs with their letter case: "02:00:00:00:00:a1"
// and "02:00:00:00:00:A1" are two entries, and a remove with the other case removes nothing.
// The web UI compares MACs without letter case and writes them in upper case.
//
// Mode "black" was tested on a router (firmware 4.7.0). An allow-list write with set_single_mac
// in mode "black" was tested on a router (firmware 4.8.3). Mode "white", SetConfig with the allow
// list and the mode switch are from the web UI code of firmware 4.8.3 and were not tested on a
// router.
// The web UI does not let the user save an allow list without the MAC of the client that shows
// the web UI (label key clients.whitelist_self_error_tips), or a block list with that MAC.
type AccessControlService service

// The modes of the black_white_list module.
const (
	AccessControlBlockList = "black"
	AccessControlAllowList = "white"
)

// AccessControlConfig is the result of black_white_list get_config. Mode is
// AccessControlBlockList or AccessControlAllowList. BlackMAC is the block list and WhiteMAC is
// the allow list; the router leaves a list out when it is empty.
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

// SetConfig sets the mode to mode and replaces the list of that mode with macs, with one
// black_white_list set_config {"mode": mode, "mac": macs}, as the access control dialog of the
// web UI does. An empty macs empties the list.
//
// Verified on a router (firmware 4.7.0) only with mode "black": the router replaces the block
// list with macs as written, also with duplicates, and does not check the MACs. The web UI sends
// at most 1000 MACs, in upper case, with ":" separators. Not tested on a router: mode "white", a
// mode switch, and what a write does with the list of the other mode.
func (s *AccessControlService) SetConfig(ctx context.Context, mode string, macs []string) error {
	params := struct {
		Mode string   `json:"mode"`
		MAC  []string `json:"mac"`
	}{Mode: mode, MAC: append([]string{}, macs...)}
	return s.client.Call(ctx, "black_white_list", "set_config", params, nil)
}

// AddBlockedMAC adds mac to the block list. The next GetConfig has it.
//
// The router does not check mac: it keeps any string, also "not-a-mac", but ignores a MAC with "-"
// separators, such as "02-00-00-00-00-a1". It does not add a MAC that the list has with the same
// letter case.
// The call sends mode "black"; what it does when the router is in mode "white" was not tested.
func (s *AccessControlService) AddBlockedMAC(ctx context.Context, mac string) error {
	return s.setSingleMAC(ctx, AccessControlBlockList, "add", mac)
}

// RemoveBlockedMAC removes mac from the block list. The router returns no error when the list
// does not have mac with the same letter case.
func (s *AccessControlService) RemoveBlockedMAC(ctx context.Context, mac string) error {
	return s.setSingleMAC(ctx, AccessControlBlockList, "del", mac)
}

// AddAllowedMAC adds mac to the allow list, with black_white_list set_single_mac and mode
// "white". The web UI sends this call with the current mode, so it sends mode "white" only when
// the router is in mode "white".
//
// Tested on a router (GL-AXT1800, firmware 4.8.3) in mode "black": the router added mac to the
// allow list, and the mode stayed "black". So the allow list can be filled before a change to
// mode "white". Not tested on a router in mode "white".
func (s *AccessControlService) AddAllowedMAC(ctx context.Context, mac string) error {
	return s.setSingleMAC(ctx, AccessControlAllowList, "add", mac)
}

// RemoveAllowedMAC removes mac from the allow list, with black_white_list set_single_mac and mode
// "white".
//
// Tested on a router (GL-AXT1800, firmware 4.8.3) in mode "black": the router removed mac from
// the allow list. Not tested on a router in mode "white".
func (s *AccessControlService) RemoveAllowedMAC(ctx context.Context, mac string) error {
	return s.setSingleMAC(ctx, AccessControlAllowList, "del", mac)
}

func (s *AccessControlService) setSingleMAC(ctx context.Context, mode, operate, mac string) error {
	return s.client.Call(ctx, "black_white_list", "set_single_mac", singleMACRequest{Mode: mode, Operate: operate, MAC: mac}, nil)
}
