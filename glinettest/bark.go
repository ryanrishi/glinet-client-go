package glinettest

import (
	"encoding/json"
	"maps"
	"sync"
)

// Bark serves the bark module. The test router does not have it (each read returns -32601); a test
// can replace the methods with ErrMethodNotFound to model that router.
//
// The fake is an assumption from the web UI code of firmware 4.8.3 and the archived API reference:
// get_config returns {"enable": false, "ipv6_enable": false} until set_config sets the fields that
// it gets. get_status returns the status that SetStatus sets, by default
// {"status": 1, "pairing_code": ""}; logout sets pairing_code to "". The writes return null, as in
// the reference. It does not check the values.
type Bark struct {
	mu     sync.Mutex
	config map[string]any
	status map[string]any
}

func newBark() *Bark {
	return &Bark{
		config: map[string]any{"enable": false, "ipv6_enable": false},
		status: map[string]any{"status": 1, "pairing_code": ""},
	}
}

// Config returns the result of get_config.
func (m *Bark) Config() json.RawMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return mustMarshal(m.config)
}

// SetStatus sets the result of get_status.
func (m *Bark) SetStatus(result json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = nil
	mustUnmarshal(string(result), &m.status)
}

func (m *Bark) register(s *Server) {
	handle := func(method string, f func(p map[string]any) any) {
		s.Handle("bark", method, func(params json.RawMessage) (any, error) {
			p, err := paramsObject(params)
			if err != nil {
				return nil, err
			}
			m.mu.Lock()
			defer m.mu.Unlock()
			return mustMarshal(f(p)), nil
		})
	}
	handle("get_config", func(map[string]any) any { return m.config })
	handle("set_config", func(p map[string]any) any {
		maps.Copy(m.config, p)
		return nil
	})
	handle("get_status", func(map[string]any) any { return m.status })
	handle("logout", func(map[string]any) any {
		m.status["pairing_code"] = ""
		return nil
	})
}
