package glinettest

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sync"
)

// parentalControlConfig is the result of parental-control get_config of the test router: disabled,
// no groups and the two preset rulesets.
const parentalControlConfig = `{
	"auto_update": false,
	"drop_anonymous": false,
	"enable": false,
	"enable_app": false,
	"groups": [],
	"rules": [
		{"color": " #e982a5", "id": "drop", "manual": true, "name": "drop_all", "preset": true, "url_success": "0"},
		{"color": "#4bd9cc", "id": "accept", "manual": true, "name": "accept_all", "preset": true, "url_success": "0"}
	]
}`

// parentalControlApps has 4 of the 206 apps of parental-control get_app_list of the test router.
const parentalControlApps = `{
	"apps": [
		{"id": 1001, "name": "QQ", "type": 1},
		{"id": 4054, "name": "Lazada", "type": 4},
		{"id": 7031, "name": "TeamViewer", "type": 7},
		{"id": 8009, "name": "hao123", "type": 8}
	]
}`

// ParentalControl serves the parental-control module. get_config and get_app_list return the
// results of the test router (4 of its apps) until a test or a write changes them.
//
// The writes are assumptions, because they were not verified on a router; they follow the web UI
// code of firmware 4.8.3 and the archived API reference:
//   - set_config sets enable, drop_anonymous and auto_update when it gets them.
//   - add_group and add_rule add the item with the sent fields and a new ID (a made-up string
//     "cfg" with 6 hex digits), and return {"id": <id>}. add_rule also sets preset to false.
//     add_group without name, default_rule or macs, and add_rule without name, return
//     {"err_code": -1, ...}.
//   - set_group and set_rule set the fields that they get and keep the others. A schedule
//     without an ID gets a new ID; a schedule without week, begin, end or rule gives
//     {"err_code": -2, ...}.
//   - remove_rule returns {"err_code": -2, ...} while a group uses the ruleset as default_rule, and
//     {"err_code": -3, ...} while a group uses it as temporary ruleset.
//   - remove_group also stops the temporary ruleset of the group.
//   - set_brief with enable true stores the sent fields as the result of get_brief; with enable
//     false it stops the temporary ruleset. get_brief of a group with no temporary ruleset returns
//     {"enable": false}.
//   - get_status returns time_valid true and, for each group, the temporary ruleset or else the
//     default ruleset. It does not evaluate the schedules.
//   - get_mode returns {"mode": 1} until set_mode sets another mode.
//   - The writes return {}, as in the reference. An unknown group or ruleset ID gives
//     ErrInvalidParams.
//
// It does not check the values.
type ParentalControl struct {
	mu     sync.Mutex
	config map[string]any
	briefs map[string]map[string]any
	mode   any
	nextID int

	apps readResult
}

var parentalControlWriteResult = json.RawMessage(`{}`)

func newParentalControl() *ParentalControl {
	m := &ParentalControl{briefs: map[string]map[string]any{}, mode: 1}
	mustUnmarshal(parentalControlConfig, &m.config)
	m.apps.set(json.RawMessage(parentalControlApps))
	return m
}

// Config returns the result of get_config.
func (m *ParentalControl) Config() json.RawMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return mustMarshal(m.config)
}

// SetConfig sets the result of get_config.
func (m *ParentalControl) SetConfig(result json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = nil
	mustUnmarshal(string(result), &m.config)
}

// SetApps sets the result of get_app_list.
func (m *ParentalControl) SetApps(result json.RawMessage) {
	m.apps.set(result)
}

func (m *ParentalControl) register(s *Server) {
	handle := func(method string, f func(p map[string]any) (any, error)) {
		s.Handle("parental-control", method, func(params json.RawMessage) (any, error) {
			p, err := paramsObject(params)
			if err != nil {
				return nil, err
			}
			m.mu.Lock()
			defer m.mu.Unlock()
			res, err := f(p)
			if err != nil {
				return nil, err
			}
			return mustMarshal(res), nil
		})
	}
	handle("get_config", func(map[string]any) (any, error) { return m.config, nil })
	handle("set_config", func(p map[string]any) (any, error) {
		for _, k := range []string{"enable", "drop_anonymous", "auto_update"} {
			if v, ok := p[k]; ok {
				m.config[k] = v
			}
		}
		return parentalControlWriteResult, nil
	})
	handle("get_status", m.status)
	s.Handle("parental-control", "get_app_list", m.apps.handle)
	handle("add_group", m.addGroup)
	handle("set_group", m.setGroup)
	handle("remove_group", m.removeGroup)
	handle("add_rule", m.addRule)
	handle("set_rule", m.setRule)
	handle("remove_rule", m.removeRule)
	handle("get_brief", m.getBrief)
	handle("set_brief", m.setBrief)
	handle("get_mode", func(map[string]any) (any, error) { return map[string]any{"mode": m.mode}, nil })
	handle("set_mode", func(p map[string]any) (any, error) {
		m.mode = p["mode"]
		return parentalControlWriteResult, nil
	})
}

func (m *ParentalControl) list(key string) []any {
	l, _ := m.config[key].([]any)
	return l
}

