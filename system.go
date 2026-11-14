package glinet

import "context"

// SystemService calls the system module.
type SystemService service

// GetSystemStatusResponse is the result of system get_status. Most values are runtime values.
// Wifi[].Passwd is the Wi-Fi key of that interface, which is a secret.
type GetSystemStatusResponse struct {
	Network []struct {
		Online    bool   `json:"online"`
		Up        bool   `json:"up"`
		Interface string `json:"interface"`
	} `json:"network"`
	Wifi []struct {
		Guest      bool   `json:"guest"`
		Ssid       string `json:"ssid"`
		Up         bool   `json:"up"`
		Channel    int    `json:"channel"`
		Band       string `json:"band"`
		Name       string `json:"name"`
		Passwd     string `json:"passwd"`
		Encryption string `json:"encryption"`
		Hidden     bool   `json:"hidden"`
		MLD        bool   `json:"mld"`
	} `json:"wifi"`
	Service []struct {
		Status  int    `json:"status"`
		PeerID  int    `json:"peer_id,omitempty"`
		Name    string `json:"name"`
		GroupID int    `json:"group_id,omitempty"`
	} `json:"service"`
	Client []struct {
		CableTotal    int `json:"cable_total"`
		WirelessTotal int `json:"wireless_total"`
	} `json:"client"`
	System struct {
		LanIP           string    `json:"lan_ip"`
		LanNetmask      string    `json:"lan_netmask"`
		DDNSEnabled     bool      `json:"ddns_enabled"`
		TZOffset        string    `json:"tzoffset"`
		GuestIP         string    `json:"guest_ip"`
		GuestNetmask    string    `json:"guest_netmask"`
		NetNATEnabled   bool      `json:"netnat_enabled"`
		FlashApp        int64     `json:"flash_app"`
		FlashTotal      int64     `json:"flash_total"`
		MemoryTotal     int64     `json:"memory_total"`
		MemoryFree      int64     `json:"memory_free"`
		Ipv6Enabled     bool      `json:"ipv6_enabled"`
		MemoryBuffCache int64     `json:"memory_buff_cache"`
		Uptime          float64   `json:"uptime"`
		LoadAverage     []float64 `json:"load_average"`
		CPU             struct {
			Temperature int `json:"temperature"`
		} `json:"cpu"`
		Mode      int   `json:"mode"`
		FlashFree int64 `json:"flash_free"`
		Timestamp int64 `json:"timestamp"`
	} `json:"system"`
}

// SystemInfo is the result of system get_info: facts about the device and its firmware.
// They change only with a firmware upgrade or a reset.
//
// SN and SNBak are the serial numbers of the device, and DDNS is its device ID for the GL.iNet
// DDNS service. They are secrets.
type SystemInfo struct {
	Vendor          string          `json:"vendor"`
	Model           string          `json:"model"`
	MAC             string          `json:"mac"`
	SN              string          `json:"sn"`
	SNBak           string          `json:"sn_bak"`
	DDNS            string          `json:"ddns"`
	CountryCode     string          `json:"country_code"`
	CPUNum          int             `json:"cpu_num"`
	DeviceType      int             `json:"device_type"`
	HardwareVersion string          `json:"hardware_version"`
	FirmwareVersion string          `json:"firmware_version"`
	FirmwareType    string          `json:"firmware_type"`
	FirmwareDate    string          `json:"firmware_date"`
	BoardInfo       BoardInfo       `json:"board_info"`
	HardwareFeature HardwareFeature `json:"hardware_feature"`
	SoftwareFeature SoftwareFeature `json:"software_feature"`
	HiddenFeatures  []string        `json:"hidden_features"`
}

// BoardInfo is the OpenWrt board information of SystemInfo.
type BoardInfo struct {
	Architecture   string `json:"architecture"`
	Hostname       string `json:"hostname"`
	KernelVersion  string `json:"kernel_version"`
	Model          string `json:"model"`
	OpenWrtVersion string `json:"openwrt_version"`
}

