package glinettest

import (
	"encoding/json"
	"slices"
	"strconv"
	"sync"
)

// LEDSchedule is the LED schedule of the timer module.
type LEDSchedule struct {
	Enabled     bool   `json:"enable"`
	TurnOnHour  string `json:"turnon_hour"`
	TurnOnMin   string `json:"turnon_min"`
	TurnOffHour string `json:"turnoff_hour"`
	TurnOffMin  string `json:"turnoff_min"`
	Week        []int  `json:"week"`
}

// RebootSchedule is the reboot schedule of the timer module.
type RebootSchedule struct {
	Enabled bool   `json:"enable"`
	Hour    string `json:"hour"`
	Min     string `json:"min"`
	Week    []int  `json:"week"`
}

// Timer serves the LED and reboot schedules of the timer module. Like the router, a write that
// does not enable a schedule (no enable, or enable false) only disables it and keeps the other
// values. A write that enables it needs all values, or it returns the result
// {"err_code":-1,"err_msg":"parameter missing"} and changes nothing. The values are not checked:
// a time is kept as a string, also when it was written as a number, and an empty week becomes all days.
// The enabled write of the reboot schedule was not verified on the router; it works like the LED schedule.
type Timer struct {
	mu     sync.Mutex
	led    LEDSchedule
	reboot RebootSchedule
}

var (
	allDays          = []int{0, 1, 2, 3, 4, 5, 6}
	parameterMissing = map[string]any{"err_code": -1, "err_msg": "parameter missing"}
)

func newTimer() *Timer {
	return &Timer{
		led:    LEDSchedule{TurnOnHour: "07", TurnOnMin: "00", TurnOffHour: "22", TurnOffMin: "00", Week: allDays},
		reboot: RebootSchedule{Hour: "02", Min: "15", Week: []int{3, 5}},
	}
}

// LEDSchedule returns the LED schedule.
func (m *Timer) LEDSchedule() LEDSchedule {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.led
	c.Week = slices.Clone(c.Week)
	return c
}

// SetLEDSchedule sets the LED schedule.
func (m *Timer) SetLEDSchedule(c LEDSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.Week = slices.Clone(c.Week)
	m.led = c
}

// RebootSchedule returns the reboot schedule.
func (m *Timer) RebootSchedule() RebootSchedule {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := m.reboot
	c.Week = slices.Clone(c.Week)
	return c
}

// SetRebootSchedule sets the reboot schedule.
func (m *Timer) SetRebootSchedule(c RebootSchedule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.Week = slices.Clone(c.Week)
	m.reboot = c
}

func (m *Timer) register(s *Server) {
	s.Handle("timer", "get_led", func(json.RawMessage) (any, error) {
		return m.LEDSchedule(), nil
	})
	s.Handle("timer", "set_led", m.setLED)
	s.Handle("timer", "get_reboot", func(json.RawMessage) (any, error) {
		return m.RebootSchedule(), nil
	})
	s.Handle("timer", "set_reboot", m.setReboot)
}

func (m *Timer) setLED(params json.RawMessage) (any, error) {
	w, err := parseSchedule(params, "turnon_hour", "turnon_min", "turnoff_hour", "turnoff_min")
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	switch {
	case !w.enable:
		m.led.Enabled = false
	case w.times == nil || w.week == nil:
		return parameterMissing, nil
	default:
		m.led = LEDSchedule{Enabled: true, TurnOnHour: w.times[0], TurnOnMin: w.times[1], TurnOffHour: w.times[2], TurnOffMin: w.times[3], Week: w.week}
	}
	return writeResult, nil
}

func (m *Timer) setReboot(params json.RawMessage) (any, error) {
	w, err := parseSchedule(params, "hour", "min")
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	switch {
	case !w.enable:
		m.reboot.Enabled = false
	case w.times == nil || w.week == nil:
		return parameterMissing, nil
	default:
		m.reboot = RebootSchedule{Enabled: true, Hour: w.times[0], Min: w.times[1], Week: w.week}
	}
	return writeResult, nil
}

type scheduleWrite struct {
	enable bool
	times  []string // nil when a time is missing or empty
	week   []int    // nil when week is missing
}

// parseSchedule reads a timer write. A time is a string or a number, and a day of the week is a
// number or a string with a number. Other types give ErrInvalidParams.
func parseSchedule(params json.RawMessage, timeKeys ...string) (scheduleWrite, error) {
	p, err := objectParams(params)
	if err != nil {
		return scheduleWrite{}, err
	}
	w := scheduleWrite{enable: truthy(p["enable"])}

	var times []string
	complete := true
	for _, key := range timeKeys {
		raw, ok := p[key]
		if !ok {
			complete = false
			continue
		}
		t, err := stringOrNumber(raw)
		if err != nil {
			return scheduleWrite{}, err
		}
		complete = complete && t != ""
		times = append(times, t)
	}
	if complete {
		w.times = times
	}

	if raw, ok := p["week"]; ok {
		var days []json.RawMessage
		if err := json.Unmarshal(raw, &days); err != nil {
			return scheduleWrite{}, ErrInvalidParams
		}
		w.week = []int{}
		for _, d := range days {
			s, err := stringOrNumber(d)
			if err != nil {
				return scheduleWrite{}, err
			}
			n, err := strconv.Atoi(s)
			if err != nil {
				return scheduleWrite{}, ErrInvalidParams
			}
			w.week = append(w.week, n)
		}
		if len(w.week) == 0 {
			w.week = slices.Clone(allDays)
		}
	}
	return w, nil
}

func stringOrNumber(raw json.RawMessage) (string, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", ErrInvalidParams
	}
	switch v := v.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		return "", ErrInvalidParams
	}
}
