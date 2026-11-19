package glinet

import (
	"context"
	"encoding/json"
	"strings"
)

// VPNPolicyService calls the vpn-policy module: which traffic goes through the VPN client tunnel
// (WireGuard or OpenVPN), and the global VPN policy. Firmware 4.8.3 does not have this module:
// each read returns -32601 (method not found).
type VPNPolicyService service

const vpnPolicyModule = "vpn-policy"

// VPN proxy modes, the values of vpn-policy get_proxy_mode. The domain, MAC and VLAN modes use the
// lists of GetDomainPolicy, GetMACPolicy and GetVLANPolicy. The route modes use the route rules
// of the tunnel (VPNRoute) for VPNProxyCustomRoute; the web UI does not explain VPNProxyAutoDetect.
const (
	VPNProxyGlobal      = 0 // all traffic
	VPNProxyAutoDetect  = 1
	VPNProxyCustomRoute = 2
	VPNProxyDomain      = 3 // by domain or IP address
	VPNProxyMAC         = 4 // by client MAC address
	VPNProxyVLAN        = 5 // by network (VLAN)
)

// VPNGlobalPolicy is the global VPN policy. The router value of each field is 0 or 1; the client
// decodes and encodes these numbers as false and true. The meanings come from the labels of the
// web UI; the label texts are loaded at run time and were not read, so only their keys are known.
//
//   - KillSwitch true blocks traffic that does not go through the VPN (key vpn.kill_switch). The
//     web UI sets WANAccess to false when KillSwitch is true.
//   - WANAccess is shown only while KillSwitch is false (key vpn.use_vpn_only).
//   - ServicePolicy is a switch that is on for false (key vpn.disable_vpn_for_services, or
//     vpn.own_services_use_vpn on some firmware): it is about the traffic of the router's own
//     services, such as the GL.iNet cloud.
//   - VPNServerPolicy is a switch that is on for false (key vpn.enable_vpn_cascading): with false
//     the traffic of the VPN server clients goes through the VPN client tunnel.
type VPNGlobalPolicy struct {
	KillSwitch      bool `json:"kill_switch"`
	WANAccess       bool `json:"wan_access"`
	ServicePolicy   bool `json:"service_policy"`
	VPNServerPolicy bool `json:"vpn_server_policy"`
}

type vpnGlobalPolicyJSON struct {
	KillSwitch      flag `json:"kill_switch"`
	WANAccess       flag `json:"wan_access"`
	ServicePolicy   flag `json:"service_policy"`
	VPNServerPolicy flag `json:"vpn_server_policy"`
}

// MarshalJSON encodes p as the router does, with 0 or 1 for each field.
func (p VPNGlobalPolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(vpnGlobalPolicyJSON{flag(p.KillSwitch), flag(p.WANAccess), flag(p.ServicePolicy), flag(p.VPNServerPolicy)})
}

// UnmarshalJSON decodes the result of vpn-policy get_global_policy. Each field accepts 0, 1, false
// and true.
func (p *VPNGlobalPolicy) UnmarshalJSON(data []byte) error {
	var v vpnGlobalPolicyJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*p = VPNGlobalPolicy{bool(v.KillSwitch), bool(v.WANAccess), bool(v.ServicePolicy), bool(v.VPNServerPolicy)}
	return nil
}

// VPNDomainPolicy is the list of the proxy mode VPNProxyDomain.
//
// DefaultPolicy 0 sends only the traffic to Domains through the VPN; 1 sends all traffic except
// the traffic to Domains through the VPN. It is a number because it selects one of two policies;
// it is not an on/off switch. Domains has domain names, IPv4 addresses and IPv4
// networks, for example "example.com" or "192.168.8.0/24"; a line can end with " #comment". With
// Manual false, the router loads the list from URL instead.
type VPNDomainPolicy struct {
	DefaultPolicy int      `json:"default_policy"`
	Manual        bool     `json:"manual"`
	Domains       []string `json:"domain_list"`
	URL           string   `json:"url"`
}

// UnmarshalJSON decodes the result of vpn-policy get_domain_policy. The web UI accepts
// domain_list as a list or as one string with a domain on each line, so Domains accepts both.
func (p *VPNDomainPolicy) UnmarshalJSON(data []byte) error {
	type plain VPNDomainPolicy
	v := struct {
		*plain
		Domains json.RawMessage `json:"domain_list"`
	}{plain: (*plain)(p)}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	p.Domains = nil
	if len(v.Domains) == 0 || string(v.Domains) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(v.Domains, &s); err == nil {
		p.Domains = strings.FieldsFunc(s, func(r rune) bool { return r == '\n' })
		return nil
	}
	return json.Unmarshal(v.Domains, &p.Domains)
}

// VPNMACPolicy is the list of the proxy mode VPNProxyMAC. DefaultPolicy is as in VPNDomainPolicy,
// for the clients with a MAC address in MACs.
type VPNMACPolicy struct {
	DefaultPolicy int      `json:"default_policy"`
	MACs          []string `json:"mac_list"`
}

// VPNVLAN is an item of the policy of the proxy mode VPNProxyVLAN. ID 1 is the LAN ("Private" in
// the web UI) and 2 the guest network. VPN true sends the traffic of the network through the VPN.
// The router value of VPN is 0 or 1; the client decodes and encodes these numbers.
type VPNVLAN struct {
	ID  int  `json:"id"`
	VPN bool `json:"vpn"`
}

type vpnVLANJSON struct {
	ID  int  `json:"id"`
	VPN flag `json:"vpn"`
}

