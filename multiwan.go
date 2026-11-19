package glinet

import "context"

// MultiWANService calls the kmwan module: the multi-WAN of the router. In failover mode the router
// uses the online WAN interface with the lowest metric. In load balance mode it divides the
// traffic between the online WAN interfaces by their weights. The router finds out if an
// interface is online by tracking: it sends requests to the track hosts of the interface.
//
// The reads were called on a router (firmware 4.7.0 and 4.8.3). The writes and GetSensitivity come
// from the web UI code of firmware 4.8.3 (view multiwan) and were not tested on a router.
//
// A write can change which WAN interface the router uses, and so can stop the Internet access of
// the router and of its clients. After SetFailover, SetLoadBalance and SetInterface, the web UI
// waits 10 s and then reads the configuration and the system status again. After SetSensitivity
// it waits 1 s.
//
// The web UI shows only the interfaces of GetConfig that are not in SystemInfo.HiddenFeatures.
type MultiWANService service

// The modes of the multi-WAN.
const (
	MultiWANFailover    = 0
	MultiWANLoadBalance = 1
)

// The track modes of an interface. The web UI labels them "normal", "low data" (badge "passive")
// and "strict". The label texts and the router code are not available, so how the modes differ is
// not known.
const (
	MultiWANTrackLowData = 0
	MultiWANTrackNormal  = 1
	MultiWANTrackStrict  = 2
)

// The track protocols of an interface.
const (
	MultiWANTrackIPv4 = 0
	MultiWANTrackIPv6 = 1
	MultiWANTrackBoth = 2
)

// MultiWANTrackPing is the track method ping, the only track method in the web UI.
const MultiWANTrackPing = 0

// The values of MultiWANInterfaceStatus.StatusV4 and StatusV6, as the web UI shows them.
const (
	MultiWANOnline  = 0
	MultiWANOffline = 1
	MultiWANError   = 2
)

// The sensitivity levels of the tracking. With MultiWANSensitivityCustom, the web UI lets the
// user set MultiWANSensitivity.Interval.
const (
	MultiWANSensitivityLow    = "low"
	MultiWANSensitivityMedium = "medium"
	MultiWANSensitivityHigh   = "high"
	MultiWANSensitivityCustom = "custom"
)

// MultiWANConfig is the result of kmwan get_config. Mode is MultiWANFailover or
// MultiWANLoadBalance.
type MultiWANConfig struct {
	Mode       int                 `json:"mode"`
	Interfaces []MultiWANInterface `json:"interfaces"`
}

// MultiWANInterface is one WAN interface in MultiWANConfig, for example "wan", "secondwan",
// "wwan" (repeater), "tethering" or a modem interface such as "modem_1_1_2".
//
// Metric is the priority in failover mode: a lower metric comes first. The router had 10, 20,
// 30 and 40; SetFailover sends 1, 2, 3 and so on. Weight is the weight in load balance mode,
// 0 to 10 in the web UI.
type MultiWANInterface struct {
	MultiWANTracking
	Metric int `json:"metric"`
	Weight int `json:"weight"`
}

// MultiWANTracking is the tracking of one WAN interface.
//
// When EnableCheck is false, the router uses only the physical state of the interface. TrackMode
// is a MultiWANTrack mode constant (MultiWANTrackLowData, MultiWANTrackNormal or
// MultiWANTrackStrict), TrackProto is MultiWANTrackIPv4, MultiWANTrackIPv6 or MultiWANTrackBoth,
// and TrackMethod is MultiWANTrackPing. TrackIPv4 and TrackIPv6 are the track hosts (IP
// addresses). EnableSSL applies only to TrackMethod 1; the web UI does not list method 1, so
// what it is is not known.
type MultiWANTracking struct {
	Interface   string   `json:"interface"`
	EnableCheck bool     `json:"enable_check"`
	TrackMode   int      `json:"track_mode"`
	TrackProto  int      `json:"track_proto"`
	TrackMethod int      `json:"track_method"`
	TrackIPv4   []string `json:"track_ipv4"`
	TrackIPv6   []string `json:"track_ipv6"`
	EnableSSL   bool     `json:"enable_ssl"`
}

