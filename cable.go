package glinet

import "context"

// CableService calls the cable module.
type CableService service

// WANPort is a wired WAN port. The cable methods send it as "secondwan": false for FirstWANPort
// and true for SecondWANPort, as the web UI does. Only a router with
// SoftwareFeature.SecondWAN true in system get_info has a second WAN port; the web UI uses
// SecondWANPort only on such a router.
type WANPort int

const (
	FirstWANPort WANPort = iota
	SecondWANPort
)

type wanPortParams struct {
	SecondWAN bool `json:"secondwan"`
}

func (p WANPort) params() wanPortParams {
	return wanPortParams{SecondWAN: p == SecondWANPort}
}

// CableStatus is the result of cable get_status: the state of the first wired WAN port, and in
// SecondWAN the state of the second wired WAN port. The values are runtime values.
//
// The 2023 API reference gives Status 0 as failed, 1 as connected, 2 as connecting and 3 as not
// connected physically. The test router, without a WAN cable, gave Mode 0 and Status 3.
type CableStatus struct {
	Mode      int             `json:"mode"`
	Status    int             `json:"status"`
	SecondWAN SecondWANStatus `json:"secondwan"`
}

// SecondWANStatus is the state of the second wired WAN port. A router without one gives ErrCode
// -4 and ErrMsg "Unsupported features!" in a result without an error, so GetStatus returns no
// error. On a router with a second WAN port, the web UI of firmware 4.8.3 reads Mode and Status
// as for the first port; this was not seen on a router.
type SecondWANStatus struct {
	Mode    int    `json:"mode,omitempty"`
	Status  int    `json:"status,omitempty"`
	ErrCode int    `json:"err_code,omitempty"`
	ErrMsg  string `json:"err_msg,omitempty"`
}

// GetStatus returns the state of the wired WAN ports.
func (s *CableService) GetStatus(ctx context.Context) (*CableStatus, error) {
	var res CableStatus
	if err := s.client.Call(ctx, "cable", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// CableConfig is the configuration of the wired WAN port: the result of cable get_config and the
// params of cable set_config.
//
// Protocol is "dhcp", "static" or "pppoe". The test router, with Protocol "dhcp", gave only
// Protocol. The other fields come from the web UI code and are not verified on a router:
//   - IPv4 is the static address, for Protocol "static". IPv6 is an optional static IPv6 address,
//     which the web UI sends only while IPv6 is enabled.
//   - Username and Password are the PPPoE account, for Protocol "pppoe". Password is a secret,
//     and get_config returns it.
//   - VLANID (1-4094), TTL and TTLIPv6 (1-255) and MTU (576-9200) are optional; 0 means not set.
type CableConfig struct {
	Protocol string     `json:"protocol"`
	IPv4     *CableIPv4 `json:"ipv4,omitempty"`
	IPv6     *CableIPv6 `json:"ipv6,omitempty"`
	Username string     `json:"username,omitempty"`
	Password string     `json:"password,omitempty"`
	VLANID   int        `json:"vlanid,omitempty"`
	TTL      int        `json:"ttl,omitempty"`
	TTLIPv6  int        `json:"ttl_ipv6,omitempty"`
	MTU      int        `json:"mtu,omitempty"`
}

// CableIPv4 is the static IPv4 address of the wired WAN port. DNS has one or two servers.
type CableIPv4 struct {
	IP      string   `json:"ip"`
	Netmask string   `json:"netmask"`
	Gateway string   `json:"gateway"`
	DNS     []string `json:"dns"`
}

// CableIPv6 is the static IPv6 address of the wired WAN port. IP has a prefix length, for example
// "2001:db8::2/64". DNS has one or two servers.
type CableIPv6 struct {
	IP      string   `json:"ip"`
	Gateway string   `json:"gateway"`
	DNS     []string `json:"dns"`
}

// GetConfig returns the configuration of the wired WAN port port, with
// {"secondwan": bool} as the web UI sends it. The fixtures of the test router are reads with {};
// a read with "secondwan" was not tested on a router.
func (s *CableService) GetConfig(ctx context.Context, port WANPort) (*CableConfig, error) {
	var res CableConfig
	if err := s.client.Call(ctx, "cable", "get_config", port.params(), &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the configuration of the wired WAN port port. It sends config with
// "secondwan": bool, as the web UI does. The web UI sends the full configuration of the
// protocol, only the optional fields that are set, and waits 4 s after the write. For Protocol
// "pppoe" it always sends Username and Password, also when they are empty, and for the other
// protocols it does not send them. An error -1 is a VLAN ID conflict.
//
// Not verified on a router: the parameters come from the web UI code. A wrong configuration cuts
// the wired uplink.
func (s *CableService) SetConfig(ctx context.Context, port WANPort, config CableConfig) error {
	params := struct {
		CableConfig
		Username *string `json:"username,omitempty"`
		Password *string `json:"password,omitempty"`
		wanPortParams
	}{CableConfig: config, wanPortParams: port.params()}
	if config.Protocol == "pppoe" {
		params.Username = &config.Username
		params.Password = &config.Password
	}
	return s.client.Call(ctx, "cable", "set_config", params, nil)
}
