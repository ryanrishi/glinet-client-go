package glinet

import "context"

// ClientsService calls the clients module.
type ClientsService service

// ClientDevice is a client device that the router knows, online or offline.
//
// MAC, Name, Alias, Class, Blocked, LimitRx and LimitTx are stable configuration. The router
// gives Alias and Class only when they are set (see SetInfo). Blocked and the limits are settings
// of the router for this client.
//
// LimitRx is the download limit and LimitTx the upload limit of QoSService.SetSpeedLimit, in
// KiB/s; 0 is no limit. The web UI shows them as KB/s with a base of 1024.
//
// The other fields are runtime values. Type is 0 for Iface "2.4G", 1 for "5G" and 2 for "cable"
// on the test router. OnlineTime and LastUpdateRate are Unix times. Rx and Tx are the current
// rates and TotalRx and TotalTx the byte counts, as the router gives them. IPv6 is the IPv6
// addresses of the client; the web UI shows each item as an address. LastRx and LastTx are lists
// of 60 numbers on the test router (firmware 4.7.0 and 4.8.3); the web UI does not use them, and
// their meaning is not known.
type ClientDevice struct {
	MAC     string `json:"mac"`
	Name    string `json:"name"`
	Alias   string `json:"alias,omitempty"`
	Class   string `json:"class,omitempty"`
	Blocked bool   `json:"blocked"`
	LimitRx int64  `json:"limit_rx"`
	LimitTx int64  `json:"limit_tx"`

	IP             string   `json:"ip"`
	IPv6           []string `json:"ipv6"`
	Iface          string   `json:"iface"`
	Type           int      `json:"type"`
	Online         bool     `json:"online"`
	OnlineTime     int64    `json:"online_time"`
	Remote         bool     `json:"remote"`
	Rx             int64    `json:"rx"`
	Tx             int64    `json:"tx"`
	LastRx         []int64  `json:"last_rx"`
	LastTx         []int64  `json:"last_tx"`
	TotalRx        int64    `json:"total_rx"`
	TotalTx        int64    `json:"total_tx"`
	TotalRxInit    int64    `json:"total_rx_init"`
	TotalTxInit    int64    `json:"total_tx_init"`
	LastUpdateRate int64    `json:"last_update_rate"`
}

// ListClients returns the client devices that the router knows.
func (s *ClientsService) ListClients(ctx context.Context) ([]ClientDevice, error) {
	var res struct {
		Clients []ClientDevice `json:"clients"`
	}
	if err := s.client.Call(ctx, "clients", "get_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Clients, nil
}

// ClientInfo is the name and the icon that the user gives a client device, as the edit dialog of
// the web UI sends them with clients set_info.
//
// MAC is the MAC of the client as ListClients gives it. Alias is the name; "" removes it, and
// the web UI then shows the Name that the router found. The web UI accepts at most 32 characters
// and no "," or '"'. Class is the icon; "" removes it. The web UI offers the classes "computer",
// "phone", "pad", "camera", "watch", "laptop", "printer", "sound", "television",
// "smartappliances", "games", "gateway", "nas", "server" and "switch".
type ClientInfo struct {
	MAC   string `json:"mac"`
	Alias string `json:"alias"`
	Class string `json:"class"`
}

// SetInfo sets the name and the icon of a client device. The web UI always sends all three
// fields, also when Alias or Class is "".
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not settled: whether the
// router removes Alias and Class for "" or keeps an empty value, whether it accepts a MAC that
// is not in ListClients, and whether it compares MACs with their letter case.
func (s *ClientsService) SetInfo(ctx context.Context, info ClientInfo) error {
	return s.client.Call(ctx, "clients", "set_info", info, nil)
}
