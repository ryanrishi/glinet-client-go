package glinet

import "context"

// TimerService calls the timer module, which runs scheduled tasks.
//
// Times are the hour and the minute as strings in local time of the router. The router does not
// check them and keeps them as written; the web UI writes them with two digits, such as "07" and "00".
// Week has the days of the week to run the task, 0 (Sunday) to 6 (Saturday). The router keeps the
// order and duplicates of Week as written, and does not check the values.
type TimerService service

// LEDSchedule turns the LEDs off at TurnOffHour:TurnOffMin and on at TurnOnHour:TurnOnMin
// on the days in Week.
type LEDSchedule struct {
	Enabled     bool   `json:"enable"`
	TurnOnHour  string `json:"turnon_hour,omitempty"`
	TurnOnMin   string `json:"turnon_min,omitempty"`
	TurnOffHour string `json:"turnoff_hour,omitempty"`
	TurnOffMin  string `json:"turnoff_min,omitempty"`
	Week        []int  `json:"week,omitzero"`
}

// GetLEDSchedule returns the LED schedule.
func (s *TimerService) GetLEDSchedule(ctx context.Context) (*LEDSchedule, error) {
	var res LEDSchedule
	if err := s.client.Call(ctx, "timer", "get_led", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetLEDSchedule sets the LED schedule.
//
// When schedule.Enabled is false, the router only disables the schedule and keeps the other values.
// When it is true, the router needs all values: an empty time or Week gives the *RPCError
// -1 "parameter missing", and the router changes nothing.
func (s *TimerService) SetLEDSchedule(ctx context.Context, schedule LEDSchedule) error {
	return s.client.Call(ctx, "timer", "set_led", schedule, nil)
}

// RebootSchedule reboots the router at Hour:Min on the days in Week.
type RebootSchedule struct {
	Enabled bool   `json:"enable"`
	Hour    string `json:"hour,omitempty"`
	Min     string `json:"min,omitempty"`
	Week    []int  `json:"week,omitzero"`
}

// GetRebootSchedule returns the reboot schedule.
func (s *TimerService) GetRebootSchedule(ctx context.Context) (*RebootSchedule, error) {
	var res RebootSchedule
	if err := s.client.Call(ctx, "timer", "get_reboot", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetRebootSchedule sets the reboot schedule.
//
// When schedule.Enabled is false, the router only disables the schedule and keeps the other values.
// A write with Enabled true was not tested on a router. The web UI sends Hour, Min and Week only
// with Enabled true, as for the LED schedule.
func (s *TimerService) SetRebootSchedule(ctx context.Context, schedule RebootSchedule) error {
	return s.client.Call(ctx, "timer", "set_reboot", schedule, nil)
}

// ScreenSchedule turns the LCD screen off at TurnOffHour:TurnOffMin and on at
// TurnOnHour:TurnOnMin on the days in Week. Only a router with a screen
// (SystemInfo.HardwareFeature.Screen) has it.
type ScreenSchedule struct {
	Enabled     bool   `json:"enable"`
	TurnOnHour  string `json:"turnon_hour,omitempty"`
	TurnOnMin   string `json:"turnon_min,omitempty"`
	TurnOffHour string `json:"turnoff_hour,omitempty"`
	TurnOffMin  string `json:"turnoff_min,omitempty"`
	Week        []int  `json:"week,omitzero"`
}

// GetScreenSchedule returns the screen schedule, with timer get_screen.
//
// From the web UI code of firmware 4.8.3, not seen on a router: the result has the fields of the
// LED schedule. The test router has no screen, and the result on such a router is not known.
func (s *TimerService) GetScreenSchedule(ctx context.Context) (*ScreenSchedule, error) {
	var res ScreenSchedule
	if err := s.client.Call(ctx, "timer", "get_screen", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetScreenSchedule sets the screen schedule, with timer set_screen.
//
// From the web UI code of firmware 4.8.3, not tested on a router. The web UI sends the same
// params as for the LED schedule: enable and week always, and the times only with enable true.
// It does not let the user set the same on and off time. That the router needs all values with
// enable true, as for the LED schedule, is not known.
func (s *TimerService) SetScreenSchedule(ctx context.Context, schedule ScreenSchedule) error {
	return s.client.Call(ctx, "timer", "set_screen", schedule, nil)
}
