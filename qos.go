package glinet

import "context"

// QoSService calls the qos module: the speed limit of a client device.
//
// ListClients gives the limits of each client (ClientDevice LimitRx and LimitTx).
//
// The web UI offers the speed limit only when the router is in mode router (system mode 0) or
// passthrough (6), and offers a new limit only when network acceleration is off (system status
// netnat_enabled false). The web UI says that network acceleration can stop the limits.
type QoSService service

// SpeedLimit is the speed limit of the client device with MAC, in KiB/s, as the web UI sends it
// with qos set_speed_limit_rule. Download is LimitRx and Upload is LimitTx of ClientDevice.
// 0 is no limit in that direction. The web UI accepts 1 to 1073741824 for a limit.
type SpeedLimit struct {
	MAC      string `json:"mac"`
	Download int64  `json:"download"`
	Upload   int64  `json:"upload"`
}

// SetSpeedLimit sets the speed limit of a client device. The web UI waits up to 30 seconds for
// the result.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not settled: what the router
// does when Download and Upload are both 0 (the web UI permits this), whether it accepts a MAC
// that is not in ListClients, and whether it compares MACs with their letter case. In the Kbps
// display mode the web UI sends the value divided by 8, which can be a fraction; SpeedLimit sends
// only whole KiB/s.
func (s *QoSService) SetSpeedLimit(ctx context.Context, limit SpeedLimit) error {
	return s.client.Call(ctx, "qos", "set_speed_limit_rule", limit, nil)
}

// RemoveSpeedLimit removes the speed limit of the client device with mac, with qos
// remove_speed_limit_rule {"mac": mac}. The web UI waits up to 30 seconds for the result. The web
// UI offers it only when the client has a limit, also while network acceleration is on.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not settled: the result for a
// MAC that has no limit.
func (s *QoSService) RemoveSpeedLimit(ctx context.Context, mac string) error {
	params := struct {
		MAC string `json:"mac"`
	}{MAC: mac}
	return s.client.Call(ctx, "qos", "remove_speed_limit_rule", params, nil)
}
