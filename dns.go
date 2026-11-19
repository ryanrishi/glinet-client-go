package glinet

import (
	"context"
	"encoding/json"
)

// DNSService calls the dns module.
type DNSService service

// DNSConfig is the DNS configuration of the router.
//
// Mode is "auto", "manual", "proxy" or "secure". Only "auto" was tested on a router.
// The meaning of Server depends on Mode:
//   - "auto": the upstream DNS servers that the router learned, each as "<interface> <ip>", for
//     example "wwan 192.168.100.53" or "wg 10.66.0.1". They change with the uplink and the VPN.
//   - "manual": the DNS server addresses that were written.
//   - "proxy": one item "<ip>#<port>".
//   - "secure" with Proto "DoH", "DNSCrypt" or "oDoH": the names of the servers from GetInfo.
//
// The router does not return Proto in mode "auto". The web UI reads Proto only in mode
// "secure". It is not known what the router returns in Server for Proto "DoT".
// DoTProvider, NextDNSID, ControlDID and FreeControlDID are for Proto "DoT"; see
// SetDNSConfigRequest.
type DNSConfig struct {
	Mode             string   `json:"mode"`
	ForceDNS         bool     `json:"force_dns"`
	RebindProtection bool     `json:"rebind_protection"`
	OverrideVPN      bool     `json:"override_vpn"`
	Server           []string `json:"server"`
	Proto            string   `json:"proto,omitempty"`
	DoTProvider      string   `json:"dot_provider"`
	NextDNSID        string   `json:"nextdns_id"`
	ControlDID       string   `json:"controld_id"`
	FreeControlDID   string   `json:"free_controld_id"`
}

// SetDNSConfigRequest is the params of dns set_config.
// The router sets ForceDNS, RebindProtection and OverrideVPN to the sent values: a write without
// one of them sets it to false. An empty Mode keeps the mode. The web UI sends no Mode and no
// other field while AdGuard Home is enabled.
//
// Each mode uses only its own fields. The web UI sends:
//   - "auto": no other field.
//   - "manual": Server with 1 to 4 IPv4 or IPv6 addresses. At least one address is IPv4.
//   - "proxy": Server with one item "<ip>#<port>", for example "192.0.2.10#5353". Port is 1 to 65535.
//   - "secure": Proto, one of the protocols of GetInfo ("DoT", "DNSCrypt", "DoH", "oDoH").
//     For "DoT", also DoTProvider: "0" Cloudflare, "1" NextDNS with NextDNSID (the web UI
//     permits an empty ID), "2" Control D with the resolver ID of a paid plan in ControlDID,
//     "3" Control D with a free resolver in FreeControlDID (for example "p0"). For "DoH",
//     "DNSCrypt" and "oDoH", also Server with one or more server names from GetInfo.
//
// The formats of "manual", "proxy" and "secure" come from the web UI code of firmware 4.8.3 and
// were not tested on a router. The web UI shows error -3 as an invalid resolver ID.
type SetDNSConfigRequest struct {
	Mode             string   `json:"mode,omitempty"`
	ForceDNS         bool     `json:"force_dns"`
	RebindProtection bool     `json:"rebind_protection"`
	OverrideVPN      bool     `json:"override_vpn"`
	Server           []string `json:"server,omitempty"`
	Proto            string   `json:"proto,omitempty"`
	DoTProvider      string   `json:"dot_provider,omitempty"`
	NextDNSID        string   `json:"nextdns_id,omitempty"`
	ControlDID       string   `json:"controld_id,omitempty"`
	FreeControlDID   string   `json:"free_controld_id,omitempty"`
}

// DNSInfo is the encrypted DNS protocols and servers that the router supports.
type DNSInfo struct {
	Protos          []string             `json:"protos"`
	ServerList      []DNSEncryptedServer `json:"serverlist"`
	DNSCryptVersion int                  `json:"dnscrypt_version"`
}

// DNSEncryptedServer is a server for mode "secure" with Proto "DoH", "DNSCrypt" or "oDoH".
// Name is the value for SetDNSConfigRequest.Server.
type DNSEncryptedServer struct {
	Name     string `json:"name"`
	Proto    string `json:"proto"`
	DNSSEC   bool   `json:"dnssec"`
	IPv6     bool   `json:"ipv6"`
	NoFilter bool   `json:"nofilter"`
	NoLog    bool   `json:"nolog"`
}

// MarshalJSON sends an empty nextdns_id with DoTProvider "1", as the web UI does, so that a write
// can remove the ID.
func (r SetDNSConfigRequest) MarshalJSON() ([]byte, error) {
	type plain SetDNSConfigRequest
	if r.DoTProvider != "1" || r.NextDNSID != "" {
		return json.Marshal(plain(r))
	}
	return json.Marshal(struct {
		plain
		NextDNSID string `json:"nextdns_id"`
	}{plain: plain(r)})
}

// GetConfig returns the DNS configuration.
func (s *DNSService) GetConfig(ctx context.Context) (*DNSConfig, error) {
	var res DNSConfig
	if err := s.client.Call(ctx, "dns", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetInfo returns the encrypted DNS protocols and servers.
func (s *DNSService) GetInfo(ctx context.Context) (*DNSInfo, error) {
	var res DNSInfo
	if err := s.client.Call(ctx, "dns", "get_info", nil, &res); err != nil {
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
// The DNS server of the router does not use the new file at once: in the test, it did not answer
// with a new name within 5 minutes.
func (s *DNSService) SetHosts(ctx context.Context, content string) error {
	params := struct {
		Content string `json:"content"`
	}{Content: content}
	return s.client.Call(ctx, "dns", "set_host", params, nil)
}