func (m *ParentalControl) index(key string, id any) (int, error) {
	i := slices.IndexFunc(m.list(key), func(item any) bool {
		v, _ := item.(map[string]any)
		return v["id"] == id
	})
	if i < 0 {
		return 0, ErrInvalidParams
	}
	return i, nil
}

func (m *ParentalControl) newID() string {
	m.nextID++
	return fmt.Sprintf("cfg%06x", m.nextID)
}

func parentalControlError(code int, msg string) map[string]any {
	return map[string]any{"err_code": code, "err_msg": msg}
}

func (m *ParentalControl) status(map[string]any) (any, error) {
	groups := []any{}
	for _, item := range m.list("groups") {
		g, _ := item.(map[string]any)
		st := map[string]any{"id": g["id"], "rule": g["default_rule"], "brief": false}
		if b, ok := m.briefs[fmt.Sprint(g["id"])]; ok {
			st["rule"] = b["rule_id"]
			st["brief"] = true
		}
		groups = append(groups, st)
	}
	return map[string]any{"groups": groups, "time_valid": true}, nil
}

// setSchedules gives each schedule without an ID a new ID.
func (m *ParentalControl) setSchedules(v any) error {
	schedules, _ := v.([]any)
	for _, item := range schedules {
		sch, _ := item.(map[string]any)
		for _, k := range []string{"week", "begin", "end", "rule"} {
			if _, ok := sch[k]; !ok {
				return fmt.Errorf("schedule without %s", k)
			}
		}
		if _, ok := sch["id"]; !ok {
			sch["id"] = m.newID()
		}
	}
	return nil
}

func (m *ParentalControl) addGroup(p map[string]any) (any, error) {
	for _, k := range []string{"name", "default_rule", "macs"} {
		if _, ok := p[k]; !ok {
			return parentalControlError(-1, "missing required parameters"), nil
		}
	}
	if s, ok := p["schedules"]; ok {
		if err := m.setSchedules(s); err != nil {
			return parentalControlError(-2, err.Error()), nil
		}
	}
	group := maps.Clone(p)
	group["id"] = m.newID()
	m.config["groups"] = append(m.list("groups"), group)
	return map[string]any{"id": group["id"]}, nil
}

func (m *ParentalControl) setGroup(p map[string]any) (any, error) {
	i, err := m.index("groups", p["id"])
	if err != nil {
		return nil, err
	}
	if s, ok := p["schedules"]; ok {
		if err := m.setSchedules(s); err != nil {
			return parentalControlError(-2, err.Error()), nil
		}
	}
	maps.Copy(m.list("groups")[i].(map[string]any), p)
	return parentalControlWriteResult, nil
}

func (m *ParentalControl) removeGroup(p map[string]any) (any, error) {
	i, err := m.index("groups", p["id"])
	if err != nil {
		return nil, err
	}
	m.config["groups"] = slices.Delete(m.list("groups"), i, i+1)
	delete(m.briefs, fmt.Sprint(p["id"]))
	return parentalControlWriteResult, nil
}

func (m *ParentalControl) addRule(p map[string]any) (any, error) {
	if _, ok := p["name"]; !ok {
		return parentalControlError(-1, "missing required parameters"), nil
	}
	rule := maps.Clone(p)
	rule["id"] = m.newID()
	rule["preset"] = false
	m.config["rules"] = append(m.list("rules"), rule)
	return map[string]any{"id": rule["id"]}, nil
}

func (m *ParentalControl) setRule(p map[string]any) (any, error) {
	i, err := m.index("rules", p["id"])
	if err != nil {
		return nil, err
	}
	maps.Copy(m.list("rules")[i].(map[string]any), p)
	return parentalControlWriteResult, nil
}

func (m *ParentalControl) removeRule(p map[string]any) (any, error) {
	i, err := m.index("rules", p["id"])
	if err != nil {
		return nil, err
	}
	for _, item := range m.list("groups") {
		if g, _ := item.(map[string]any); g["default_rule"] == p["id"] {
			return parentalControlError(-2, "rule is used as default rule"), nil
		}
	}
	for _, b := range m.briefs {
		if b["rule_id"] == p["id"] {
			return parentalControlError(-3, "rule is used as temporary rule"), nil
		}
	}
	m.config["rules"] = slices.Delete(m.list("rules"), i, i+1)
	return parentalControlWriteResult, nil
}

func (m *ParentalControl) getBrief(p map[string]any) (any, error) {
	if _, err := m.index("groups", p["group_id"]); err != nil {
		return nil, err
	}
	if b, ok := m.briefs[fmt.Sprint(p["group_id"])]; ok {
		return b, nil
	}
	return map[string]any{"enable": false}, nil
}

func (m *ParentalControl) setBrief(p map[string]any) (any, error) {
	if _, err := m.index("groups", p["group_id"]); err != nil {
		return nil, err
	}
	id := fmt.Sprint(p["group_id"])
	if p["enable"] == true {
		m.briefs[id] = maps.Clone(p)
	} else {
		delete(m.briefs, id)
	}
	return parentalControlWriteResult, nil
}
