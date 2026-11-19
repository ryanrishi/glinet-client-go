package glinettest

import "encoding/json"

// Clients serves the clients module and the speed limit methods of the qos module. The result of
// get_list has 3 clients of the test router until a test sets it.
//
// The writes are from the web UI code of firmware 4.8.3 and were not tested on a router. They
// change the client in the get_list result that has the params mac with the same letter case,
// and change nothing when no client has it. The fake does not check the values. These are
// assumptions:
//
//   - clients set_info sets alias and class, and removes each one that is "".
//   - qos set_speed_limit_rule sets limit_rx to download and limit_tx to upload.
//   - qos remove_speed_limit_rule sets limit_rx and limit_tx to 0.
type Clients struct {
	list readResult
}

func newClients() *Clients {
	m := &Clients{}
	m.list.set(json.RawMessage(clientsList))
	return m
}

// List returns the result of get_list.
func (m *Clients) List() json.RawMessage {
	return m.list.get()
}

// SetList sets the result of get_list.
func (m *Clients) SetList(result json.RawMessage) {
	m.list.set(result)
}

func (m *Clients) register(s *Server) {
	s.Handle("clients", "get_list", m.list.handle)
	s.Handle("clients", "set_info", m.setInfo)
	s.Handle("qos", "set_speed_limit_rule", m.setSpeedLimit)
	s.Handle("qos", "remove_speed_limit_rule", m.removeSpeedLimit)
}

func (m *Clients) setInfo(params json.RawMessage) (any, error) {
	return m.updateClient(params, func(c, p map[string]any) {
		for _, key := range []string{"alias", "class"} {
			v, ok := p[key]
			switch {
			case !ok:
			case v == "":
				delete(c, key)
			default:
				c[key] = v
			}
		}
	})
}

func (m *Clients) setSpeedLimit(params json.RawMessage) (any, error) {
	return m.updateClient(params, func(c, p map[string]any) {
		c["limit_rx"] = p["download"]
		c["limit_tx"] = p["upload"]
	})
}

func (m *Clients) removeSpeedLimit(params json.RawMessage) (any, error) {
	return m.updateClient(params, func(c, _ map[string]any) {
		c["limit_rx"] = 0
		c["limit_tx"] = 0
	})
}

// updateClient calls f with each client in the get_list result that has the params mac, and with
// the params.
func (m *Clients) updateClient(params json.RawMessage, f func(client, params map[string]any)) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}

	err = m.list.update(func(v any) (any, error) {
		list, _ := v.(map[string]any)
		clients, _ := list["clients"].([]any)
		for _, item := range clients {
			if c, ok := item.(map[string]any); ok && c["mac"] == p["mac"] {
				f(c, p)
			}
		}
		return v, nil
	})
	if err != nil {
		return nil, err
	}
	return writeResult, nil
}
