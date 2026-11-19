package glinet

import "context"

// BarkService calls the bark module: the router client of the Bark parental control service
// (bark.us), the second parental control mode (ParentalControlModeBark).
//
// The web UI shows Bark only on a router with system get_info software_feature.bark, firmware
// category "2c" and country code "us". The test router does not have the module: each read
// returns -32601 (method not found). The calls and fields are from the web UI code of firmware
// 4.8.3 and the archived API reference, and were not tested on a router.
type BarkService service

// BarkConfig is the configuration of the Bark client.
type BarkConfig struct {
	Enabled     bool `json:"enable"`
	IPv6Enabled bool `json:"ipv6_enable"`
}

// BarkStatus is the status of the Bark client. Status is 1 connecting, 2 ready or 3 online.
// PairingCode pairs the router with a Bark account; treat it as a secret.
type BarkStatus struct {
	Status      int    `json:"status"`
	PairingCode string `json:"pairing_code"`
}

// GetConfig returns the configuration.
func (s *BarkService) GetConfig(ctx context.Context) (*BarkConfig, error) {
	var res BarkConfig
	if err := s.client.Call(ctx, "bark", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the configuration with bark set_config {"enable", "ipv6_enable"}.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI sends back each field
// that get_config returned. After it enables Bark, it calls ParentalControl SetMode with
// ParentalControlModeBark when the mode was another one. It warns when the DNS override
// (dns get_config force_dns) is on while Bark is online.
func (s *BarkService) SetConfig(ctx context.Context, config BarkConfig) error {
	return s.client.Call(ctx, "bark", "set_config", config, nil)
}

// GetStatus returns the status. The web UI calls it only while Bark is enabled.
func (s *BarkService) GetStatus(ctx context.Context) (*BarkStatus, error) {
	var res BarkStatus
	if err := s.client.Call(ctx, "bark", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Logout removes the pairing of the router with the Bark account, with bark logout {}.
//
// From the web UI code of firmware 4.8.3, not tested on a router.
func (s *BarkService) Logout(ctx context.Context) error {
	return s.client.Call(ctx, "bark", "logout", nil, nil)
}
