package glinet

import "context"

// FanService calls the fan module of a router with a fan (SystemInfo.HardwareFeature.Fan).
// Temperatures are in degrees Celsius.
type FanService service

// FanConfig is the result of fan get_config. The fan starts when the CPU temperature is more
// than Temperature. WarnTemperature is a value that the user cannot set in the web UI: the web
// UI shows a warning when Temperature is WarnTemperature or less.
type FanConfig struct {
	Temperature     int `json:"temperature"`
	WarnTemperature int `json:"warn_temperature"`
}

// FanStatus is the result of fan get_status. Running is true while the fan turns, and Speed is
// its speed in RPM.
type FanStatus struct {
	Speed   int  `json:"speed"`
	Running bool `json:"status"`
}

// GetConfig returns the fan configuration.
func (s *FanService) GetConfig(ctx context.Context) (*FanConfig, error) {
	var res FanConfig
	if err := s.client.Call(ctx, "fan", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetStatus returns the fan status.
func (s *FanService) GetStatus(ctx context.Context) (*FanStatus, error) {
	var res FanStatus
	if err := s.client.Call(ctx, "fan", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetTemperature sets the temperature at which the fan starts, with fan set_config
// {"temperature": temperature}. This is the only value that the web UI of firmware 4.8.3 sends.
//
// Tested on a router (GL-AXT1800, firmware 4.8.3) with 60: the router returned [], the next
// GetConfig had Temperature 60, and the fan started. Writes of 62 and 60 were also read back, and
// WarnTemperature stayed 75 in each read. The web UI offers only 70 to 90 in steps of 5, so the
// router accepts values that the web UI does not offer. The lowest value that the firmware accepts
// is not known.
func (s *FanService) SetTemperature(ctx context.Context, temperature int) error {
	params := struct {
		Temperature int `json:"temperature"`
	}{Temperature: temperature}
	return s.client.Call(ctx, "fan", "set_config", params, nil)
}