// MultiWANStatus is the result of kmwan get_status.
type MultiWANStatus struct {
	Interfaces []MultiWANInterfaceStatus `json:"interfaces"`
}

// MultiWANInterfaceStatus is the state of one WAN interface in MultiWANStatus. StatusV4 and
// StatusV6 are MultiWANOnline, MultiWANOffline or MultiWANError. The web UI uses StatusV4 for
// track protocol IPv4, StatusV6 for IPv6, and both for MultiWANTrackBoth or when EnableCheck is
// false.
type MultiWANInterfaceStatus struct {
	Interface string `json:"interface"`
	StatusV4  int    `json:"status_v4"`
	StatusV6  int    `json:"status_v6"`
}

// MultiWANWeight is the weight of one WAN interface for SetLoadBalance.
type MultiWANWeight struct {
	Interface string `json:"interface"`
	Weight    int    `json:"weight"`
}

// MultiWANSensitivity is the sensitivity of the tracking. Level is a MultiWANSensitivity level
// constant. Interval is a time in seconds: the web UI lets the user set 0.5 or a whole number
// from 1 to 90, only for MultiWANSensitivityCustom, and sends the Interval of GetSensitivity with
// the other levels.
type MultiWANSensitivity struct {
	Level    string  `json:"level"`
	Interval float64 `json:"val"`
}

type multiWANMetric struct {
	Interface string `json:"interface"`
	Metric    int    `json:"metric"`
}

type multiWANInterfaceRequest struct {
	Interface   string    `json:"interface"`
	EnableCheck bool      `json:"enable_check"`
	TrackProto  *int      `json:"track_proto,omitempty"`
	TrackMethod *int      `json:"track_method,omitempty"`
	TrackMode   *int      `json:"track_mode,omitempty"`
	TrackIPv4   *[]string `json:"track_ipv4,omitempty"`
	TrackIPv6   *[]string `json:"track_ipv6,omitempty"`
	EnableSSL   *bool     `json:"enable_ssl,omitempty"`
}

