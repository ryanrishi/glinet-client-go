package glinet

import "context"

// ZeroTierService calls the zerotier module.
type ZeroTierService service

// ZeroTierConfig is the ZeroTier configuration. ID is the ZeroTier network ID to join, 16 hex
// digits in the web UI; treat it as a secret. The router leaves ID out when it is not set (not
// verified: no ID was set on the test router). LANEnabled and WANEnabled let the ZeroTier network
// reach the LAN and the WAN of the router.
type ZeroTierConfig struct {
	Enabled    bool   `json:"enabled"`
	ID         string `json:"id,omitempty"`
	LANEnabled bool   `json:"lan_enabled"`
	WANEnabled bool   `json:"wan_enabled"`
}

// GetConfig returns the ZeroTier configuration.
func (s *ZeroTierService) GetConfig(ctx context.Context) (*ZeroTierConfig, error) {
	var res ZeroTierConfig
	if err := s.client.Call(ctx, "zerotier", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the ZeroTier configuration. It returns after about 1 s.
//
// With config.Enabled false, the router only disables ZeroTier and keeps LANEnabled and
// WANEnabled; it ignores the sent values.
//
// Not verified on a router: a write with Enabled true, which joins the network ID and contacts
// the ZeroTier servers, and a write of ID. The web UI sends all four fields to enable it, only
// Enabled false to disable it, and reads the configuration again 5 s after the write.
func (s *ZeroTierService) SetConfig(ctx context.Context, config ZeroTierConfig) error {
	return s.client.Call(ctx, "zerotier", "set_config", config, nil)
}
