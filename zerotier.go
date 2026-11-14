package glinet

import (
	"context"
	"encoding/json"
	"fmt"
)

// ZeroTierService calls the zerotier module.
type ZeroTierService service

// ZeroTierConfig is the ZeroTier configuration. ID is the ZeroTier network ID to join, 16 hex
// digits in the web UI; treat it as a secret. The router reads ID as it was written and leaves it
// out before the first ID is written. LANEnabled and WANEnabled let the ZeroTier network reach the
// LAN and the WAN of the router.
type ZeroTierConfig struct {
	Enabled    bool   `json:"enabled"`
	ID         string `json:"id"`
	LANEnabled bool   `json:"lan_enabled"`
	WANEnabled bool   `json:"wan_enabled"`
}

// ZeroTierStatus is the result of zerotier get_status.
//
// Status is nil while ZeroTier is disabled and for about 1 s after it is enabled. Then it was 1
// ("connecting") on the test router, with a network ID that does not exist, for as long as
// ZeroTier was enabled (about 1 minute). The other values are from the web UI code and not
// verified: 0 connected, 2 not authorized, 3 network not found.
type ZeroTierStatus struct {
	Status *int `json:"status,omitempty"`
}

// GetConfig returns the ZeroTier configuration.
func (s *ZeroTierService) GetConfig(ctx context.Context) (*ZeroTierConfig, error) {
	var res ZeroTierConfig
	if err := s.client.Call(ctx, "zerotier", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetStatus returns the ZeroTier status.
func (s *ZeroTierService) GetStatus(ctx context.Context) (*ZeroTierStatus, error) {
	var raw json.RawMessage
	if err := s.client.Call(ctx, "zerotier", "get_status", nil, &raw); err != nil {
		return nil, err
	}

	var res ZeroTierStatus
	if string(raw) == "[]" {
		return &res, nil
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("glinet: zerotier get_status: decode result: %w", err)
	}

	return &res, nil
}

// SetConfig sets the ZeroTier configuration.
//
// With config.Enabled true, the router sets ID, LANEnabled and WANEnabled to the sent values,
// starts ZeroTier and joins the network ID. It returns after about 2.5 s (11 s for the first
// join on the test router), and the next GetConfig has the values. An empty ID gives an *RPCError
// with code 1 ("parameter error") and changes nothing.
//
// With config.Enabled false, the router only disables ZeroTier, in about 1-2 s. It keeps ID,
// LANEnabled and WANEnabled and ignores the sent values. So after the first write of an ID, the
// router always has an ID: no write removes it.
//
// Observed on a test router with a network ID that does not exist: the firewall zones, the
// firewall rules and the DNS servers did not change, also with LANEnabled true. Not verified:
// the firewall of a router that is connected to a ZeroTier network.
func (s *ZeroTierService) SetConfig(ctx context.Context, config ZeroTierConfig) error {
	return s.client.Call(ctx, "zerotier", "set_config", config, nil)
}
