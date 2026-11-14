package glinettest

import "encoding/json"

// Cable serves the get_status method of the cable module. Its result is that of the test router
// until a test sets it.
type Cable struct {
	status readResult
}

func newCable() *Cable {
	m := &Cable{}
	m.status.set(json.RawMessage(cableStatus))
	return m
}

// Status returns the result of get_status.
func (m *Cable) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *Cable) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

func (m *Cable) register(s *Server) {
	s.Handle("cable", "get_status", m.status.handle)
}
