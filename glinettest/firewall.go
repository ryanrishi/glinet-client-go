package glinettest

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"sync"
)

// PortForward is a port forward of the firewall module. An empty string means the router has no
// value for that field. The ID is not stored: the router gives it from the position in the list.
type PortForward struct {
	Name     string `json:"name,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Src      string `json:"src,omitempty"`
	SrcDport string `json:"src_dport,omitempty"`
	Dest     string `json:"dest,omitempty"`
	DestIP   string `json:"dest_ip,omitempty"`
	DestPort string `json:"dest_port,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// Rule is a traffic rule of the firewall module. An empty string means the router has no
// value for that field. The ID is not stored: the router gives it from the position in the list.
type Rule struct {
	Name     string `json:"name,omitempty"`
	Src      string `json:"src,omitempty"`
	SrcIP    string `json:"src_ip,omitempty"`
	SrcMAC   string `json:"src_mac,omitempty"`
	SrcPort  string `json:"src_port,omitempty"`
	Proto    string `json:"proto,omitempty"`
	Dest     string `json:"dest,omitempty"`
	DestIP   string `json:"dest_ip,omitempty"`
	DestPort string `json:"dest_port,omitempty"`
	Target   string `json:"target,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// DMZ is the DMZ configuration of the firewall module. An empty DestIP means the router has no value.
type DMZ struct {
	Enabled  bool   `json:"enabled"`
	DestIP   string `json:"dest_ip,omitempty"`
	Priority bool   `json:"priority"`
}

// Firewall serves the port forward, rule, DMZ and zone methods of the firewall module. Like the router,
// it does not check the values and accepts duplicates. It keeps the port forwards and rules in one
// list of firewall sections and gives each one the ID of its position in that list, so the IDs
// after a removed port forward or rule change. Like the router, the set and remove methods of
// one type also write or remove a section of the other type when they get its ID.
//
// The first set_dmz adds a DMZ section without an ID at the end of the list. Like the router,
// add_port_forward moves it to the end again after it gives the ID of the new port forward,
// so that ID is one more than the ID in the list when the DMZ section was before the new forward.
//
// order_port_forward puts the port forwards of id_list in that order, and then the other port
// forwards in their order, in the positions that the port forwards had. So a port forward that
// moves gets the ID of its new position. It ignores unknown IDs. This is an assumption: the
// method is from the web UI code of firmware 4.8.3 and was not tested on a router.
//
// remove_port_forward and remove_rule with {"all": true} remove every section of their own type
// and keep the other sections. This is an assumption: the call is from the web UI code of
// firmware 4.8.3 and was not tested on a router.
//
// The result of get_zone_list is that of the test router until a test sets it.
type Firewall struct {
	mu       sync.Mutex
	sections []section

	zones readResult
}

type section struct {
	suffix string
	values map[string]any
}

// The ID suffixes that the router gives from the section type. The DMZ section has a name and no ID.
const (
	portForwardSuffix = "3837"
	ruleSuffix        = "92bd"
	dmzSuffix         = ""
)

// firstSectionIndex is the index of the first port forward or rule on the test router,
// where the firewall configuration has 35 other sections first.
const firstSectionIndex = 0x23

var (
	portForwardKeys = []string{"name", "proto", "src", "src_dport", "dest", "dest_ip", "dest_port"}
	ruleKeys        = []string{"name", "src", "src_ip", "src_mac", "src_port", "proto", "dest", "dest_ip", "dest_port", "target"}
)

// PortForwardID returns the ID that the router gives to a port forward at index i of the
// port forwards and rules together.
func PortForwardID(i int) string {
	return sectionID(i, portForwardSuffix)
}

// RuleID returns the ID that the router gives to a rule at index i of the
// port forwards and rules together.
func RuleID(i int) string {
	return sectionID(i, ruleSuffix)
}

func sectionID(i int, suffix string) string {
	return fmt.Sprintf("cfg%02x%s", firstSectionIndex+i, suffix)
}

func newFirewall() *Firewall {
	m := &Firewall{}
	m.zones.set(json.RawMessage(firewallZones))
	return m
}

// Zones returns the result of get_zone_list.
func (m *Firewall) Zones() json.RawMessage {
	return m.zones.get()
}

// SetZones sets the result of get_zone_list.
func (m *Firewall) SetZones(result json.RawMessage) {
	m.zones.set(result)
}

// PortForwards returns a copy of the port forwards.
func (m *Firewall) PortForwards() []PortForward {
	m.mu.Lock()
	defer m.mu.Unlock()
	return sectionsOf[PortForward](m.sections, portForwardSuffix)
}

// SetPortForwards replaces the port forwards. The new port forwards go after the rules
// and before the DMZ section.
func (m *Firewall) SetPortForwards(forwards []PortForward) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sections = replaceSections(m.sections, portForwardSuffix, forwards)
	m.moveDMZToEnd()
}

// Rules returns a copy of the rules.
func (m *Firewall) Rules() []Rule {
	m.mu.Lock()
	defer m.mu.Unlock()
	return sectionsOf[Rule](m.sections, ruleSuffix)
}

// SetRules replaces the rules. The new rules go after the port forwards
// and before the DMZ section.
func (m *Firewall) SetRules(rules []Rule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sections = replaceSections(m.sections, ruleSuffix, rules)
	m.moveDMZToEnd()
}

// DMZ returns the DMZ configuration.
func (m *Firewall) DMZ() DMZ {
	m.mu.Lock()
	defer m.mu.Unlock()
	var dmz DMZ
	if i := m.dmzIndex(); i >= 0 {
		data, _ := json.Marshal(m.sections[i].values)
		_ = json.Unmarshal(data, &dmz)
	}
	return dmz
}

// SetDMZ replaces the DMZ configuration. It adds the DMZ section at the end if there is none.
func (m *Firewall) SetDMZ(dmz DMZ) {
	m.mu.Lock()
	defer m.mu.Unlock()
	values := map[string]any{"enabled": dmz.Enabled, "priority": dmz.Priority}
	if dmz.DestIP != "" {
		values["dest_ip"] = dmz.DestIP
	}
	m.setDMZValues(values)
}

func (m *Firewall) dmzIndex() int {
	return slices.IndexFunc(m.sections, func(s section) bool { return s.suffix == dmzSuffix })
}

func (m *Firewall) setDMZValues(values map[string]any) {
	if i := m.dmzIndex(); i >= 0 {
		m.sections[i].values = values
		return
	}
	m.sections = append(m.sections, section{suffix: dmzSuffix, values: values})
}

func (m *Firewall) moveDMZToEnd() {
	if i := m.dmzIndex(); i >= 0 {
		dmz := m.sections[i]
		m.sections = append(slices.Delete(m.sections, i, i+1), dmz)
	}
}

func sectionsOf[T any](sections []section, suffix string) []T {
	res := []T{}
	for _, s := range sections {
		if s.suffix != suffix {
			continue
		}
		var v T
		data, _ := json.Marshal(s.values)
		_ = json.Unmarshal(data, &v)
		res = append(res, v)
	}
	return res
}

func replaceSections[T any](sections []section, suffix string, items []T) []section {
	res := slices.DeleteFunc(slices.Clone(sections), func(s section) bool { return s.suffix == suffix })
	for _, item := range items {
		var values map[string]any
		data, _ := json.Marshal(item)
		_ = json.Unmarshal(data, &values)
		res = append(res, section{suffix: suffix, values: values})
	}
	return res
}

func (m *Firewall) register(s *Server) {
	s.Handle("firewall", "get_port_forward_list", m.list(portForwardSuffix))
	s.Handle("firewall", "add_port_forward", m.add(portForwardSuffix, portForwardKeys))
	s.Handle("firewall", "set_port_forward", m.set(portForwardKeys))
	s.Handle("firewall", "remove_port_forward", m.remove(portForwardSuffix))
	s.Handle("firewall", "order_port_forward", m.orderPortForwards)
	s.Handle("firewall", "get_rule_list", m.list(ruleSuffix))
	s.Handle("firewall", "add_rule", m.add(ruleSuffix, ruleKeys))
	s.Handle("firewall", "set_rule", m.set(ruleKeys))
	s.Handle("firewall", "remove_rule", m.remove(ruleSuffix))
	s.Handle("firewall", "get_zone_list", m.zones.handle)
	s.Handle("firewall", "get_dmz", m.getDMZ)
	s.Handle("firewall", "set_dmz", m.setDMZ)
}

func (m *Firewall) list(suffix string) Method {
	return func(json.RawMessage) (any, error) {
		m.mu.Lock()
		defer m.mu.Unlock()
		res := []map[string]any{}
		for i, s := range m.sections {
			if s.suffix != suffix {
				continue
			}
			item := maps.Clone(s.values)
			item["id"] = sectionID(i, suffix)
			res = append(res, item)
		}
		return map[string]any{"res": res}, nil
	}
}

func (m *Firewall) index(id string) int {
	for i, s := range m.sections {
		if s.suffix != dmzSuffix && sectionID(i, s.suffix) == id {
			return i
		}
	}
	return -1
}

// apply writes each of keys that params has to values. A string field takes a string or a
// number, and "" removes it. enabled takes a bool, or a number that is true when it is not 0;
// the router takes other values as true.
func apply(values map[string]any, params map[string]json.RawMessage, keys []string) error {
	for _, key := range keys {
		raw, ok := params[key]
		if !ok {
			continue
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return ErrInvalidParams
		}
		switch v := v.(type) {
		case string:
			if v == "" {
				delete(values, key)
			} else {
				values[key] = v
			}
		case float64:
			values[key] = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			return ErrInvalidParams
		}
	}

	if raw, ok := params["enabled"]; ok {
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return ErrInvalidParams
		}
		switch v := v.(type) {
		case bool:
			values["enabled"] = v
		case float64:
			values["enabled"] = v != 0
		default:
			values["enabled"] = true
		}
	}
	return nil
}

func (m *Firewall) add(suffix string, keys []string) Method {
	return func(params json.RawMessage) (any, error) {
		var p map[string]json.RawMessage
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, ErrInvalidParams
		}
		values := map[string]any{"enabled": true}
		if err := apply(values, p, keys); err != nil {
			return nil, err
		}

		m.mu.Lock()
		defer m.mu.Unlock()
		m.sections = append(m.sections, section{suffix: suffix, values: values})
		id := sectionID(len(m.sections)-1, suffix)
		if suffix == portForwardSuffix {
			m.moveDMZToEnd()
		}

		return map[string]string{"id": id}, nil
	}
}

// set writes the fields that params has and keeps the others, except name:
// no name removes the name, like the router. An unknown ID changes nothing.
func (m *Firewall) set(keys []string) Method {
	return func(params json.RawMessage) (any, error) {
		var p map[string]json.RawMessage
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, ErrInvalidParams
		}
		var id string
		if err := json.Unmarshal(p["id"], &id); err != nil {
			return nil, invalidParams("id: string required")
		}

		m.mu.Lock()
		defer m.mu.Unlock()
		i := m.index(id)
		if i < 0 {
			return writeResult, nil
		}
		values := maps.Clone(m.sections[i].values)
		delete(values, "name")
		if err := apply(values, p, keys); err != nil {
			return nil, err
		}
		m.sections[i].values = values

		return writeResult, nil
	}
}

// remove fails with HTTP status 500 when no section has the ID or the section has
// no dest_port, like the router.
func (m *Firewall) remove(suffix string) Method {
	return func(params json.RawMessage) (any, error) {
		var p struct {
			ID  string `json:"id"`
			All bool   `json:"all"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, ErrInvalidParams
		}

		m.mu.Lock()
		defer m.mu.Unlock()
		if p.All {
			m.sections = slices.DeleteFunc(m.sections, func(s section) bool { return s.suffix == suffix })
			return writeResult, nil
		}
		i := m.index(p.ID)
		if i < 0 || m.sections[i].values["dest_port"] == nil {
			return nil, ErrInternalServerError
		}
		m.sections = slices.Delete(m.sections, i, i+1)

		return writeResult, nil
	}
}

