package glinet

import "context"

// DDNSService calls the ddns module: the dynamic DNS service of GL.iNet.
type DDNSService service

// DDNSConfig is the result of ddns get_config. DeviceID is the device ID of the router for the
// GL.iNet DDNS service (the first part of its DDNS name, also SystemInfo.DDNS). It is a secret and
// read only.
type DDNSConfig struct {
	Enabled  bool   `json:"enable_ddns"`
	DeviceID string `json:"device_id"`
}

// DDNSStatus is the result of ddns get_status.
//
// Address is an IPv4 address (a public one on the test router). The router adds it about 5 s
// after DDNS is enabled and removes it about 2 s after DDNS is disabled. It is a secret.
// The meaning of Status is not known: it was 2 on the test router with DDNS enabled and disabled.
// IPs has the addresses of the WAN interfaces.
type DDNSStatus struct {
	Address string             `json:"ddns,omitempty"`
	Status  int                `json:"status"`
	IPs     []DDNSInterfaceIPs `json:"ips"`
}

// DDNSInterfaceIPs is the addresses of one interface in DDNSStatus, for example "wwan".
type DDNSInterfaceIPs struct {
	Interface string   `json:"interface"`
	IP        []string `json:"ip"`
}

// GetConfig returns the DDNS configuration.
func (s *DDNSService) GetConfig(ctx context.Context) (*DDNSConfig, error) {
	var res DDNSConfig
	if err := s.client.Call(ctx, "ddns", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetStatus returns the DDNS status. The router takes about 0.6 s to answer.
func (s *DDNSService) GetStatus(ctx context.Context) (*DDNSStatus, error) {
	var res DDNSStatus
	if err := s.client.Call(ctx, "ddns", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetEnabled enables or disables DDNS. When enabled, the router registers its name with the
// DDNS service of GL.iNet. The router returns after about 30 ms, and the next GetConfig and
// SystemService.GetStatus (DDNSEnabled) have the new value. DDNS does not change the firewall or
// the access over the WAN.
func (s *DDNSService) SetEnabled(ctx context.Context, enabled bool) error {
	params := struct {
		Enabled bool `json:"enable_ddns"`
	}{Enabled: enabled}
	return s.client.Call(ctx, "ddns", "set_config", params, nil)
}
