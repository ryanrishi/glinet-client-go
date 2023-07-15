package glinet

type SystemService service

type GetSystemStatusResponse struct {
	Network []struct {
		Online    bool   `json:"online"`
		Up        bool   `json:"up"`
		Interface string `json:"interface"`
	} `json:"network"`
	Wifi []struct {
		Guest   bool   `json:"guest"`
		Ssid    string `json:"ssid"`
		Up      bool   `json:"up"`
		Channel int    `json:"channel"`
		Band    string `json:"band"`
		Name    string `json:"name"`
		Passwd  string `json:"passwd"`
	} `json:"wifi"`
	Service []struct {
		Status  int    `json:"status"`
		PeerId  int    `json:"peer_id,omitempty"`
		Name    string `json:"name"`
		GroupId int    `json:"group_id,omitempty"`
	} `json:"service"`
	Client []struct {
		CableTotal    int `json:"cable_total"`
		WirelessTotal int `json:"wireless_total"`
	} `json:"client"`
	System struct {
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
	} `json:"system"`
}

type GetSystemTimezoneConfigResponse struct {
	Zonename            string `json:"zonename"`
	TZOffset            string `json:"tzoffset"`
	AutoTimezoneEnabled bool   `json:"autotimezone_enabled"`
	Localtime           int    `json:"localtime"`
	Timezone            string `json:"timezone"`
}

type SetSystemTimezoneConfigRequest struct {
	Zonename  string `json:"zonename"`
	Localtime int    `json:"localtime,omitempty"`
	Timezone  string `json:"timezone"`
}

func (s *SystemService) GetStatus() (*GetSystemStatusResponse, error) {
	var res GetSystemStatusResponse

	err := s.client.CallWithStringSlice("call", []string{s.client.Sid, "system", "get_status"}, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *SystemService) GetTimezoneConfig() (*GetSystemTimezoneConfigResponse, error) {
	var res GetSystemTimezoneConfigResponse

	err := s.client.CallWithStringSlice("call", []string{s.client.Sid, "system", "get_timezone_config"}, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *SystemService) SetTimezoneConfig(req SetSystemTimezoneConfigRequest) (*[]any, error) {
	var res []any

	var params = make([]interface{}, 4)
	params[0] = s.client.Sid
	params[1] = "system"
	params[2] = "set_timezone_config"
	params[3] = req

	err := s.client.CallWithInterfaceSlice("call", params, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
