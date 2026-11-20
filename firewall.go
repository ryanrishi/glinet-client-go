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
//
// After the first SetDMZ, the firewall configuration has a DMZ section, and the router moves it
// to the end after each AddPortForward. When the DMZ section was before the new forward, the
// returned ID is one more than the ID of the forward in the list, and the IDs of the rules and
// forwards that were after the DMZ section change (firmware 4.7.0). Thus the returned ID can be
// different from the ID of the new forward in the list. Do not keep the returned ID for a later
// write: find the forward in ListPortForwards before each SetPortForward or RemovePortForward.
// SetPortForward with an ID that no forward has changes nothing and gives no error. An ID that
// another forward now has makes SetPortForward replace that forward, and RemovePortForward with
// the ID of a rule removes that rule.
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

// OrderPortForwards sets the order of the port forwards to the order of ids, with one
// firewall order_port_forward {"id_list": ids}. The web UI of firmware 4.8.3 sends the IDs of all
// port forwards, in the new order, and then reads the list again.
//
// Not tested on a router: the call is from the web UI code of firmware 4.8.3, and firmware 4.7.0
// does not have it. Not known: if the IDs change, because an ID is the position of the forward in
// the firewall configuration; what the router does with a list that does not have all IDs or has
// an unknown ID; how the order applies to the traffic.
func (s *FirewallService) OrderPortForwards(ctx context.Context, ids []string) error {
	params := struct {
		IDList []string `json:"id_list"`
	}{IDList: append([]string{}, ids...)}
	return s.client.Call(ctx, "firewall", "order_port_forward", params, nil)
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

// RemoveAllPortForwards removes every port forward with firewall remove_port_forward
// {"all": true}, as the "remove all" button of the web UI does.
//
// Not verified on a router: the parameters come from the web UI code of firmware 4.8.3. Whether
// the router also removes rules, and how it changes the IDs of the rules, is not known.
func (s *FirewallService) RemoveAllPortForwards(ctx context.Context) error {
	return s.client.Call(ctx, "firewall", "remove_port_forward", removeAllParams, nil)
}

var removeAllParams = struct {
	All bool `json:"all"`
}{All: true}

// FirewallRule is a traffic rule that applies Target to the traffic from the Src zone to the Dest
// zone. A rule without Dest is for traffic to the router itself.
//
// The router gives the ID. It is the position of the rule in the firewall configuration
// (for example "cfg2392bd"), which rules share with port forwards, so the IDs of the rules
// after a removed rule or port forward change. The router does not check the values and accepts duplicates.
//
// Ports are strings because they can be a range, such as "50004-50005".
// Proto can be "tcp", "udp" or "tcp udp", and Target is "ACCEPT", "REJECT" or "DROP".
// The router writes an empty string as no value.
//
// AddRule was tested on firmware 4.8.3 with an IPv6 SrcIP or DestIP, IPv4 and IPv6 CIDR blocks,
// the Proto values "esp" and "icmpv6", and an upper-case SrcMAC. The router accepted each rule,
// and its nftables rule set (fw4) then had a correct rule.
type FirewallRule struct {
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
func (s *FirewallService) ListRules(ctx context.Context) ([]FirewallRule, error) {
	var res struct {
		Res []FirewallRule `json:"res"`
	}
	if err := s.client.Call(ctx, "firewall", "get_rule_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Res, nil
}

// FindRule returns the rule with the ID id.
// It returns nil and no error when there is no such rule.
func (s *FirewallService) FindRule(ctx context.Context, id string) (*FirewallRule, error) {
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
// Enabled false adds a disabled rule. AddPortForward can change the ID later.
func (s *FirewallService) AddRule(ctx context.Context, rule FirewallRule) (string, error) {
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
// The ID is the position of the rule and can change, so look it up before each call.
//
// Firmware 4.8.3 cannot change a rule that has no dest_port: it returns HTTP status 500 with a Lua
// error page, not a JSON-RPC error. See RemoveRule.
func (s *FirewallService) SetRule(ctx context.Context, rule FirewallRule) error {
	return s.client.Call(ctx, "firewall", "set_rule", rule, nil)
}

// RemoveRule removes the rule with the ID id.
// The router returns HTTP status 500 and changes nothing when no firewall section has that ID
// or when the section has no dest_port. It does not check the section type: with the ID of a
// port forward, it removes that port forward. The ID is the position of the rule and can change,
// so look it up before each call.
//
// On firmware 4.8.3 the error page for a rule that has no dest_port is a Lua error page, not a
// JSON-RPC error. The API cannot remove such a rule. Remove it over SSH with
// uci delete firewall.<section>, then uci commit firewall, then a firewall reload.
func (s *FirewallService) RemoveRule(ctx context.Context, id string) error {
	params := struct {
		ID string `json:"id"`
	}{ID: id}
	return s.client.Call(ctx, "firewall", "remove_rule", params, nil)
}

// RemoveAllRules removes every rule with firewall remove_rule {"all": true}, as the "remove all"
// button of the web UI does.
//
// Not verified on a router: the parameters come from the web UI code of firmware 4.8.3. Whether
// the router also removes port forwards, and how it changes the IDs of the port forwards, is not
// known.
func (s *FirewallService) RemoveAllRules(ctx context.Context) error {
	return s.client.Call(ctx, "firewall", "remove_rule", removeAllParams, nil)
}

// DMZ is the DMZ host configuration. When Enabled, the router forwards incoming traffic
// from the wan zone to the host DestIP. The router does not check DestIP.
//
// Priority is the priority of the DMZ: true has the label key firewall.dmz_priority_higher_label
// and false has firewall.dmz_priority_lower_label in the web UI of firmware 4.8.3. The web UI
// shows the DMZ on the port forward page and reads the port forwards again after each set_dmz,
// but its code does not have the label texts, so it does not settle to what the priority is
// compared. Priority is nil when the router does not return priority: the router returns no
// priority before the first set_dmz, and false after the first set_dmz when no write sent it.
// The web UI shows true when get_dmz has no priority.
type DMZ struct {
	Enabled  bool   `json:"enabled"`
	DestIP   string `json:"dest_ip,omitempty"`
	Priority *bool  `json:"priority,omitempty"`
}

// GetDMZ returns the DMZ configuration. DestIP is empty when the router has no DMZ host.
func (s *FirewallService) GetDMZ(ctx context.Context) (*DMZ, error) {
	var res DMZ
	if err := s.client.Call(ctx, "firewall", "get_dmz", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetDMZ replaces the DMZ configuration. An empty DestIP removes the DMZ host.
// The first SetDMZ adds a DMZ section to the firewall configuration, which can change the IDs
// that AddPortForward returns. The router keeps that section when the DMZ is disabled.
//
// SetDMZ sends priority only when Enabled is true and Priority is not nil. The web UI of firmware
// 4.8.3 sends priority in each write with enabled true, with true when get_dmz had no priority.
// After writes without priority, firmware 4.7.0 returned priority false. The priority write is
// from the web UI code and was not tested on a router: what the router does with it, and if a
// write without priority keeps the value on firmware 4.8.3, is not known.
func (s *FirewallService) SetDMZ(ctx context.Context, dmz DMZ) error {
	params := struct {
		Enabled  bool   `json:"enabled"`
		DestIP   string `json:"dest_ip,omitempty"`
		Priority *bool  `json:"priority,omitempty"`
	}{Enabled: dmz.Enabled, DestIP: dmz.DestIP}
	if dmz.Enabled {
		params.Priority = dmz.Priority
	}
	return s.client.Call(ctx, "firewall", "set_dmz", params, nil)
}

// FirewallZones is the result of firewall get_zone_list: names of the zones of the firewall
// configuration. On the test router both lists are "lan", "wan", "guest" and "wgclient".
type FirewallZones struct {
	Internals []string `json:"internals"`
	Externals []string `json:"externals"`
}

// GetZones returns the firewall zones.
func (s *FirewallService) GetZones(ctx context.Context) (*FirewallZones, error) {
	var res FirewallZones
	if err := s.client.Call(ctx, "firewall", "get_zone_list", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// WANAccess is the access to the router from the WAN (the remote access card of the web UI view
// security). EnableHTTPS permits the web UI over HTTPS, EnableSSH permits SSH and EnablePing
// permits ping. When EnableWhitelist is true, only the sources in Whitelist have this access.
//
// The fixtures of firmware 4.7.0 and 4.8.3 have all fields with an empty Whitelist. The fields of
// a Whitelist item come from the web UI code of firmware 4.8.3.
type WANAccess struct {
	EnableHTTPS     bool              `json:"enable_https"`
	EnableSSH       bool              `json:"enable_ssh"`
	EnablePing      bool              `json:"enable_ping"`
	EnableWhitelist bool              `json:"enable_whitelist"`
	Whitelist       []WANAccessSource `json:"whitelist"`
}

// WANAccessSource is an item of WANAccess.Whitelist. IPAddr is an IPv4 or IPv6 address; the
// web UI does not check for a prefix length. Name is a description of at most 64 characters, and
// the web UI sends "" when the user gives none.
type WANAccessSource struct {
	IPAddr string `json:"ipaddr"`
	Name   string `json:"name"`
}

// GetWANAccess returns the access to the router from the WAN.
func (s *FirewallService) GetWANAccess(ctx context.Context) (*WANAccess, error) {
	var res WANAccess
	if err := s.client.Call(ctx, "firewall", "get_wan_access", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetWANAccess replaces the access to the router from the WAN with one
// firewall set_wan_access {"enable_https", "enable_ssh", "enable_ping", "enable_whitelist",
// "whitelist"}. It always sends all fields, also false and an empty whitelist, as the web UI does.
//
// Not tested on a router: the call is from the web UI code of firmware 4.8.3. After the write,
// the web UI reads the firewall rule list again, so the write possibly changes the firewall rules
// and with them the IDs of the rules and port forwards; this is not known. The web UI shows the
// EnableSSH switch only while SSH is enabled in LocalAccessConfig, and the EnableWhitelist switch
// only while one of EnableHTTPS, EnableSSH and EnablePing is true, but it always sends all fields.
func (s *FirewallService) SetWANAccess(ctx context.Context, access WANAccess) error {
	access.Whitelist = append([]WANAccessSource{}, access.Whitelist...)
	return s.client.Call(ctx, "firewall", "set_wan_access", access, nil)
}
