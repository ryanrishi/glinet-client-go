package glinettest

import "encoding/json"

// Tethering serves the tethering module. get_status returns the result of the test router (no
// phone on USB) and get_config returns {"device":""} until a test or a write sets them.
//
// The write behaviour is an assumption, because the writes and get_config were not tested on a
// router. The fake does not check the values:
//   - set_connect replaces the get_config result with its params and sets the status to 2
//     (connecting).
//   - disconnect sets the status to 0 (not connected).
type Tethering struct {
	config readResult
	status readResult
}

func newTethering() *Tethering {
	m := &Tethering{}
	m.config.set(json.RawMessage(`{"device":""}`))
	m.status.set(json.RawMessage(`{"devices":[],"status":2}`))
	return m
}

// Config returns the result of get_config.
func (m *Tethering) Config() json.RawMessage {
	return m.config.get()
}

// SetConfig sets the result of get_config.
func (m *Tethering) SetConfig(result json.RawMessage) {
	m.config.set(result)
}

// Status returns the result of get_status.
func (m *Tethering) Status() json.RawMessage {
	return m.status.get()
}

// SetStatus sets the result of get_status.
func (m *Tethering) SetStatus(result json.RawMessage) {
	m.status.set(result)
}

func (m *Tethering) register(s *Server) {
	s.Handle("tethering", "get_config", m.config.handle)
	s.Handle("tethering", "get_status", m.status.handle)
	s.Handle("tethering", "set_connect", m.setConnect)
	s.Handle("tethering", "disconnect", func(json.RawMessage) (any, error) {
		return writeResult, m.setStatus(0)
	})
}

func (m *Tethering) setConnect(params json.RawMessage) (any, error) {
	if _, err := paramsObject(params); err != nil {
		return nil, err
	}
	m.config.set(params)
	return writeResult, m.setStatus(2)
}

func (m *Tethering) setStatus(status int) error {
	return m.status.update(func(v any) (any, error) {
		st, _ := v.(map[string]any)
		if st == nil {
			st = map[string]any{}
		}
		st["status"] = status
		return st, nil
	})
}