// MarshalJSON encodes v as the router does, with VPN as 0 or 1.
func (v VPNVLAN) MarshalJSON() ([]byte, error) {
	return json.Marshal(vpnVLANJSON{v.ID, flag(v.VPN)})
}

// UnmarshalJSON decodes an item of vpn-policy get_vlan_policy. VPN accepts 0, 1, false and true.
func (v *VPNVLAN) UnmarshalJSON(data []byte) error {
	var j vpnVLANJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*v = VPNVLAN{j.ID, bool(j.VPN)}
	return nil
}

// GetProxyMode returns the proxy mode, one of the VPNProxy constants.
func (s *VPNPolicyService) GetProxyMode(ctx context.Context) (int, error) {
	var res struct {
		Mode int `json:"mode"`
	}
	if err := s.client.Call(ctx, vpnPolicyModule, "get_proxy_mode", nil, &res); err != nil {
		return 0, err
	}

	return res.Mode, nil
}

// SetProxyMode sets the proxy mode. The web UI waits up to 20 s for the result. It does not offer
// VPNProxyGlobal, VPNProxyDomain, VPNProxyMAC or VPNProxyVLAN while both the WireGuard client and
// the OpenVPN client run.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *VPNPolicyService) SetProxyMode(ctx context.Context, mode int) error {
	params := struct {
		Mode int `json:"mode"`
	}{mode}
	return s.client.Call(ctx, vpnPolicyModule, "set_proxy_mode", params, nil)
}

// GetGlobalPolicy returns the global VPN policy.
func (s *VPNPolicyService) GetGlobalPolicy(ctx context.Context) (*VPNGlobalPolicy, error) {
	var res VPNGlobalPolicy
	if err := s.client.Call(ctx, vpnPolicyModule, "get_global_policy", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetClientPolicy sets KillSwitch, WANAccess and ServicePolicy of the global VPN policy, as the
// client dialog of the web UI does. VPNServerPolicy is not sent. The web UI waits up to 50 s for
// the result.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *VPNPolicyService) SetClientPolicy(ctx context.Context, policy VPNGlobalPolicy) error {
	params := struct {
		KillSwitch    flag `json:"kill_switch"`
		WANAccess     flag `json:"wan_access"`
		ServicePolicy flag `json:"service_policy"`
	}{flag(policy.KillSwitch), flag(policy.WANAccess), flag(policy.ServicePolicy)}
	return s.client.Call(ctx, vpnPolicyModule, "set_global_policy", params, nil)
}

// SetServerPolicy sets only VPNServerPolicy of the global VPN policy, as the server dialog of the
// web UI does.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *VPNPolicyService) SetServerPolicy(ctx context.Context, vpnServerPolicy bool) error {
	params := struct {
		VPNServerPolicy flag `json:"vpn_server_policy"`
	}{flag(vpnServerPolicy)}
	return s.client.Call(ctx, vpnPolicyModule, "set_global_policy", params, nil)
}

// GetDomainPolicy returns the list of the proxy mode VPNProxyDomain.
func (s *VPNPolicyService) GetDomainPolicy(ctx context.Context) (*VPNDomainPolicy, error) {
	var res VPNDomainPolicy
	if err := s.client.Call(ctx, vpnPolicyModule, "get_domain_policy", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetDomainPolicy sets the list of the proxy mode VPNProxyDomain. With Manual true it sends
// default_policy, manual and domain_list as one string with a domain on each line, as the web UI
// does; with Manual false it sends default_policy, manual and url. The web UI requires at least
// one domain when DefaultPolicy is 0 and waits up to 30 s for the result.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *VPNPolicyService) SetDomainPolicy(ctx context.Context, policy VPNDomainPolicy) error {
	params := map[string]any{"default_policy": policy.DefaultPolicy, "manual": policy.Manual}
	if policy.Manual {
		params["domain_list"] = strings.Join(policy.Domains, "\n")
	} else {
		params["url"] = policy.URL
	}
	return s.client.Call(ctx, vpnPolicyModule, "set_domain_policy", params, nil)
}

// GetMACPolicy returns the list of the proxy mode VPNProxyMAC.
func (s *VPNPolicyService) GetMACPolicy(ctx context.Context) (*VPNMACPolicy, error) {
	var res VPNMACPolicy
	if err := s.client.Call(ctx, vpnPolicyModule, "get_mac_policy", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetMACPolicy replaces the list of the proxy mode VPNProxyMAC. The web UI requires at least one
// MAC when DefaultPolicy is 0, and compares MACs without letter case.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *VPNPolicyService) SetMACPolicy(ctx context.Context, policy VPNMACPolicy) error {
	if policy.MACs == nil {
		policy.MACs = []string{}
	}
	return s.client.Call(ctx, vpnPolicyModule, "set_mac_policy", policy, nil)
}

// GetVLANPolicy returns the networks of the proxy mode VPNProxyVLAN.
func (s *VPNPolicyService) GetVLANPolicy(ctx context.Context) ([]VPNVLAN, error) {
	var res struct {
		VLANs []VPNVLAN `json:"vlans"`
	}
	if err := s.client.Call(ctx, vpnPolicyModule, "get_vlan_policy", nil, &res); err != nil {
		return nil, err
	}

	return res.VLANs, nil
}

// SetVLANPolicy sets the networks of the proxy mode VPNProxyVLAN. The web UI sends all items of
// the read with the changed VPN values.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *VPNPolicyService) SetVLANPolicy(ctx context.Context, vlans []VPNVLAN) error {
	params := struct {
		VLANs []VPNVLAN `json:"vlans"`
	}{vlans}
	return s.client.Call(ctx, vpnPolicyModule, "set_vlan_policy", params, nil)
}