// GetConfig returns the multi-WAN configuration.
func (s *MultiWANService) GetConfig(ctx context.Context) (*MultiWANConfig, error) {
	var res MultiWANConfig
	if err := s.client.Call(ctx, "kmwan", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetStatus returns the state of the WAN interfaces. The web UI reads it again each 10 s.
func (s *MultiWANService) GetStatus(ctx context.Context) (*MultiWANStatus, error) {
	var res MultiWANStatus
	if err := s.client.Call(ctx, "kmwan", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetSensitivity returns the sensitivity of the tracking, from kmwan get_sensitivity.
//
// Not called on a router: the result shape {"sensitivity": {"level", "val"}} is from the web UI
// code of firmware 4.8.3. When the result has no sensitivity, GetSensitivity returns the zero
// value; the web UI then shows level "medium" and 5 s.
func (s *MultiWANService) GetSensitivity(ctx context.Context) (*MultiWANSensitivity, error) {
	var res struct {
		Sensitivity MultiWANSensitivity `json:"sensitivity"`
	}
	if err := s.client.Call(ctx, "kmwan", "get_sensitivity", nil, &res); err != nil {
		return nil, err
	}

	return &res.Sensitivity, nil
}

// SetFailover sets failover mode with the priority order interfaces, with kmwan set_config
// {"mode": 0, "interfaces": [{"interface", "metric"}]}. The first interface gets metric 1, the
// next metric 2, and so on, as the web UI does.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI sends all interfaces
// of GetConfig: first the interfaces that it shows, in the order of the user, then the hidden
// interfaces (SystemInfo.HiddenFeatures). It does not send the weights. Not known: what the router
// does with an interface that is not in the list, or with an unknown interface.
func (s *MultiWANService) SetFailover(ctx context.Context, interfaces []string) error {
	metrics := make([]multiWANMetric, len(interfaces))
	for i, name := range interfaces {
		metrics[i] = multiWANMetric{Interface: name, Metric: i + 1}
	}
	params := struct {
		Mode       int              `json:"mode"`
		Interfaces []multiWANMetric `json:"interfaces"`
	}{Mode: MultiWANFailover, Interfaces: metrics}
	return s.client.Call(ctx, "kmwan", "set_config", params, nil)
}

// SetLoadBalance sets load balance mode with the weights, with kmwan set_config
// {"mode": 1, "interfaces": [{"interface", "weight"}]}.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI sends the interfaces
// that it shows (not the hidden interfaces, SystemInfo.HiddenFeatures), each with a weight from 0
// to 10. It does not send the metrics. When fewer than 2 interfaces have a weight that is not 0,
// the web UI shows a caution and sends the write only when the user accepts it. Not known: what
// weight 0 does, and what the router does with an interface that is not in the list.
func (s *MultiWANService) SetLoadBalance(ctx context.Context, weights []MultiWANWeight) error {
	params := struct {
		Mode       int              `json:"mode"`
		Interfaces []MultiWANWeight `json:"interfaces"`
	}{Mode: MultiWANLoadBalance, Interfaces: append([]MultiWANWeight{}, weights...)}
	return s.client.Call(ctx, "kmwan", "set_config", params, nil)
}

// SetInterface sets the tracking of one WAN interface with kmwan set_interface. It sends the
// fields as the web UI does:
//
//   - always interface and enable_check;
//   - when EnableCheck is true, also track_proto, track_method and track_mode, track_ipv4 when
//     TrackProto is MultiWANTrackIPv4 or MultiWANTrackBoth, track_ipv6 when TrackProto is
//     MultiWANTrackIPv6 or MultiWANTrackBoth, and enable_ssl when TrackMethod is 1.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI sends 2 to 4 hosts for
// each protocol that it sends: IPv4 addresses in track_ipv4 and IPv6 addresses in track_ipv6. It
// lets the user select the track protocol and set IPv6 hosts only when IPv6 is enabled
// (SystemStatus.System.Ipv6Enabled). Not known: if the router keeps the fields that the write does
// not send, and if it checks the hosts. Hosts that do not answer make the router think that the
// interface is offline, so it can stop using it.
func (s *MultiWANService) SetInterface(ctx context.Context, tracking MultiWANTracking) error {
	params := multiWANInterfaceRequest{Interface: tracking.Interface, EnableCheck: tracking.EnableCheck}
	if tracking.EnableCheck {
		params.TrackProto = &tracking.TrackProto
		params.TrackMethod = &tracking.TrackMethod
		params.TrackMode = &tracking.TrackMode
		if tracking.TrackProto == MultiWANTrackIPv4 || tracking.TrackProto == MultiWANTrackBoth {
			hosts := append([]string{}, tracking.TrackIPv4...)
			params.TrackIPv4 = &hosts
		}
		if tracking.TrackProto == MultiWANTrackIPv6 || tracking.TrackProto == MultiWANTrackBoth {
			hosts := append([]string{}, tracking.TrackIPv6...)
			params.TrackIPv6 = &hosts
		}
		if tracking.TrackMethod == 1 {
			params.EnableSSL = &tracking.EnableSSL
		}
	}
	return s.client.Call(ctx, "kmwan", "set_interface", params, nil)
}

// SetSensitivity sets the sensitivity of the tracking for all interfaces, with kmwan
// set_sensitivity {"sensitivity": {"level", "val"}}.
//
// From the web UI code of firmware 4.8.3, not tested on a router. Not known: what Interval does
// with a level other than MultiWANSensitivityCustom.
func (s *MultiWANService) SetSensitivity(ctx context.Context, sensitivity MultiWANSensitivity) error {
	params := struct {
		Sensitivity MultiWANSensitivity `json:"sensitivity"`
	}{Sensitivity: sensitivity}
	return s.client.Call(ctx, "kmwan", "set_sensitivity", params, nil)
}
