package glinet

import "context"

// AdGuardHomeService calls the adguardhome module.
type AdGuardHomeService service

// AdGuardHomeConfig is the result of adguardhome get_config and the params of adguardhome
// set_config.
type AdGuardHomeConfig struct {
	Enabled    bool `json:"enabled"`
	DNSEnabled bool `json:"dns_enabled"`
}

// GetConfig returns the AdGuard Home configuration.
func (s *AdGuardHomeService) GetConfig(ctx context.Context) (*AdGuardHomeConfig, error) {
	var res AdGuardHomeConfig
	if err := s.client.Call(ctx, "adguardhome", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the AdGuard Home configuration. DNSEnabled makes AdGuard Home answer the
// DNS queries of the clients. The router returns before AdGuard Home has started: in the test,
// the DNS server of the router did not answer for up to 16 s after a write that enabled it.
func (s *AdGuardHomeService) SetConfig(ctx context.Context, config AdGuardHomeConfig) error {
	return s.client.Call(ctx, "adguardhome", "set_config", config, nil)
}
