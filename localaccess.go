package glinet

import "context"

// LocalAccessService calls the local-access module: the ports of the web UI, LuCI and SSH on the
// LAN, and the session timeout of the web UI (the local access card of the web UI view security).
type LocalAccessService service

// LocalAccessConfig is the local access configuration.
//
// HTTPPort and HTTPSPort are the ports of the web UI and of the JSON-RPC API. RedirectHTTPS
// redirects HTTP to HTTPS. LuCIHTTPPort, LuCIHTTPSPort and LuCIRedirectHTTPS are the same for
// LuCI; the web UI shows them only when LuCI is installed (luci get_status gives status 2).
// SessionTimeout is the session timeout of the web UI in seconds. SSHEnabled
// enables the SSH server and SSHPort is its port.
//
// Firmware 4.7.0 does not return ssh_enabled, so SSHEnabled is false there.
type LocalAccessConfig struct {
	HTTPPort          int  `json:"http_port"`
	HTTPSPort         int  `json:"https_port"`
	RedirectHTTPS     bool `json:"redirect_https"`
	LuCIHTTPPort      int  `json:"luci_http_port"`
	LuCIHTTPSPort     int  `json:"luci_https_port"`
	LuCIRedirectHTTPS bool `json:"luci_redirect_https"`
	SessionTimeout    int  `json:"session_timeout"`
	SSHEnabled        bool `json:"ssh_enabled"`
	SSHPort           int  `json:"ssh_port"`
}

// GetConfig returns the local access configuration.
func (s *LocalAccessService) GetConfig(ctx context.Context) (*LocalAccessConfig, error) {
	var res LocalAccessConfig
	if err := s.client.Call(ctx, "local-access", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig replaces the local access configuration with one local-access set_config that has
// all fields of config, as the web UI does, also the LuCI fields when LuCI is not installed.
//
// Not tested on a router: the call is from the web UI code of firmware 4.8.3. What firmware 4.7.0
// does with ssh_enabled is not known. The web UI permits:
//   - Ports 1 to 65535, each port different from the four other ports. For the four web ports,
//     it does not permit the ports that browsers block (for example 22, 25 and 6000).
//   - SessionTimeout 60 to 10800 (1 to 180 minutes) in steps of 60.
//
// The router returns these err_code values as an *RPCError: -1 when HTTPPort is in use, -2 when
// HTTPSPort is in use, -3 when the SSH port or a LuCI port is in use (err_msg tells which), and
// -4 when the admin password is too weak for SSH (the web UI shows this for SSHEnabled true).
//
// A change of HTTPPort, HTTPSPort or RedirectHTTPS changes the address of the web UI and of the
// API that this client calls: after the write, the web UI waits 3 s and loads the page from the
// new port. A client of the old address then gets no answer. A change of SSHEnabled or SSHPort
// can end SSH access.
func (s *LocalAccessService) SetConfig(ctx context.Context, config LocalAccessConfig) error {
	return s.client.Call(ctx, "local-access", "set_config", config, nil)
}