// getDMZ returns only enabled until the first set_dmz, like the router.
func (m *Firewall) getDMZ(json.RawMessage) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i := m.dmzIndex()
	if i < 0 {
		return map[string]any{"enabled": false}, nil
	}
	return maps.Clone(m.sections[i].values), nil
}

// setDMZ writes enabled, dest_ip and priority. No enabled is true and no dest_ip removes the IP,
// like the router. No priority is false, as the router gave after writes without priority; a
// write of priority was not tested on a router.
func (m *Firewall) setDMZ(params json.RawMessage) (any, error) {
	var p map[string]json.RawMessage
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}
	values := map[string]any{"enabled": true, "priority": false}
	if err := apply(values, p, []string{"dest_ip"}); err != nil {
		return nil, err
	}
	if raw, ok := p["priority"]; ok {
		var priority bool
		if err := json.Unmarshal(raw, &priority); err != nil {
			return nil, ErrInvalidParams
		}
		values["priority"] = priority
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.setDMZValues(values)

	return writeResult, nil
}

func (m *Firewall) orderPortForwards(params json.RawMessage) (any, error) {
	var p struct {
		IDList []string `json:"id_list"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	var positions, order []int
	for i, s := range m.sections {
		if s.suffix == portForwardSuffix {
			positions = append(positions, i)
		}
	}
	for _, id := range p.IDList {
		if i := m.index(id); i >= 0 && m.sections[i].suffix == portForwardSuffix && !slices.Contains(order, i) {
			order = append(order, i)
		}
	}
	for _, i := range positions {
		if !slices.Contains(order, i) {
			order = append(order, i)
		}
	}
	sections := slices.Clone(m.sections)
	for j, i := range order {
		m.sections[positions[j]] = sections[i]
	}

	return writeResult, nil
}
