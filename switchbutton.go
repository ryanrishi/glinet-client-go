package glinet

import (
	"context"
	"encoding/json"
)

// SwitchButtonService calls the switch-button module: the function of the side switch of the
// router (SystemInfo.HardwareFeature.SwitchButton). The switch turns the function on in one
// position and off in the other.
type SwitchButtonService service

// SwitchButtonConfig is the function of the switch. Func is "none" or a Func of GetFuncs, and
// SubFunc is one of its SubFuncs, or empty when it has none.
//
// The values on firmware 4.8.3 include "adguardhome", "led", "repeater", "tor", "wifi" with
// SubFunc "main_wifi" or "guest_wifi", and "vpn" with SubFunc a tunnel ID of vpn-client
// get_tunnel as a string, such as "10". Firmware 4.7.0 has "openvpn" and "wireguard" in place
// of "vpn", with no SubFunc. What get_config returns when the function is "none" was not seen
// on a router.
type SwitchButtonConfig struct {
	Func    string `json:"func"`
	SubFunc string `json:"sub_func,omitempty"`
}

// SwitchButtonFunc is a function that the switch can have. SubFuncs has only Func values.
type SwitchButtonFunc struct {
	Func     string             `json:"func"`
	SubFuncs []SwitchButtonFunc `json:"sub_func,omitempty"`
}

// UnmarshalJSON decodes a function of firmware 4.8.3, which is an object, and a function of
// firmware 4.7.0, which is a string with only Func.
func (f *SwitchButtonFunc) UnmarshalJSON(data []byte) error {
	var name string
	if json.Unmarshal(data, &name) == nil {
		*f = SwitchButtonFunc{Func: name}
		return nil
	}
	type plain SwitchButtonFunc
	return json.Unmarshal(data, (*plain)(f))
}

// GetConfig returns the function of the switch.
func (s *SwitchButtonService) GetConfig(ctx context.Context) (*SwitchButtonConfig, error) {
	var res SwitchButtonConfig
	if err := s.client.Call(ctx, "switch-button", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetFuncs returns the functions that the switch can have. The list depends on the router and
// its configuration, for example the VPN tunnels. "none" is not in the list.
func (s *SwitchButtonService) GetFuncs(ctx context.Context) ([]SwitchButtonFunc, error) {
	var res struct {
		Funcs []SwitchButtonFunc `json:"funcs"`
	}
	if err := s.client.Call(ctx, "switch-button", "get_funcs", nil, &res); err != nil {
		return nil, err
	}

	return res.Funcs, nil
}

// SetConfig sets the function of the switch, with switch-button set_config
// {"func": config.Func, "sub_func": config.SubFunc, "sync": sync}. An empty SubFunc is not sent.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI sends sync true only
// when the user accepts to sync the function with the current switch position, after
// switch-button check_sync_status returned sync_status 0. Each other write sends sync false.
// What sync true does on the router is not known: it can turn the function on or off at once.
// The web UI does not let the user select a function that is not installed (AdGuard Home, Tor) or
// a VPN function when VPN is not supported. In global proxy mode of the VPN, the web UI shows one
// "VPN client" item and sends "vpn" with the first SubFunc of the "vpn" function.
func (s *SwitchButtonService) SetConfig(ctx context.Context, config SwitchButtonConfig, sync bool) error {
	params := struct {
		SwitchButtonConfig
		Sync bool `json:"sync"`
	}{SwitchButtonConfig: config, Sync: sync}
	return s.client.Call(ctx, "switch-button", "set_config", params, nil)
}
