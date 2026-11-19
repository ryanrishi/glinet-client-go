package glinettest

import (
	"maps"
	"slices"
)

// vpnRoutes keeps the route rules of a VPN tunnel. Not verified on a router: add_route gives a new
// rule_id, set_route sets the fields that it gets, and remove_route removes the rule. route_flag 6
// selects the IPv6 rules; another value the IPv4 rules. An unknown rule_id gives ErrInvalidParams.
type vpnRoutes struct {
	ipv4, ipv6 []map[string]any
	nextID     int
}

func (r *vpnRoutes) result() map[string]any {
	return map[string]any{"ipv4_route_rules": nonNil(r.ipv4), "ipv6_route_rules": nonNil(r.ipv6)}
}

func (r *vpnRoutes) list(p map[string]any) *[]map[string]any {
	if toInt64(p["route_flag"]) == 6 {
		return &r.ipv6
	}
	return &r.ipv4
}

func ruleFields(p map[string]any) map[string]any {
	rule := maps.Clone(p)
	for _, k := range []string{"route_flag", "group_id", "peer_id"} {
		delete(rule, k)
	}
	return rule
}

func (r *vpnRoutes) add(p map[string]any) {
	r.nextID++
	rule := ruleFields(p)
	rule["rule_id"] = r.nextID
	l := r.list(p)
	*l = append(*l, rule)
}

func (r *vpnRoutes) index(p map[string]any) (*[]map[string]any, int) {
	l := r.list(p)
	id := toInt64(p["rule_id"])
	return l, slices.IndexFunc(*l, func(rule map[string]any) bool { return toInt64(rule["rule_id"]) == id })
}

func (r *vpnRoutes) set(p map[string]any) error {
	l, i := r.index(p)
	if i < 0 {
		return ErrInvalidParams
	}
	maps.Copy((*l)[i], ruleFields(p))
	return nil
}

func (r *vpnRoutes) remove(p map[string]any) error {
	l, i := r.index(p)
	if i < 0 {
		return ErrInvalidParams
	}
	*l = slices.Delete(*l, i, i+1)
	return nil
}

func nonNil(l []map[string]any) []map[string]any {
	if l == nil {
		return []map[string]any{}
	}
	return l
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}
