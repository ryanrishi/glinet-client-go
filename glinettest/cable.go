package glinettest

import "encoding/json"

// Cable serves the get_status, get_config and set_config methods of the cable module. The results
// are those of the test router until a test or a write sets them.
//
// The write behaviour is an assumption, because set_config was not verified on a router:
// set_config replaces the configuration with its params without "secondwan", as the web UI sends
// the full configuration. It does not check the values.
type Cable struct {
	status readResult
	config readResult
}

func newCable() *Cable {
	m := &Cable{}
	m.status.set(json.RawMessage(cableStatus))
	m.config.set(json.RawMessage(cableConfig))
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

// Config returns the result of get_config.
func (m *Cable) Config() json.RawMessage {
	return m.config.get()
}

// SetConfig sets the result of get_config.
func (m *Cable) SetConfig(result json.RawMessage) {
	m.config.set(result)
}

func (m *Cable) register(s *Server) {
	s.Handle("cable", "get_status", m.status.handle)
	s.Handle("cable", "get_config", m.config.handle)
	s.Handle("cable", "set_config", m.setConfig)
}

func (m *Cable) setConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	delete(p, "secondwan")

	m.config.set(mustMarshal(p))
	return writeResult, nil
}
