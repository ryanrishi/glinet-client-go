package glinet

import "context"

// TetheringService calls the tethering module: the USB tethering uplink of a phone.
//
// The read get_status is verified on a router (firmware 4.7.0 and 4.8.3; no phone was
// connected). The other calls come from the web UI code of firmware 4.8.3 and were not tested on
// a router.
type TetheringService service

// TetheringStatus is the result of tethering get_status. The values are runtime values.
//
// Status is 0 (not connected), 1 (connected) or 2 (connecting), from the web UI code. The test
// router, with no phone, gave Status 2 and no devices. Devices is the list of phones on USB.
// IPv4 is the address of the uplink, from the web UI code. The web UI also reads ipv6, which this
// client does not decode, because its shape is not known.
type TetheringStatus struct {
	Status  int               `json:"status"`
	Devices []TetheringDevice `json:"devices"`
	IPv4    *UplinkIPv4       `json:"ipv4,omitempty"`
}

// TetheringDevice is a phone on USB, from the web UI code. Device is the name of its network
// device. Type is 0 for Android and 1 for iOS; the web UI also shows a device without a known
// type. Use is true for the device that the uplink uses.
type TetheringDevice struct {
	Device string `json:"device"`
	Type   *int   `json:"type,omitempty"`
	Use    bool   `json:"use"`
}

// TetheringConfig is the result of tethering get_config and the params of tethering
// set_connect, from the web UI code of firmware 4.8.3. Device is the device of a
// TetheringDevice. TTL and TTLIPv6 (1-255) and MTU (576-9200) are optional; 0 means not set, and
// the web UI sends them only when they are set.
type TetheringConfig struct {
	Device  string `json:"device"`
	TTL     int    `json:"ttl,omitempty"`
	TTLIPv6 int    `json:"ttl_ipv6,omitempty"`
	MTU     int    `json:"mtu,omitempty"`
}

// GetStatus returns the tethering status.
func (s *TetheringService) GetStatus(ctx context.Context) (*TetheringStatus, error) {
	var res TetheringStatus
	if err := s.client.Call(ctx, "tethering", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetConfig returns the tethering configuration.
//
// From the web UI code of firmware 4.8.3; get_config was not read on a router.
func (s *TetheringService) GetConfig(ctx context.Context) (*TetheringConfig, error) {
	var res TetheringConfig
	if err := s.client.Call(ctx, "tethering", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Connect connects the tethering uplink with config, with tethering set_connect. The web UI
// enables the connect button only when a phone is on USB, and then reads the status again
// after 4 s.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not known: whether the router
// keeps config when the connection fails, and what it does when no phone is on USB.
func (s *TetheringService) Connect(ctx context.Context, config TetheringConfig) error {
	return s.client.Call(ctx, "tethering", "set_connect", config, nil)
}

// Disconnect disconnects the tethering uplink, with tethering disconnect {}.
//
// From the web UI code of firmware 4.8.3, not tested on a router.
func (s *TetheringService) Disconnect(ctx context.Context) error {
	return s.client.Call(ctx, "tethering", "disconnect", nil, nil)
}
