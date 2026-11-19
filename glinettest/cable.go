package glinettest

import "encoding/json"

// Cable serves the get_status, get_config, set_config, get_ports_config and set_port_config methods
// of the cable module. The results are those of the test router until a test or a write sets them.
//
// get_config and set_config use the configuration of the first WAN port, or of the second WAN
// port when the params have "secondwan": true. The configuration of the second WAN port is
// {"protocol":"dhcp"} until a test or a write sets it. This is an assumption: the test router has
// no second WAN port, and what a router does with "secondwan": true is not known.
//
// The write behaviour is an assumption, because set_config was not verified on a router:
// set_config replaces the configuration with its params without "secondwan", as the web UI sends
// the full configuration. It does not check the values.
//
// get_ports_config gives one port "wan" with the made-up MAC address 02:00:00:00:00:02 until a test or a write sets it. This is an assumption: the
// test router did not give this result. set_port_config is also an assumption, because it was not
// tested on a router: it sets mode of the port with the params name, and replaces its macaddr
// with the params macaddr when the params have one, with the defmacaddr of the port kept. It
// changes nothing for an unknown name. It does not check the values.
type Cable struct {
	status       readResult
	config       readResult
	secondConfig readResult
	ports        readResult
}

const cablePorts = `{
	"ports": [
		{
			"default_mode": "wan",
			"macaddr": {"defmacaddr": "02:00:00:00:00:02", "macaddr": "02:00:00:00:00:02", "mode": "default"},
			"mode": "wan",
			"name": "wan",
			"silk": "WAN"
		}
	]
}`

func newCable() *Cable {
	m := &Cable{}
	m.status.set(json.RawMessage(cableStatus))
	m.config.set(json.RawMessage(cableConfig))
	m.secondConfig.set(json.RawMessage(`{"protocol":"dhcp"}`))
	m.ports.set(json.RawMessage(cablePorts))
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

// Config returns the result of get_config for the first WAN port.
func (m *Cable) Config() json.RawMessage {
	return m.config.get()
}

// SetConfig sets the result of get_config for the first WAN port.
func (m *Cable) SetConfig(result json.RawMessage) {
	m.config.set(result)
}

// SecondConfig returns the result of get_config for the second WAN port.
func (m *Cable) SecondConfig() json.RawMessage {
	return m.secondConfig.get()
}

// SetSecondConfig sets the result of get_config for the second WAN port.
func (m *Cable) SetSecondConfig(result json.RawMessage) {
	m.secondConfig.set(result)
}

func (m *Cable) register(s *Server) {
	s.Handle("cable", "get_status", m.status.handle)
	s.Handle("cable", "get_config", m.getConfig)
	s.Handle("cable", "set_config", m.setConfig)
	s.Handle("cable", "get_ports_config", m.ports.handle)
	s.Handle("cable", "set_port_config", m.setPortConfig)
}

// Ports returns the result of get_ports_config.
func (m *Cable) Ports() json.RawMessage {
	return m.ports.get()
}

// SetPorts sets the result of get_ports_config.
func (m *Cable) SetPorts(result json.RawMessage) {
	m.ports.set(result)
}

func (m *Cable) port(p map[string]any) *readResult {
	if p["secondwan"] == true {
		return &m.secondConfig
	}
	return &m.config
}

func (m *Cable) getConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	return m.port(p).get(), nil
}

func (m *Cable) setConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	config := m.port(p)
	delete(p, "secondwan")

	config.set(mustMarshal(p))
	return writeResult, nil
}

func (m *Cable) setPortConfig(params json.RawMessage) (any, error) {
	p, err := paramsObject(params)
	if err != nil {
		return nil, err
	}
	err = m.ports.update(func(v any) (any, error) {
		result, _ := v.(map[string]any)
		ports, _ := result["ports"].([]any)
		for _, port := range ports {
			port, _ := port.(map[string]any)
			if port == nil || port["name"] != p["name"] {
				continue
			}
			port["mode"] = p["mode"]
			mac, ok := p["macaddr"].(map[string]any)
			if !ok {
				continue
			}
			if old, _ := port["macaddr"].(map[string]any); old != nil && old["defmacaddr"] != nil {
				mac["defmacaddr"] = old["defmacaddr"]
			}
			port["macaddr"] = mac
		}
		return v, nil
	})
	if err != nil {
		return nil, err
	}
	return writeResult, nil
}
