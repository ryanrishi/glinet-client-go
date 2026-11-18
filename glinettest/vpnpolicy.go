package glinettest

import (
	"encoding/json"
	"maps"
)

// VPNPolicy serves the vpn-policy module. The results of the get methods are those of the test
// router until a test or a write sets them.
//
// The writes are assumptions, because they were not verified on a router; they follow the web UI
// code: each set method sets the fields that it gets in the result of the get method with the same
// name and keeps the others, and returns []. set_global_policy changes get_global_policy. The fake
// keeps domain_list as it was sent (the web UI sends one string with a domain on each line). It does
// not check the values.
type VPNPolicy struct {
	proxyMode, globalPolicy, domainPolicy, macPolicy, vlanPolicy readResult
}

func newVPNPolicy() *VPNPolicy {
	m := &VPNPolicy{}
	m.proxyMode.set(json.RawMessage(vpnProxyMode))
	m.globalPolicy.set(json.RawMessage(vpnGlobalPolicy))
	m.domainPolicy.set(json.RawMessage(vpnDomainPolicy))
	m.macPolicy.set(json.RawMessage(vpnMACPolicy))
	m.vlanPolicy.set(json.RawMessage(vpnVLANPolicy))
	return m
}

// DomainPolicy returns the result of get_domain_policy.
func (m *VPNPolicy) DomainPolicy() json.RawMessage {
	return m.domainPolicy.get()
}

// SetDomainPolicy sets the result of get_domain_policy.
func (m *VPNPolicy) SetDomainPolicy(result json.RawMessage) {
	m.domainPolicy.set(result)
}

func (m *VPNPolicy) register(s *Server) {
	for _, r := range []struct {
		name   string
		result *readResult
	}{
		{"proxy_mode", &m.proxyMode},
		{"global_policy", &m.globalPolicy},
		{"domain_policy", &m.domainPolicy},
		{"mac_policy", &m.macPolicy},
		{"vlan_policy", &m.vlanPolicy},
	} {
		s.Handle("vpn-policy", "get_"+r.name, r.result.handle)
		s.Handle("vpn-policy", "set_"+r.name, func(params json.RawMessage) (any, error) {
			p, err := paramsObject(params)
			if err != nil {
				return nil, err
			}
			err = r.result.update(func(v any) (any, error) {
				result, _ := v.(map[string]any)
				if result == nil {
					result = map[string]any{}
				}
				maps.Copy(result, p)
				return result, nil
			})
			if err != nil {
				return nil, err
			}
			return writeResult, nil
		})
	}
}
