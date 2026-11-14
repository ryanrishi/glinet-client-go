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
	Week        []int  `json:"week,omitempty"`
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
