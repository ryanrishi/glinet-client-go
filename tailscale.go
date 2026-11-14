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

// TailscaleStatus is the result of tailscale get_status.
//
// Status is nil while Tailscale is disabled. While it is enabled and the router is not logged in
// to a tailnet, Status is 1 ("login needed"). The other values are from the web UI code and not
// verified: 3 running, 4 connecting. DNS is the DNS servers of the router.
type TailscaleStatus struct {
	Status *int     `json:"status,omitempty"`
	DNS    []string `json:"dns"`
}

// SetTailscaleConfigRequest is the params of tailscale set_config.
//
// The router sets all fields to the sent values, also with Enabled false: a write without
// LANEnabled or WANEnabled sets it to false, and a write without ExitNodeIP removes the exit node.
// So a disable ({"enabled":false}, as the web UI sends it) also sets LANEnabled and WANEnabled to
// false.
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

// GetStatus returns the Tailscale status. The router takes about 2 s to answer while Tailscale is
// disabled and about 45 ms while it is enabled.
func (s *TailscaleService) GetStatus(ctx context.Context) (*TailscaleStatus, error) {
	var res TailscaleStatus
	if err := s.client.Call(ctx, "tailscale", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the Tailscale configuration. The router returns after about 15 ms, also when the
// write starts or stops Tailscale, and the next GetConfig has the values. A write with Enabled
// true starts Tailscale, which connects to the Tailscale servers. After a write with Enabled
// false, GetStatus has no status within 1 s.
//
// Observed on a test router that was never logged in to Tailscale: GetStatus has status 1 within
// 2 s of the enable, and the firewall zones, the firewall rules and the DNS servers did not
// change, also with LANEnabled true. Not verified: the status and the firewall of a router that
// is logged in.
func (s *TailscaleService) SetConfig(ctx context.Context, config SetTailscaleConfigRequest) error {
	return s.client.Call(ctx, "tailscale", "set_config", config, nil)
}
