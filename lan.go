package glinet

import "context"

// LANService calls the lan module.
type LANService service

// StaticBind is a DHCP static IP binding. The router identifies a binding by its exact MAC
// string, so "AA:..." and "aa:..." are two bindings. An empty Name means the binding has no name.
type StaticBind struct {
	MAC  string `json:"mac"`
	IP   string `json:"ip"`
	Name string `json:"name,omitempty"`
}

// ListStaticBinds returns the DHCP static IP bindings.
func (s *LANService) ListStaticBinds(ctx context.Context) ([]StaticBind, error) {
	var res struct {
		StaticBindList []StaticBind `json:"static_bind_list"`
	}
	if err := s.client.Call(ctx, "lan", "get_static_bind_list", nil, &res); err != nil {
		return nil, err
	}

	return res.StaticBindList, nil
}

// FindStaticBind returns the binding whose MAC string is equal to mac, with the same letter case,
// as the router compares MACs. It returns nil and no error when there is no such binding.
func (s *LANService) FindStaticBind(ctx context.Context, mac string) (*StaticBind, error) {
	binds, err := s.ListStaticBinds(ctx)
	if err != nil {
		return nil, err
	}

	for _, b := range binds {
		if b.MAC == mac {
			return &b, nil
		}
	}

	return nil, nil
}

// AddStaticBind adds a binding. The router first removes each binding that has the same MAC or IP.
func (s *LANService) AddStaticBind(ctx context.Context, bind StaticBind) error {
	return s.client.Call(ctx, "lan", "add_static_bind", bind, nil)
}

// SetStaticBind sets the IP and name of the binding with the same MAC. An empty Name removes the name.
// The router also removes each other binding that has the same IP.
// The router changes nothing and returns no error when no binding has that MAC.
func (s *LANService) SetStaticBind(ctx context.Context, bind StaticBind) error {
	return s.client.Call(ctx, "lan", "set_static_bind", bind, nil)
}

// RemoveStaticBind removes the binding with the MAC mac.
// The router returns no error when no binding has that MAC.
func (s *LANService) RemoveStaticBind(ctx context.Context, mac string) error {
	params := struct {
		Mode int    `json:"mode"`
		MAC  string `json:"mac"`
	}{Mode: 0, MAC: mac}
	return s.client.Call(ctx, "lan", "remove_static_bind", params, nil)
}

// WANInfo is an item of the result of lan get_wan_info: the IPv4 address of an uplink interface,
// for example "wan" or "wwan". The values are runtime values.
type WANInfo struct {
	Interface string     `json:"interface"`
	Info      WANAddress `json:"info"`
}

// WANAddress is the IPv4 address and network of an uplink interface.
type WANAddress struct {
	IPAddr    string `json:"ipaddr"`
	Netmask   string `json:"netmask"`
	Prefix    int    `json:"prefix"`
	Network   string `json:"network"`
	Broadcast string `json:"broadcast"`
}

// GetWANInfo returns the IPv4 addresses of the uplink interfaces. On the test router, the result
// had only the interface that was connected.
func (s *LANService) GetWANInfo(ctx context.Context) ([]WANInfo, error) {
	var res struct {
		WANInfo []WANInfo `json:"wan_info"`
	}
	if err := s.client.Call(ctx, "lan", "get_wan_info", nil, &res); err != nil {
		return nil, err
	}

	return res.WANInfo, nil
}
