package glinettest

import "encoding/json"

// Wifi serves the get_status method of the wifi module. Its result is that of the test router
// until a test sets it.
type Wifi struct {
	status readResult
}

func newWifi() *Wifi {
	m := &Wifi{}
	m.status.set(json.RawMessage(wifiStatus))
	return m
}

// Status returns the result of get_status.
func (m *Wifi) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *Wifi) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

func (m *Wifi) register(s *Server) {
	s.Handle("wifi", "get_status", m.status.handle)
}
