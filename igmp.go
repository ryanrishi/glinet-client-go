package glinet

import "context"

// IGMPService calls the igmp module: IGMP snooping (web UI view igmp, menu "IGMP Snooping"). The
// web UI shows the view only in the router network mode.
type IGMPService service

// IGMPConfig is the result of igmp get_config and the params of igmp set_config. Version is the
// IGMP version: the web UI offers 1, 2 and 3. The test router gave Enabled false and Version 3.
type IGMPConfig struct {
	Enabled bool `json:"enable"`
	Version int  `json:"version"`
}

// GetConfig returns the IGMP snooping configuration.
func (s *IGMPService) GetConfig(ctx context.Context) (*IGMPConfig, error) {
	var res IGMPConfig
	if err := s.client.Call(ctx, "igmp", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the IGMP snooping configuration. It sends {"enable": bool, "version": int},
// always with both values, as the web UI does.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI waits 15 s after the
// write before it reads the configuration again. What the router restarts for the write is not
// known.
func (s *IGMPService) SetConfig(ctx context.Context, config IGMPConfig) error {
	return s.client.Call(ctx, "igmp", "set_config", config, nil)
}
