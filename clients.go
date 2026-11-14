package glinet

import "context"

// ClientsService calls the clients module.
type ClientsService service

// ClientDevice is a client device that the router knows, online or offline.
//
// MAC, Name, Alias, Class, Blocked, LimitRx and LimitTx are stable configuration. The router
// gives Alias and Class only when they are set. Blocked and the limits are settings of the
// router for this client.
//
// The other fields are runtime values. Type is 0 for Iface "2.4G", 1 for "5G" and 2 for "cable"
// on the test router. OnlineTime and LastUpdateRate are Unix times. Rx and Tx are the current
// rates and TotalRx and TotalTx the byte counts, as the router gives them.
type ClientDevice struct {
	MAC     string `json:"mac"`
	Name    string `json:"name"`
	Alias   string `json:"alias,omitempty"`
	Class   string `json:"class,omitempty"`
	Blocked bool   `json:"blocked"`
	LimitRx int64  `json:"limit_rx"`
	LimitTx int64  `json:"limit_tx"`

	IP             string `json:"ip"`
	Iface          string `json:"iface"`
	Type           int    `json:"type"`
	Online         bool   `json:"online"`
	OnlineTime     int64  `json:"online_time"`
	Remote         bool   `json:"remote"`
	Rx             int64  `json:"rx"`
	Tx             int64  `json:"tx"`
	TotalRx        int64  `json:"total_rx"`
	TotalTx        int64  `json:"total_tx"`
	TotalRxInit    int64  `json:"total_rx_init"`
	TotalTxInit    int64  `json:"total_tx_init"`
	LastUpdateRate int64  `json:"last_update_rate"`
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
