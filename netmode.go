package glinet

import (
	"context"
	"encoding/json"
	"fmt"
)

// NetworkModeService calls the netmode module: the network mode of the router (web UI view bridge,
// menu "Network Mode").
//
// A change of the network mode reconfigures the whole router. In the modes NetworkModeAccessPoint,
// NetworkModeExtender and NetworkModeWDS the router gets its address from the upstream network, so
// it is no longer at its LAN address: the web UI then tells the user to find the router on the
// upstream network. In the other modes, the web UI shows only some views (key show_mode in
// ui get_menu_list), and most modules of this client are for the router mode.
type NetworkModeService service

// The network modes that the web UI can set. The web UI shows each mode only when the router
// supports it:
//   - NetworkModeAccessPoint: when "ap" is not in SystemInfo.HiddenFeatures. The web UI calls it
//     "Bridge" when the router has no Wi-Fi, or no extender and WDS mode.
//   - NetworkModeExtender ("relay", repeater bridge): when the router has a radio
//     (SystemInfo.HardwareFeature.Radio) and "relay" is not hidden.
//   - NetworkModeWDS: when the router has a radio, HardwareFeature.NoWDS is false, and "wds" is
//     not hidden.
//   - NetworkModePassthrough (IP passthrough): when SoftwareFeature.Passthrough is true and
//     "passthrough" is not hidden. The test router does not support it.
//
// The menu list of the router also has the mode "mesh", which the web UI cannot set.
const (
	NetworkModeRouter      = "router"
	NetworkModeAccessPoint = "ap"
	NetworkModeExtender    = "relay"
	NetworkModeWDS         = "wds"
	NetworkModePassthrough = "passthrough"
)

// NetworkModeConfig is the result of netmode get_mode. The test router gave only Mode "router".
// The web UI reads Mask and MAC for NetworkModePassthrough (see NetworkModeSettings), with the
// defaults "auto" and "" when the result does not have them. These fields were not seen on a
// router.
type NetworkModeConfig struct {
	Mode string `json:"mode"`
	Mask string `json:"mask,omitempty"`
	MAC  string `json:"mac,omitempty"`
}

// NetworkModeSettings is the network mode that SetMode sets.
//
// For NetworkModeExtender and NetworkModeWDS, the router connects to the upstream Wi-Fi network
// SSID. Key is its password, and Identity is the user name of an 802.1X (EAP) network; for EAP,
// Key is the password of that user. Key is a secret.
//
// For NetworkModePassthrough, Mask is "auto" or "24" (the web UI label is "force /24 subnet"),
// and MAC is an optional MAC address (empty for none) that is not 00:00:00:00:00:00.
type NetworkModeSettings struct {
	Mode     string
	SSID     string
	Key      string
	Identity string
	Mask     string
	MAC      string
}

// NetworkModeResult is the result of netmode set_mode. After a change to NetworkModeRouter, the
// web UI opens the web UI at LANIP, the new LAN address of the router. The web UI reads LANIP only
// after a change to the router mode; what the result has for the other modes is not known.
type NetworkModeResult struct {
	LANIP string `json:"lanip,omitempty"`
}

// GetMode returns the network mode.
func (s *NetworkModeService) GetMode(ctx context.Context) (*NetworkModeConfig, error) {
	var res NetworkModeConfig
	if err := s.client.Call(ctx, "netmode", "get_mode", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetMode sets the network mode with netmode set_mode. It sends the params that the web UI sends
// for settings.Mode:
//   - NetworkModeRouter and NetworkModeAccessPoint: {"mode"}.
//   - NetworkModeExtender and NetworkModeWDS: {"mode", "ssid"}, and "key" and "identity" only when
//     they are not empty.
//   - NetworkModePassthrough: {"mode", "mask", "mac"}, also when mac is empty.
//
// Other modes send {"mode"}.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not settled: what the router does
// with a call that the web UI does not make. The web UI offers only the change from the router mode
// to another mode and back to the router mode, never from one other mode to another. Before the
// change to NetworkModeAccessPoint it requires CableEnabled true from Network.CheckWANCable for
// the first WAN port, and before the change to NetworkModeExtender or NetworkModeWDS it requires
// CableEnabled false for each WAN port. In both cases it also requires MACCloneEnabled false.
// Before the change to NetworkModePassthrough it requires that cable get_status does not give
// mode 3 (edge router) and that only one WAN port is used.
//
// Side effects: the web UI waits up to 60 s for the answer, then 2 s (access point, extender,
// WDS), 10 s (router, passthrough) or 40 s (from passthrough to router). The router can then be at
// another address (see NetworkModeService), and the session can be invalid. The web UI gives these
// meanings to the err_code of the result: 1 the connection to the upstream Wi-Fi failed, 2 the
// router got no IP address from the upstream network, 3 the upstream network conflicts with the
// LAN, 4 the Wi-Fi password (or the EAP account) is wrong.
func (s *NetworkModeService) SetMode(ctx context.Context, settings NetworkModeSettings) (*NetworkModeResult, error) {
	params := map[string]any{"mode": settings.Mode}
	switch settings.Mode {
	case NetworkModeExtender, NetworkModeWDS:
		params["ssid"] = settings.SSID
		if settings.Key != "" {
			params["key"] = settings.Key
		}
		if settings.Identity != "" {
			params["identity"] = settings.Identity
		}
	case NetworkModePassthrough:
		params["mask"] = settings.Mask
		params["mac"] = settings.MAC
	}

	var raw json.RawMessage
	if err := s.client.Call(ctx, "netmode", "set_mode", params, &raw); err != nil {
		return nil, err
	}

	var res NetworkModeResult
	// The result of a change to a mode other than the router mode can be [].
	if string(raw) == "[]" {
		return &res, nil
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("glinet: netmode set_mode: decode result: %w", err)
	}
	return &res, nil
}
