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

// GetConfig returns the DDNS configuration.
func (s *DDNSService) GetConfig(ctx context.Context) (*DDNSConfig, error) {
	var res DDNSConfig
	if err := s.client.Call(ctx, "ddns", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetEnabled enables or disables DDNS. When enabled, the router registers its name with the
// DDNS service of GL.iNet.
//
// Not verified on a router: the params {"enable_ddns": enabled} are those of the web UI, which
// reads the configuration again 3 s after the write.
func (s *DDNSService) SetEnabled(ctx context.Context, enabled bool) error {
	params := struct {
		Enabled bool `json:"enable_ddns"`
	}{Enabled: enabled}
	return s.client.Call(ctx, "ddns", "set_config", params, nil)
}
