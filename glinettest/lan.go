package glinettest

import (
	"encoding/json"
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

// LAN serves the static bind methods of the lan module.
// Like the router, it identifies a binding by its exact MAC string. A write removes each binding
// that has the same MAC or IP and appends the new binding at the end of the list.
type LAN struct {
	mu    sync.Mutex
	binds []StaticBind
}

var macPattern = regexp.MustCompile(`^[0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5}$`)

var writeResult = json.RawMessage(`[]`)

func newLAN() *LAN {
	return &LAN{binds: []StaticBind{}}
}

func (m *LAN) StaticBinds() []StaticBind {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.binds)
}

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
