package glinettest

import (
	"encoding/json"
	"fmt"
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

// Firewall serves the port forward methods of the firewall module. Like the router, it does not
// check the values, accepts duplicates, and gives each forward the ID of its position in the list,
// so the IDs after a removed forward change.
type Firewall struct {
	mu       sync.Mutex
	forwards []PortForward
}

// firstPortForwardIndex is the section index of the first port forward on the test router,
// where the firewall configuration has 35 other sections first.
const firstPortForwardIndex = 0x23

// PortForwardID returns the ID that the router gives to the forward at index i of the list.
func PortForwardID(i int) string {
	return fmt.Sprintf("cfg%02x3837", firstPortForwardIndex+i)
}

func newFirewall() *Firewall {
	return &Firewall{forwards: []PortForward{}}
}

// PortForwards returns a copy of the port forwards.
func (m *Firewall) PortForwards() []PortForward {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.forwards)
}

// SetPortForwards replaces the port forwards.
func (m *Firewall) SetPortForwards(forwards []PortForward) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.forwards = append([]PortForward{}, forwards...)
}

func (m *Firewall) register(s *Server) {
	s.Handle("firewall", "get_port_forward_list", m.getPortForwardList)
	s.Handle("firewall", "add_port_forward", m.addPortForward)
	s.Handle("firewall", "set_port_forward", m.setPortForward)
	s.Handle("firewall", "remove_port_forward", m.removePortForward)
}

type listedPortForward struct {
	ID string `json:"id"`
	PortForward
}

func (m *Firewall) getPortForwardList(json.RawMessage) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]listedPortForward, len(m.forwards))
	for i, f := range m.forwards {
		res[i] = listedPortForward{ID: PortForwardID(i), PortForward: f}
	}
	return map[string]any{"res": res}, nil
}

func (m *Firewall) index(id string) int {
	for i := range m.forwards {
		if PortForwardID(i) == id {
			return i
		}
	}
	return -1
}

// apply writes each field that params has to f. A string field takes a string or a number.
// enabled takes a bool, or a number that is true when it is not 0; the router takes other values as true.
func apply(f *PortForward, params map[string]json.RawMessage) error {
	fields := map[string]*string{
		"name": &f.Name, "proto": &f.Proto, "src": &f.Src, "src_dport": &f.SrcDport,
		"dest": &f.Dest, "dest_ip": &f.DestIP, "dest_port": &f.DestPort,
	}
	for key, dst := range fields {
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
			*dst = v
		case float64:
			*dst = strconv.FormatFloat(v, 'f', -1, 64)
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
			f.Enabled = v
		case float64:
			f.Enabled = v != 0
		default:
			f.Enabled = true
		}
	}
	return nil
}

func (m *Firewall) addPortForward(params json.RawMessage) (any, error) {
	var p map[string]json.RawMessage
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}
	f := PortForward{Enabled: true}
	if err := apply(&f, p); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.forwards = append(m.forwards, f)

	return map[string]string{"id": PortForwardID(len(m.forwards) - 1)}, nil
}

// setPortForward writes the fields that params has and keeps the others, except name:
// no name removes the name, like the router.
func (m *Firewall) setPortForward(params json.RawMessage) (any, error) {
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
	f := m.forwards[i]
	f.Name = ""
	if err := apply(&f, p); err != nil {
		return nil, err
	}
	m.forwards[i] = f

	return writeResult, nil
}

// removePortForward fails with HTTP status 500 when no forward has the ID or the forward has
// no dest_port, like the router.
func (m *Firewall) removePortForward(params json.RawMessage) (any, error) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	i := m.index(p.ID)
	if i < 0 || m.forwards[i].DestPort == "" {
		return nil, errInternalServerError
	}
	m.forwards = slices.Delete(m.forwards, i, i+1)

	return writeResult, nil
}