// HardwareFeature is the hardware of the device. Most string values are names of devices or
// interfaces, for example "eth1 eth2" for LAN and "gpio-50" for ResetButton.
type HardwareFeature struct {
	Bluetooth    bool   `json:"bluetooth"`
	BuildInModem string `json:"build_in_modem"`
	Fan          bool   `json:"fan"`
	GPS          bool   `json:"gps"`
	HWNAT        bool   `json:"hwnat"`
	LAN          string `json:"lan"`
	MCU          bool   `json:"mcu"`
	MicroSD      string `json:"microsd"`
	ModemReset   int    `json:"modem_reset"`
	NAND         bool   `json:"nand"`
	NoLED        bool   `json:"noled"`
	NoWDS        bool   `json:"nowds"`
	Radio        string `json:"radio"`
	ResetButton  string `json:"reset_button"`
	RS485        bool   `json:"rs485"`
	Slot         string `json:"slot"`
	Submodel     string `json:"submodel"`
	SwitchButton string `json:"switch_button"`
	USB          string `json:"usb"`
	USB3         string `json:"usb3"`
	USBReset     string `json:"usb_reset"`
	WAN          string `json:"wan"`
}

// SoftwareFeature tells which features the firmware has.
type SoftwareFeature struct {
	AdGuard         bool `json:"adguard"`
	Bark            bool `json:"bark"`
	CellularUpgrade bool `json:"cellular_upgrade"`
	IDSIPS          bool `json:"ids_ips"`
	IPv6            bool `json:"ipv6"`
	KSMBD           bool `json:"ksmbd"`
	MLO             bool `json:"mlo"`
	NAS             bool `json:"nas"`
	Passthrough     bool `json:"passthrough"`
	RepeaterEAP     bool `json:"repeater_eap"`
	SecondWAN       bool `json:"secondwan"`
	SMSForward      bool `json:"sms_forward"`
	Tor             bool `json:"tor"`
	VPN             bool `json:"vpn"`
}

// DiskInfo is the result of system disk_info: the use of the root and /tmp file systems.
// The values are runtime values.
type DiskInfo struct {
	Root DiskUsage `json:"root"`
	Tmp  DiskUsage `json:"tmp"`
}

// DiskUsage is the use of a file system in bytes.
type DiskUsage struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
	Free  int64 `json:"free"`
}

// GetSystemTimezoneConfigResponse is the result of system get_timezone_config.
type GetSystemTimezoneConfigResponse struct {
	Zonename            string `json:"zonename"`
	TZOffset            string `json:"tzoffset"`
	AutoTimezoneEnabled bool   `json:"autotimezone_enabled"`
	Localtime           int    `json:"localtime"`
	Timestamp           int    `json:"timestamp"`
	Timezone            string `json:"timezone"`
}

// SetSystemTimezoneRequest is the params of system set_timezone_config.
type SetSystemTimezoneRequest struct {
	Zonename string `json:"zonename"`
	Timezone string `json:"timezone"`
	// Localtime is a UTC Unix time. The router sets its clock to it. Nil leaves the clock as it is.
	Localtime *int64 `json:"localtime,omitempty"`
}

// GetInfo returns the device and firmware information.
func (s *SystemService) GetInfo(ctx context.Context) (*SystemInfo, error) {
	var res SystemInfo
	if err := s.client.Call(ctx, "system", "get_info", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetDiskInfo returns the use of the file systems.
func (s *SystemService) GetDiskInfo(ctx context.Context) (*DiskInfo, error) {
	var res DiskInfo
	if err := s.client.Call(ctx, "system", "disk_info", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetStatus returns the router status.
func (s *SystemService) GetStatus(ctx context.Context) (*GetSystemStatusResponse, error) {
	var res GetSystemStatusResponse
	if err := s.client.Call(ctx, "system", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetTimezoneConfig returns the time zone configuration.
func (s *SystemService) GetTimezoneConfig(ctx context.Context) (*GetSystemTimezoneConfigResponse, error) {
	var res GetSystemTimezoneConfigResponse
	if err := s.client.Call(ctx, "system", "get_timezone_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetTimezoneConfig sets the time zone configuration.
func (s *SystemService) SetTimezoneConfig(ctx context.Context, config *SetSystemTimezoneRequest) error {
	return s.client.Call(ctx, "system", "set_timezone_config", config, nil)
}
