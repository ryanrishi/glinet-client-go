package glinet

import "context"

// DNSService calls the dns module.
type DNSService service

// DNSConfig is the DNS configuration of the router.
//
// Mode is "auto", "secure", "manual" or "proxy" (from the web UI; only "auto" was tested).
// In mode "auto", Server is the upstream DNS servers that the router learned, each as
// "<interface> <ip>", for example "wwan 192.168.4.253" or "wg 192.168.5.1". They change with the
// uplink and the VPN. DoTProvider, NextDNSID, ControlDID and FreeControlDID are for DNS over TLS
// in mode "secure".
type DNSConfig struct {
	Mode             string   `json:"mode"`
	ForceDNS         bool     `json:"force_dns"`
	RebindProtection bool     `json:"rebind_protection"`
	OverrideVPN      bool     `json:"override_vpn"`
	Server           []string `json:"server"`
	DoTProvider      string   `json:"dot_provider"`
	NextDNSID        string   `json:"nextdns_id"`
	ControlDID       string   `json:"controld_id"`
	FreeControlDID   string   `json:"free_controld_id"`
}

// SetDNSConfigRequest is the params of dns set_config.
// The router sets ForceDNS, RebindProtection and OverrideVPN to the sent values: a write without
// one of them sets it to false. An empty Mode keeps the mode.
type SetDNSConfigRequest struct {
	Mode             string `json:"mode,omitempty"`
	ForceDNS         bool   `json:"force_dns"`
	RebindProtection bool   `json:"rebind_protection"`
	OverrideVPN      bool   `json:"override_vpn"`
}

// GetConfig returns the DNS configuration.
func (s *DNSService) GetConfig(ctx context.Context) (*DNSConfig, error) {
	var res DNSConfig
	if err := s.client.Call(ctx, "dns", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the DNS configuration. It returns after about 1 s.
func (s *DNSService) SetConfig(ctx context.Context, config SetDNSConfigRequest) error {
	return s.client.Call(ctx, "dns", "set_config", config, nil)
}

// GetHosts returns the hosts file of the router, for example
// "127.0.0.1 localhost\n\n::1     localhost ip6-localhost ip6-loopback\n".
func (s *DNSService) GetHosts(ctx context.Context) (string, error) {
	var res struct {
		Content string `json:"content"`
	}
	if err := s.client.Call(ctx, "dns", "get_host", nil, &res); err != nil {
		return "", err
	}

	return res.Content, nil
}

// SetHosts replaces the hosts file with content. Keep the localhost lines of GetHosts in it.
// The router does not check the lines and keeps content as it is, with comments, blank lines and
// "\r\n", but adds "\n" when content does not end with one.
// In the test, the DNS server of the router did not answer with a new name within 30 s.
func (s *DNSService) SetHosts(ctx context.Context, content string) error {
	params := struct {
		Content string `json:"content"`
	}{Content: content}
	return s.client.Call(ctx, "dns", "set_host", params, nil)
}
