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

// RemoveAllStaticBinds removes all bindings with lan remove_static_bind {"mode": 1}, as the
// "remove all" button of the web UI does.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *LANService) RemoveAllStaticBinds(ctx context.Context) error {
	params := struct {
		Mode int `json:"mode"`
	}{Mode: 1}
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

// LANConfig is an item of the result of lan get_config_list: the configuration of the network
// Interface "lan" or "guest" and of its DHCP server. The flags are numbers: 1 is on and 0 is off.
//
//   - IP and Netmask are the address of the router in the network.
//   - APIsolate 1 stops the Wi-Fi clients of the network from reaching each other.
//   - WANIsolate is only for the "guest" network. The web UI of firmware 4.8.3 shows it with the
//     label key lanip.block_wan_label ("block WAN") and, while it is 1, a tip with the networks of
//     lan get_wan_info that the guest network cannot reach. The router does not give it when it
//     is not set (firmware 4.8.3); then the web UI shows 0.
//   - Enable 1 enables the DHCP server. Start and End are the first and the last address that it
//     gives. LeaseTime is a duration with a unit, for example "12h" on the test router; the web
//     UI sends minutes, for example "720m".
//   - Gateway and DNS are the gateway and the DNS servers that the DHCP server gives; empty means
//     the router. DNS has at most two items in the web UI.
//   - LPR is a list of LPR servers: the web UI shows it in the DHCP server settings with the label
//     key lan_ip.lpr_serve_label, and allows at most 5 IP addresses, each one time. What the
//     router does with it was not verified.
//   - TransferEnable is only in the "guest" item (0 on the test router). The web UI does not
//     show or send it; its meaning is not known.
type LANConfig struct {
	Interface      string   `json:"interface"`
	IP             string   `json:"ip"`
	Netmask        string   `json:"netmask"`
	APIsolate      int      `json:"ap_isolate"`
	WANIsolate     int      `json:"wan_isolate,omitempty"`
	Enable         int      `json:"enable"`
	Start          string   `json:"start"`
	End            string   `json:"end"`
	LeaseTime      string   `json:"leasetime"`
	Gateway        string   `json:"gateway"`
	DNS            []string `json:"dns"`
	LPR            []string `json:"lpr"`
	TransferEnable *int     `json:"transfer_enable,omitempty"`
}

// LANAddress is the address part of the params of lan set_config, as the web UI sends it.
// See LANConfig for the fields. SetAddress sends WANIsolate only for the "guest" network.
type LANAddress struct {
	Interface  string
	IP         string
	Netmask    string
	APIsolate  int
	WANIsolate int
}

// LANDHCP is the DHCP part of the params of lan set_config. See LANConfig for the fields.
type LANDHCP struct {
	Interface string
	Enable    int
	Start     string
	End       string
	LeaseTime string
	Gateway   string
	DNS       []string
	LPR       []string
}

// ListConfigs returns the configuration of the networks "lan" and "guest".
func (s *LANService) ListConfigs(ctx context.Context) ([]LANConfig, error) {
	var res struct {
		Interfaces []LANConfig `json:"interfaces"`
	}
	if err := s.client.Call(ctx, "lan", "get_config_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Interfaces, nil
}

// GetConfig returns the configuration of the network iface, "lan" or "guest". It returns nil and
// no error when there is no such network.
func (s *LANService) GetConfig(ctx context.Context, iface string) (*LANConfig, error) {
	configs, err := s.ListConfigs(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range configs {
		if c.Interface == iface {
			return &c, nil
		}
	}

	return nil, nil
}

// SetAddress sets the address and the AP isolation of a network with one lan set_config, as the
// web UI does. A change of the "lan" IP cuts the connection to the router: the web UI goes to
// the new IP 15 s after the write. The web UI shows the errors -7 (the IP conflicts with the
// uplink), -8 (with the LAN) and -9 (with the guest network).
//
// Not verified on a router: the parameters come from the web UI code, and wan_isolate from the
// web UI code of firmware 4.8.3. A write of the DHCP server of the guest network restarted the
// Wi-Fi of the test router (see SetDHCP); whether an address write does too is not known.
func (s *LANService) SetAddress(ctx context.Context, address LANAddress) error {
	params := map[string]any{
		"interface":  address.Interface,
		"ip":         address.IP,
		"netmask":    address.Netmask,
		"ap_isolate": address.APIsolate,
	}
	if address.Interface == "guest" {
		params["wan_isolate"] = address.WANIsolate
	}
	return s.client.Call(ctx, "lan", "set_config", params, nil)
}

// SetDHCP sets the DHCP server of a network with one lan set_config, as the web UI of firmware
// 4.8.3 does: with Enable 0 it sends only interface and enable; else it sends all fields, with an
// empty DNS or LPR as []. So an empty LPR removes all LPR servers. The web UI always has at least
// one LPR input, so with Enable 1 it always sends lpr, without the empty inputs. Firmware 4.7.0
// sent the empty inputs, for example [""]. The web UI waits until the router answers again, for at
// least 10 s.
//
// Verified on the test router only for {"interface": "guest", "enable": 0} while the guest DHCP
// server was already off: the router returned the result [] (no api_extra) within 0.1 s and kept
// all fields of both networks. About 9 s later the router restarted the Wi-Fi: radio0 left the
// channel of the repeater uplink for its own channel, and the uplink on radio0 dropped and had
// not connected again by itself 16 min later. So a write to the guest network, also one that
// changes nothing, can cut an uplink over Wi-Fi. The other parameters come from the web UI code.
func (s *LANService) SetDHCP(ctx context.Context, dhcp LANDHCP) error {
	params := map[string]any{"interface": dhcp.Interface, "enable": dhcp.Enable}
	if dhcp.Enable != 0 {
		params["start"] = dhcp.Start
		params["end"] = dhcp.End
		params["leasetime"] = dhcp.LeaseTime
		params["gateway"] = dhcp.Gateway
		params["dns"] = append([]string{}, dhcp.DNS...)
		params["lpr"] = append([]string{}, dhcp.LPR...)
	}
	return s.client.Call(ctx, "lan", "set_config", params, nil)
}
