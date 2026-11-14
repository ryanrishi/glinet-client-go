package glinet

import "context"

// IPv6Service calls the ipv6 module.
type IPv6Service service

// IPv6Config is the IPv6 configuration of the router: the result of ipv6 get_ipv6.
//
// LANMode is "native", "relay" (Passthrough in the web UI), "nat6" or "static". LANDNSMode true
// means that the LAN clients get the DNS servers automatically; false means LANDNS1 and LANDNS2.
// LANIP is the IPv6 address of the LAN with a prefix length 48-64, for LANMode "static".
// The test router, with IPv6 disabled, gave only Enable, LANMode and LANDNSMode; LANIP, LANDNS1
// and LANDNS2 come from the web UI code. The web UI has no WAN setting for IPv6.
type IPv6Config struct {
	Enable     bool   `json:"enable"`
	LANMode    string `json:"lan_mode"`
	LANDNSMode bool   `json:"lan_dns_mode"`
	LANIP      string `json:"lan_ip,omitempty"`
	LANDNS1    string `json:"lan_dns1,omitempty"`
	LANDNS2    string `json:"lan_dns2,omitempty"`
}

// GetConfig returns the IPv6 configuration.
func (s *IPv6Service) GetConfig(ctx context.Context) (*IPv6Config, error) {
	var res IPv6Config
	if err := s.client.Call(ctx, "ipv6", "get_ipv6", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the IPv6 configuration. It sends the fields as the web UI does: only enable when
// Enable is false; else also lan_mode and lan_dns_mode, lan_ip for LANMode "static", and lan_dns1
// and a non-empty lan_dns2 when LANDNSMode is false. The web UI reads the configuration again 10 s
// after the write.
//
// Not verified on a router: the parameters come from the web UI code. The web UI says that IPv6
// does not work together with the firewall, VPN and cloud features.
func (s *IPv6Service) SetConfig(ctx context.Context, config IPv6Config) error {
	params := map[string]any{"enable": config.Enable}
	if config.Enable {
		params["lan_mode"] = config.LANMode
		params["lan_dns_mode"] = config.LANDNSMode
		if config.LANMode == "static" {
			params["lan_ip"] = config.LANIP
		}
		if !config.LANDNSMode {
			params["lan_dns1"] = config.LANDNS1
			if config.LANDNS2 != "" {
				params["lan_dns2"] = config.LANDNS2
			}
		}
	}
	return s.client.Call(ctx, "ipv6", "set_ipv6", params, nil)
}
