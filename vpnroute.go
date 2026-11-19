package glinet

import "strings"

// VPNRoute is a route rule of a VPN tunnel. The router uses the route rules of a tunnel in the
// proxy mode VPNProxyCustomRoute. ID is the rule_id that the router gives the rule.
//
// Not verified on a router: the test router had no route rules, so the fields come from the web
// UI code. The web UI accepts as Scope "host", "link", "global" or a number 1-65535, and as MTU a
// number 1-65535. The web UI of firmware 4.7.0 sends only the fields that are not empty; for the
// web UI of firmware 4.8.3, see WireGuardServerService.AddRoute.
type VPNRoute struct {
	ID      int    `json:"rule_id,omitempty"`
	Dest    string `json:"dest"`
	Gateway string `json:"gateway,omitempty"`
	Metric  int    `json:"metric,omitempty"`
	MTU     int    `json:"mtu,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

// VPNRoutes is the route rules of a VPN tunnel. The web UI shows only the IPv4 rules.
type VPNRoutes struct {
	IPv4 []VPNRoute `json:"ipv4_route_rules"`
	IPv6 []VPNRoute `json:"ipv6_route_rules"`
}

// routeParams returns the params of add_route and set_route as the web UI sends them: the fields
// that are not empty, and route_flag 4 or 6.
func routeParams(r VPNRoute) map[string]any {
	p := map[string]any{"dest": r.Dest, "route_flag": routeFlag(r.Dest)}
	if r.ID != 0 {
		p["rule_id"] = r.ID
	}
	if r.Gateway != "" {
		p["gateway"] = r.Gateway
	}
	if r.Metric != 0 {
		p["metric"] = r.Metric
	}
	if r.MTU != 0 {
		p["mtu"] = r.MTU
	}
	if r.Scope != "" {
		p["scope"] = r.Scope
	}
	return p
}

func routeFlag(dest string) int {
	if strings.Contains(dest, ":") {
		return 6
	}
	return 4
}
