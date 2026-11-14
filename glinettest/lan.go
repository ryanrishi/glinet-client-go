package glinettest

import (
	"encoding/json"
	"maps"
	"net/netip"
	"regexp"
	"slices"
	"sync"
)

// StaticBind is a DHCP static IP binding. An empty Name means the binding has no name.
type StaticBind struct {
	MAC  string `json:"mac"`
	IP   string `json:"ip"`
	Name string `json:"name,omitempty"`
}

// LAN serves the static bind, get_wan_info, get_config_list and set_config methods of the lan
// module. The results of get_wan_info and get_config_list are those of the test router until a
// test or a write sets them.
// Like the router, it identifies a binding by its exact MAC string. A write removes each binding
// that has the same MAC or IP and appends the new binding at the end of the list.
//
// The behaviour of set_config is an assumption, because it was not verified on a router: it sets
// each field that it gets in the item of get_config_list with the same "interface" and keeps the
// other fields. Not known: which fields the router keeps when a write does not send them. It does
// not check the values. An unknown interface gives ErrInvalidParams.
type LAN struct {
	mu    sync.Mutex
	binds []StaticBind

	wanInfo    readResult
	configList readResult
}

var macPattern = regexp.MustCompile(`^[0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5}$`)

var writeResult = json.RawMessage(`[]`)

func newLAN() *LAN {
	m := &LAN{binds: []StaticBind{}}
	m.wanInfo.set(json.RawMessage(lanWANInfo))
	m.configList.set(json.RawMessage(lanConfigList))
	return m
}

// ConfigList returns the result of get_config_list.
func (m *LAN) ConfigList() json.RawMessage {
	return m.configList.get()
}

// SetConfigList sets the result of get_config_list.
func (m *LAN) SetConfigList(result json.RawMessage) {
	m.configList.set(result)
}

// WANInfo returns the result of get_wan_info.
func (m *LAN) WANInfo() json.RawMessage {
	return m.wanInfo.get()
}

// SetWANInfo sets the result of get_wan_info.
func (m *LAN) SetWANInfo(result json.RawMessage) {
	m.wanInfo.set(result)
}

// StaticBinds returns a copy of the bindings.
func (m *LAN) StaticBinds() []StaticBind {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.binds)
}

// SetStaticBinds replaces the bindings.
func (m *LAN) SetStaticBinds(binds []StaticBind) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.binds = append([]StaticBind{}, binds...)
}

func (m *LAN) register(s *Server) {
	s.Handle("lan", "get_static_bind_list", func(json.RawMessage) (any, error) {
		return map[string]any{"static_bind_list": m.StaticBinds()}, nil
	})
	s.Handle("lan", "add_static_bind", m.addStaticBind)
	s.Handle("lan", "set_static_bind", m.setStaticBind)
	s.Handle("lan", "remove_static_bind", m.removeStaticBind)
	s.Handle("lan", "get_wan_info", m.wanInfo.handle)
	s.Handle("lan", "get_config_list", m.configList.handle)
	s.Handle("lan", "set_config", m.setConfig)
}

func (m *LAN) setConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}

	err = m.configList.update(func(v any) (any, error) {
		list, _ := v.(map[string]any)
		interfaces, _ := list["interfaces"].([]any)
		for _, item := range interfaces {
			if config, ok := item.(map[string]any); ok && config["interface"] == p["interface"] {
				maps.Copy(config, p)
				return list, nil
			}
		}
		return nil, ErrInvalidParams
	})
	if err != nil {
		return nil, err
	}
	return writeResult, nil
}

func invalidParams(detail string) *Error {
	return &Error{Code: ErrInvalidParams.Code, Message: ErrInvalidParams.Message, Err: detail}
}

func parseStaticBind(params json.RawMessage) (StaticBind, error) {
	var p struct {
		MAC  string  `json:"mac"`
		IP   string  `json:"ip"`
		Name *string `json:"name"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return StaticBind{}, ErrInvalidParams
	}
	if !macPattern.MatchString(p.MAC) {
		return StaticBind{}, invalidParams("invalid mac")
	}
	if ip, err := netip.ParseAddr(p.IP); err != nil || !ip.Is4() {
		return StaticBind{}, invalidParams("invalid ip")
	}
	if p.Name != nil && *p.Name == "" {
		return StaticBind{}, invalidParams("invalid name")
	}

	b := StaticBind{MAC: p.MAC, IP: p.IP}
	if p.Name != nil {
		b.Name = *p.Name
	}
	return b, nil
}

func (m *LAN) index(mac string) int {
	return slices.IndexFunc(m.binds, func(b StaticBind) bool { return b.MAC == mac })
}

func (m *LAN) put(b StaticBind) {
	m.binds = slices.DeleteFunc(m.binds, func(o StaticBind) bool { return o.MAC == b.MAC || o.IP == b.IP })
	m.binds = append(m.binds, b)
}

func (m *LAN) addStaticBind(params json.RawMessage) (any, error) {
	b, err := parseStaticBind(params)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.put(b)

	return writeResult, nil
}

func (m *LAN) setStaticBind(params json.RawMessage) (any, error) {
	b, err := parseStaticBind(params)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.index(b.MAC) >= 0 {
		m.put(b)
	}

	return writeResult, nil
}

func (m *LAN) removeStaticBind(params json.RawMessage) (any, error) {
	var p struct {
		Mode *int   `json:"mode"`
		MAC  string `json:"mac"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}
	if p.Mode == nil || *p.Mode != 0 {
		return nil, invalidParams("invalid mode")
	}
	if !macPattern.MatchString(p.MAC) {
		return nil, invalidParams("invalid mac")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if i := m.index(p.MAC); i >= 0 {
		m.binds = slices.Delete(m.binds, i, i+1)
	}

	return writeResult, nil
}
