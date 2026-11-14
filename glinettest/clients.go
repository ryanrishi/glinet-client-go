package glinettest

import "encoding/json"

// Clients serves the clients module. The result of get_list has 3 clients of the test router
// until a test sets it.
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
}
