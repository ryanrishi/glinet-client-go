package glinet

import "context"

type AdGuardService service

type GetAdGuardConfigResponse struct {
	Enabled    bool `json:"enabled"`
	DNSEnabled bool `json:"dns_enabled"`
}

type SetAdGuardConfigRequest struct {
	Enabled    bool `json:"enabled"`
	DNSEnabled bool `json:"dns_enabled"`
}

func (s *AdGuardService) GetAdGuardConfig(ctx context.Context) (*GetAdGuardConfigResponse, error) {
	var res GetAdGuardConfigResponse
	if err := s.client.Call(ctx, "adguardhome", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *AdGuardService) SetAdGuardConfig(ctx context.Context, config *SetAdGuardConfigRequest) error {
	return s.client.Call(ctx, "adguardhome", "set_config", config, nil)
}
