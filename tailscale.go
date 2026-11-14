package glinet

import "context"

// TailscaleService calls the tailscale module.
//
// The module has no secret in its configuration. The login of the router to a tailnet is
// interactive (the web UI calls tailscale get_auth_url and shows the URL); this client does not
// call it. The auth URL is a secret.
type TailscaleService service

// TailscaleConfig is the result of tailscale get_config.
//
// LANEnabled and WANEnabled let the tailnet reach the LAN and the WAN of the router. ExitNodeIP is
// the Tailscale IP of the exit node; the router leaves it out when it is not set. LANIP is the
// subnet of the LAN, for example "192.168.8.0/24"; it is read only.
type TailscaleConfig struct {
	Enabled    bool   `json:"enabled"`
	LANEnabled bool   `json:"lan_enabled"`
	WANEnabled bool   `json:"wan_enabled"`
	ExitNodeIP string `json:"exit_node_ip,omitempty"`
	LANIP      string `json:"lan_ip,omitempty"`
}

// SetTailscaleConfigRequest is the params of tailscale set_config.
//
// The router sets all fields to the sent values: a write without LANEnabled or WANEnabled sets it
// to false, and a write without ExitNodeIP removes the exit node. This was verified with Enabled
// false only.
type SetTailscaleConfigRequest struct {
	Enabled    bool   `json:"enabled"`
	LANEnabled bool   `json:"lan_enabled"`
	WANEnabled bool   `json:"wan_enabled"`
	ExitNodeIP string `json:"exit_node_ip,omitempty"`
}

// GetConfig returns the Tailscale configuration.
func (s *TailscaleService) GetConfig(ctx context.Context) (*TailscaleConfig, error) {
	var res TailscaleConfig
	if err := s.client.Call(ctx, "tailscale", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the Tailscale configuration. With config.Enabled false, the router stores the
// values, returns at once and the next GetConfig has them.
//
// Not verified on a router: a write with Enabled true, which starts Tailscale and connects to the
// Tailscale servers. The web UI sends all four fields to enable it (ExitNodeIP "" when the user
// chose no exit node), only Enabled false to disable it, and reads the configuration again 5 s
// after the write.
func (s *TailscaleService) SetConfig(ctx context.Context, config SetTailscaleConfigRequest) error {
	return s.client.Call(ctx, "tailscale", "set_config", config, nil)
}
