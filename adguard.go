package glinet

import "context"

// AdGuardService calls the adguardhome module.
type AdGuardService service

// GetAdGuardConfigResponse is the result of adguardhome get_config.
type GetAdGuardConfigResponse struct {
	Enabled    bool `json:"enabled"`
	DNSEnabled bool `json:"dns_enabled"`
}

// SetAdGuardConfigRequest is the params of adguardhome set_config.
type SetAdGuardConfigRequest struct {
	Enabled    bool `json:"enabled"`
	DNSEnabled bool `json:"dns_enabled"`
}

// GetAdGuardConfig returns the AdGuard Home configuration.
func (s *AdGuardService) GetAdGuardConfig(ctx context.Context) (*GetAdGuardConfigResponse, error) {
	var res GetAdGuardConfigResponse
	if err := s.client.Call(ctx, "adguardhome", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetAdGuardConfig sets the AdGuard Home configuration.
func (s *AdGuardService) SetAdGuardConfig(ctx context.Context, config *SetAdGuardConfigRequest) error {
	return s.client.Call(ctx, "adguardhome", "set_config", config, nil)
}
