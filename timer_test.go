package glinet_test

import (
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestTimerLEDSchedule(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.Timer.GetLEDSchedule(ctx)
	if err != nil {
		t.Fatalf("GetLEDSchedule: %v", err)
	}
	want := glinet.LEDSchedule{TurnOnHour: "07", TurnOnMin: "00", TurnOffHour: "22", TurnOffMin: "00", Week: []int{0, 1, 2, 3, 4, 5, 6}}
	if !reflect.DeepEqual(*res, want) {
		t.Errorf("GetLEDSchedule = %+v, want %+v", *res, want)
	}

	schedule := glinet.LEDSchedule{Enabled: true, TurnOnHour: "06", TurnOnMin: "10", TurnOffHour: "20", TurnOffMin: "50", Week: []int{2, 4}}
	if err := c.Timer.SetLEDSchedule(ctx, schedule); err != nil {
		t.Fatalf("SetLEDSchedule: %v", err)
	}
	wantState := glinettest.LEDSchedule{Enabled: true, TurnOnHour: "06", TurnOnMin: "10", TurnOffHour: "20", TurnOffMin: "50", Week: []int{2, 4}}
	if got := srv.Timer.LEDSchedule(); !reflect.DeepEqual(got, wantState) {
		t.Errorf("server state = %+v, want %+v", got, wantState)
	}

	if err := c.Timer.SetLEDSchedule(ctx, glinet.LEDSchedule{Enabled: false}); err != nil {
		t.Fatalf("SetLEDSchedule disabled: %v", err)
	}
	wantState.Enabled = false
	if got := srv.Timer.LEDSchedule(); !reflect.DeepEqual(got, wantState) {
		t.Errorf("server state after disable = %+v, want %+v", got, wantState)
	}
}

func TestTimerLEDScheduleParameterMissing(t *testing.T) {
	c, srv := newClient(t)
	initial := srv.Timer.LEDSchedule()

	err := c.Timer.SetLEDSchedule(context.Background(), glinet.LEDSchedule{Enabled: true, TurnOnHour: "06", TurnOnMin: "10", TurnOffHour: "20", TurnOffMin: "50"})
	if code := rpcCode(t, err); code != -1 {
		t.Errorf("code = %d, want -1", code)
	}
	if got := srv.Timer.LEDSchedule(); !reflect.DeepEqual(got, initial) {
		t.Errorf("server state = %+v, want %+v", got, initial)
	}
}

func TestTimerRebootSchedule(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()

	res, err := c.Timer.GetRebootSchedule(ctx)
	if err != nil {
		t.Fatalf("GetRebootSchedule: %v", err)
	}
	want := glinet.RebootSchedule{Hour: "02", Min: "15", Week: []int{3, 5}}
	if !reflect.DeepEqual(*res, want) {
		t.Errorf("GetRebootSchedule = %+v, want %+v", *res, want)
	}

	if err := c.Timer.SetRebootSchedule(ctx, glinet.RebootSchedule{Enabled: true, Hour: "04", Min: "30", Week: []int{0, 6}}); err != nil {
		t.Fatalf("SetRebootSchedule: %v", err)
	}
	wantState := glinettest.RebootSchedule{Enabled: true, Hour: "04", Min: "30", Week: []int{0, 6}}
	if got := srv.Timer.RebootSchedule(); !reflect.DeepEqual(got, wantState) {
		t.Errorf("server state = %+v, want %+v", got, wantState)
	}

	if err := c.Timer.SetRebootSchedule(ctx, glinet.RebootSchedule{Enabled: false, Hour: "05"}); err != nil {
		t.Fatalf("SetRebootSchedule disabled: %v", err)
	}
	wantState.Enabled = false
	if got := srv.Timer.RebootSchedule(); !reflect.DeepEqual(got, wantState) {
		t.Errorf("server state after disable = %+v, want %+v", got, wantState)
	}

	err = c.Timer.SetRebootSchedule(ctx, glinet.RebootSchedule{Enabled: true, Hour: "04", Min: "30"})
	if code := rpcCode(t, err); code != -1 {
		t.Errorf("code = %d, want -1", code)
	}
}

func TestTimerLEDScheduleEmptyWeek(t *testing.T) {
	c, srv := newClient(t)

	schedule := glinet.LEDSchedule{Enabled: true, TurnOnHour: "06", TurnOnMin: "10", TurnOffHour: "20", TurnOffMin: "50", Week: []int{}}
	if err := c.Timer.SetLEDSchedule(context.Background(), schedule); err != nil {
		t.Fatalf("SetLEDSchedule: %v", err)
	}
	if got, want := srv.Timer.LEDSchedule().Week, []int{0, 1, 2, 3, 4, 5, 6}; !reflect.DeepEqual(got, want) {
		t.Errorf("week = %v, want %v", got, want)
	}
}

func TestTimerScreenSchedule(t *testing.T) {
	c, srv := newClient(t)
	ctx := context.Background()
	var sent []string
	handler := srv.Handler("timer", "set_screen")
	srv.Handle("timer", "set_screen", func(params json.RawMessage) (any, error) {
		sent = append(sent, string(params))
		return handler(params)
	})

	res, err := c.Timer.GetScreenSchedule(ctx)
	if err != nil {
		t.Fatalf("GetScreenSchedule: %v", err)
	}
	want := glinet.ScreenSchedule{TurnOnHour: "07", TurnOnMin: "00", TurnOffHour: "22", TurnOffMin: "00", Week: []int{0, 1, 2, 3, 4, 5, 6}}
	if !reflect.DeepEqual(*res, want) {
		t.Errorf("GetScreenSchedule = %+v, want %+v", *res, want)
	}

	if err := c.Timer.SetScreenSchedule(ctx, glinet.ScreenSchedule{Enabled: true, TurnOnHour: "06", TurnOnMin: "10", TurnOffHour: "20", TurnOffMin: "50", Week: []int{2, 4}}); err != nil {
		t.Fatalf("SetScreenSchedule: %v", err)
	}
	wantState := glinettest.ScreenSchedule{Enabled: true, TurnOnHour: "06", TurnOnMin: "10", TurnOffHour: "20", TurnOffMin: "50", Week: []int{2, 4}}
	if got := srv.Timer.ScreenSchedule(); !reflect.DeepEqual(got, wantState) {
		t.Errorf("server state = %+v, want %+v", got, wantState)
	}

	if err := c.Timer.SetScreenSchedule(ctx, glinet.ScreenSchedule{Week: []int{2, 4}}); err != nil {
		t.Fatalf("SetScreenSchedule disabled: %v", err)
	}
	wantState.Enabled = false
	if got := srv.Timer.ScreenSchedule(); !reflect.DeepEqual(got, wantState) {
		t.Errorf("server state after disable = %+v, want %+v", got, wantState)
	}

	wantParams := []string{
		`{"enable":true,"turnon_hour":"06","turnon_min":"10","turnoff_hour":"20","turnoff_min":"50","week":[2,4]}`,
		`{"enable":false,"week":[2,4]}`,
	}
	if !slices.Equal(sent, wantParams) {
		t.Errorf("params = %v, want %v", sent, wantParams)
	}
}
