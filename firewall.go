package glinet

import "context"

// FirewallService calls the firewall module.
type FirewallService service

// PortForward is a port forward from a port of the Src zone to DestIP in the Dest zone.
//
// The router gives the ID. It is the position of the forward in the firewall configuration
// (for example "cfg243837"), which port forwards share with rules, so the IDs of the forwards
// after a removed forward or rule change. The router does not check the values and accepts duplicates.
//
// Ports are strings because they can be a range, such as "50004-50005".
// Proto is "tcp", "udp" or "tcp udp". The router writes an empty Name, Proto or DestPort as no value.
type PortForward struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Proto    string `json:"proto"`
	Src      string `json:"src"`
	SrcDport string `json:"src_dport"`
	Dest     string `json:"dest"`
	DestIP   string `json:"dest_ip"`
	DestPort string `json:"dest_port"`
	Enabled  bool   `json:"enabled"`
}

// ListPortForwards returns the port forwards in the order of the firewall configuration.
func (s *FirewallService) ListPortForwards(ctx context.Context) ([]PortForward, error) {
	var res struct {
		Res []PortForward `json:"res"`
	}
	if err := s.client.Call(ctx, "firewall", "get_port_forward_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Res, nil
}

// FindPortForward returns the port forward with the ID id.
// It returns nil and no error when there is no such port forward.
func (s *FirewallService) FindPortForward(ctx context.Context, id string) (*PortForward, error) {
	forwards, err := s.ListPortForwards(ctx)
	if err != nil {
		return nil, err
	}

	for _, f := range forwards {
		if f.ID == id {
			return &f, nil
		}
	}

	return nil, nil
}

// AddPortForward adds forward at the end of the list and returns its ID. It ignores forward.ID.
// Enabled false adds a disabled forward.
func (s *FirewallService) AddPortForward(ctx context.Context, forward PortForward) (string, error) {
	forward.ID = ""
	var res struct {
		ID string `json:"id"`
	}
	if err := s.client.Call(ctx, "firewall", "add_port_forward", forward, &res); err != nil {
		return "", err
	}

	return res.ID, nil
}

// SetPortForward replaces all values of the port forward with the ID forward.ID. It keeps its position.
// The router changes nothing and returns no error when no forward has that ID.
func (s *FirewallService) SetPortForward(ctx context.Context, forward PortForward) error {
	return s.client.Call(ctx, "firewall", "set_port_forward", forward, nil)
}

// RemovePortForward removes the port forward with the ID id.
// The router returns HTTP status 500 and changes nothing when no firewall section has that ID
// or when the section has no dest_port. It does not check the section type: with the ID of a
// rule, it removes that rule.
func (s *FirewallService) RemovePortForward(ctx context.Context, id string) error {
	params := struct {
		ID string `json:"id"`
	}{ID: id}
	return s.client.Call(ctx, "firewall", "remove_port_forward", params, nil)
}

// Rule is a traffic rule that applies Target to the traffic from the Src zone to the Dest zone.
// A rule without Dest is for traffic to the router itself.
//
// The router gives the ID. It is the position of the rule in the firewall configuration
// (for example "cfg2392bd"), which rules share with port forwards, so the IDs of the rules
// after a removed rule or port forward change. The router does not check the values and accepts duplicates.
//
// Ports are strings because they can be a range, such as "50004-50005".
// Proto is "tcp", "udp" or "tcp udp", and Target is "ACCEPT", "REJECT" or "DROP".
// The router writes an empty string as no value.
type Rule struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Src      string `json:"src"`
	SrcIP    string `json:"src_ip"`
	SrcMAC   string `json:"src_mac"`
	SrcPort  string `json:"src_port"`
	Proto    string `json:"proto"`
	Dest     string `json:"dest"`
	DestIP   string `json:"dest_ip"`
	DestPort string `json:"dest_port"`
	Target   string `json:"target"`
	Enabled  bool   `json:"enabled"`
}

// ListRules returns the rules in the order of the firewall configuration.
func (s *FirewallService) ListRules(ctx context.Context) ([]Rule, error) {
	var res struct {
		Res []Rule `json:"res"`
	}
	if err := s.client.Call(ctx, "firewall", "get_rule_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Res, nil
}

// FindRule returns the rule with the ID id.
// It returns nil and no error when there is no such rule.
func (s *FirewallService) FindRule(ctx context.Context, id string) (*Rule, error) {
	rules, err := s.ListRules(ctx)
	if err != nil {
		return nil, err
	}

	for _, r := range rules {
		if r.ID == id {
			return &r, nil
		}
	}

	return nil, nil
}

// AddRule adds rule at the end of the list and returns its ID. It ignores rule.ID.
// Enabled false adds a disabled rule.
func (s *FirewallService) AddRule(ctx context.Context, rule Rule) (string, error) {
	rule.ID = ""
	var res struct {
		ID string `json:"id"`
	}
	if err := s.client.Call(ctx, "firewall", "add_rule", rule, &res); err != nil {
		return "", err
	}

	return res.ID, nil
}

// SetRule replaces all values of the rule with the ID rule.ID. It keeps its position.
// The router changes nothing and returns no error when no firewall section has that ID.
// It does not check the section type: with the ID of a port forward, it writes to that port forward.
func (s *FirewallService) SetRule(ctx context.Context, rule Rule) error {
	return s.client.Call(ctx, "firewall", "set_rule", rule, nil)
}

// RemoveRule removes the rule with the ID id.
// The router returns HTTP status 500 and changes nothing when no firewall section has that ID
// or when the section has no dest_port. It does not check the section type: with the ID of a
// port forward, it removes that port forward.
func (s *FirewallService) RemoveRule(ctx context.Context, id string) error {
	params := struct {
		ID string `json:"id"`
	}{ID: id}
	return s.client.Call(ctx, "firewall", "remove_rule", params, nil)
}
