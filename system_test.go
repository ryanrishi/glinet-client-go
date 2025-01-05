package glinet

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// mockClient implements a mock version of the necessary client interfaces
type mockClient struct {
	Sid string
	// Store the last parameters passed to CallWithStringSlice/CallWithInterface
	lastMethod string
	lastParams interface{}

	// Mock responses
	mockStatusResponse   *GetSystemStatusResponse
	mockTimezoneResponse *GetSystemTimezoneConfigResponse
	mockError            error
}

func (m *mockClient) GetSid() string {
	return m.Sid
}

func (m *mockClient) CallWithStringSlice(method string, params []string, result interface{}) error {
	m.lastMethod = method
	m.lastParams = params

	if m.mockError != nil {
		return m.mockError
	}

	switch params[2] {
	case "get_status":
		responseBytes, _ := json.Marshal(m.mockStatusResponse)
		json.Unmarshal(responseBytes, result)
	case "get_timezone_config":
		responseBytes, _ := json.Marshal(m.mockTimezoneResponse)
		json.Unmarshal(responseBytes, result)
	}

	return nil
}

func (m *mockClient) CallWithInterface(method string, params interface{}, result interface{}) error {
	m.lastMethod = method
	m.lastParams = params

	return m.mockError
}

func (m *mockClient) CallWithInterfaceSlice(method string, params []interface{}, result interface{}) error {
	m.lastMethod = method
	m.lastParams = params

	return m.mockError
}

func TestGetStatus(t *testing.T) {
	// Create mock response
	mockResponse := &GetSystemStatusResponse{
		Network: []struct {
			Online    bool   `json:"online"`
			Up        bool   `json:"up"`
			Interface string `json:"interface"`
		}{
			{Online: true, Up: true, Interface: "eth0"},
		},
		System: struct {
			LanIp           string    `json:"lan_ip"`
			DDNSEnabled     bool      `json:"ddns_enabled"`
			TZOffset        string    `json:"tzoffset"`
			GuestIp         string    `json:"guest_ip"`
			FlashApp        int       `json:"flash_app"`
			FlashTotal      int       `json:"flash_total"`
			MemoryTotal     int       `json:"memory_total"`
			MemoryFree      int       `json:"memory_free"`
			Ipv6Enabled     bool      `json:"ipv6_enabled"`
			MemoryBuffCache int       `json:"memory_buff_cache"`
			Uptime          int       `json:"uptime"`
			LoadAverage     []float64 `json:"load_average"`
			CPU             struct {
				Temperature int `json:"temperature"`
			} `json:"cpu"`
			Mode      int `json:"mode"`
			FlashFree int `json:"flash_free"`
			Timestamp int `json:"timestamp"`
		}{
			LanIp:       "192.168.1.1",
			DDNSEnabled: true,
			TZOffset:    "UTC",
		},
	}

	tests := []struct {
		name    string
		mock    *mockClient
		want    *GetSystemStatusResponse
		wantErr bool
	}{
		{
			name: "successful get status",
			mock: &mockClient{
				mockStatusResponse: mockResponse,
				mockError:          nil,
			},
			want:    mockResponse,
			wantErr: false,
		},
		{
			name: "client error",
			mock: &mockClient{
				mockStatusResponse: nil,
				mockError:          ErrInvalidCredentials,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SystemService{client: tt.mock}
			got, err := s.GetStatus()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTimezoneConfig(t *testing.T) {
	mockResponse := &GetSystemTimezoneConfigResponse{
		Zonename:            "America/Los_Angeles",
		TZOffset:            "-0800",
		AutoTimezoneEnabled: true,
		Localtime:           int(time.Now().Unix()),
		Timezone:            "PST8PDT,M3.2.0,M11.1.0",
	}

	tests := []struct {
		name    string
		mock    *mockClient
		want    *GetSystemTimezoneConfigResponse
		wantErr bool
	}{
		{
			name: "successful get timezone config",
			mock: &mockClient{
				mockTimezoneResponse: mockResponse,
				mockError:            nil,
			},
			want:    mockResponse,
			wantErr: false,
		},
		{
			name: "client error",
			mock: &mockClient{
				mockTimezoneResponse: nil,
				mockError:            ErrInvalidCredentials,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SystemService{client: tt.mock}
			got, err := s.GetTimezoneConfig()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTimezoneConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTimezoneConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetTimezoneConfig(t *testing.T) {
	fixedTimestamp := int64(1672531200) // Use a fixed timestamp for tests

	tests := []struct {
		name    string
		mock    *mockClient
		input   *SetSystemTimezoneRequest
		wantErr bool
	}{
		{
			name: "successful set timezone config",
			mock: &mockClient{
				mockError: nil,
			},
			input: &SetSystemTimezoneRequest{
				Zonename:  "America/Los_Angeles",
				Timezone:  "PST8PDT,M3.2.0,M11.1.0",
				Localtime: fixedTimestamp,
			},
			wantErr: false,
		},
		{
			name: "client error",
			mock: &mockClient{
				mockError: ErrInvalidCredentials,
			},
			input: &SetSystemTimezoneRequest{
				Zonename:  "America/Los_Angeles",
				Timezone:  "PST8PDT,M3.2.0,M11.1.0",
				Localtime: fixedTimestamp,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SystemService{client: tt.mock}
			err := s.SetTimezoneConfig(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetTimezoneConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the correct method was called
				if tt.mock.lastMethod != "call" {
					t.Errorf("SetTimezoneConfig() called method = %v, want %v", tt.mock.lastMethod, "call")
				}

				// Verify parameters were passed correctly
				params, ok := tt.mock.lastParams.([]interface{})
				if !ok {
					t.Errorf("SetTimezoneConfig() invalid params type = %T", tt.mock.lastParams)
					return
				}

				if len(params) != 4 {
					t.Errorf("SetTimezoneConfig() params length = %v, want 4", len(params))
					return
				}

				// Check if the parameters are in the correct order
				if params[1] != "system" || params[2] != "set_timezone_config" {
					t.Errorf("SetTimezoneConfig() invalid params order = %v", params)
				}

				// Check if the timezone config was passed correctly
				config, ok := params[3].(map[string]interface{})
				if !ok {
					t.Errorf("SetTimezoneConfig() invalid config type = %T", params[3])
					return
				}

				if config["zonename"] != tt.input.Zonename {
					t.Errorf("SetTimezoneConfig() zonename = %v, want %v", config["zonename"], tt.input.Zonename)
				}
				if config["timezone"] != tt.input.Timezone {
					t.Errorf("SetTimezoneConfig() timezone = %v, want %v", config["timezone"], tt.input.Timezone)
				}
				if config["localtime"].(int64) != tt.input.Localtime {
					t.Errorf("SetTimezoneConfig() localtime = %v, want %v", config["localtime"], tt.input.Localtime)
				}
			}
		})
	}
}
