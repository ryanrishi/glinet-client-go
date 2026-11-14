package glinettest

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSetTimezoneConfigTZOffset(t *testing.T) {
	const november, july = 1700000000, 1720000000
	tests := []struct {
		timezone  string
		localtime int64
		want      string
	}{
		{"UTC0", november, "+0000"},
		{"JST-9", november, "+0900"},
		{"IST-5:30", november, "+0530"},
		{"<+0545>-5:45", november, "+0545"},
		{"EST5EDT,M3.2.0,M11.1.0", november, "-0500"},
		{"EST5EDT,M3.2.0,M11.1.0", july, "-0400"},
		{"CET-1CEST,M3.5.0,M10.5.0/3", july, "+0200"},
		{"NZST-12NZDT,M9.5.0,M4.1.0/3", november, "+1300"},
		{"not a tz string", november, "+0000"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s at %d", tt.timezone, tt.localtime), func(t *testing.T) {
			m := newSystem()
			params, _ := json.Marshal(map[string]any{"zonename": "Test", "timezone": tt.timezone, "localtime": tt.localtime})
			if _, err := m.setTimezoneConfig(params); err != nil {
				t.Fatal(err)
			}
			if got := m.TimezoneConfig().TZOffset; got != tt.want {
				t.Errorf("tzoffset = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSystemReadsMatchRouterFixtures replays the fixtures of the read-only system methods.
func TestSystemReadsMatchRouterFixtures(t *testing.T) {
	replayFixtures(t, "system.get_info.json", "system.get_status.json", "system.disk_info.json")
}
