package glinettest

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sync"
)

// WireGuardServer serves the wg-server module. The configuration and setting are those of the
// test router (not initialized, no peers), with made-up addresses, until a test or a write changes
// them. The reads of that state match the router.
//
// The writes are assumptions, because they were not verified on a router; they follow the web UI
// code:
//   - generate_key sets initialization to true and sets made-up keys.
//   - set_config and set_setting set the fields that they get and keep the others.
//   - add_peer adds a peer with the sent fields, a new peer_id and the first unused client_ip, and returns
//     {"peer_id": <id>}; set_peer sets the fields that it gets; remove_peer removes the peer.
//   - generate_peer returns a configuration built from the server and the peer.
//   - start and stop set server.status in get_status to 1 and 0.
//   - The other writes return []. An unknown peer gives ErrInvalidParams.
//
// It does not check the values.
type WireGuardServer struct {
	mu      sync.Mutex
	config  map[string]any
	setting map[string]any
	status  int
	peers   []map[string]any
	routes  vpnRoutes
	nextID  int
}

func newWireGuardServer() *WireGuardServer {
	m := &WireGuardServer{peers: []map[string]any{}, nextID: 20000}
	mustUnmarshal(wgServerConfig, &m.config)
	mustUnmarshal(wgServerSetting, &m.setting)
	return m
}

// Config returns the result of get_config.
func (m *WireGuardServer) Config() json.RawMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return mustMarshal(m.config)
}

// SetConfig sets the result of get_config.
func (m *WireGuardServer) SetConfig(result json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mustUnmarshal(string(result), &m.config)
}

func (m *WireGuardServer) register(s *Server) {
	handle := func(method string, f func(p map[string]any) (any, error)) {
		s.Handle("wg-server", method, func(params json.RawMessage) (any, error) {
			p, err := paramsObject(params)
			if err != nil {
				return nil, err
			}
			m.mu.Lock()
			defer m.mu.Unlock()
			res, err := f(p)
			if err != nil {
				return nil, err
			}
			return mustMarshal(res), nil
		})
	}
	handle("get_config", func(map[string]any) (any, error) { return m.config, nil })
	handle("generate_key", func(map[string]any) (any, error) {
		m.config["initialization"] = true
		m.config["private_key"] = "MADE+UP+SERVER+PRIVATE+KEY+FOR+TESTS+0000000="
		m.config["public_key"] = "MADE+UP+SERVER+PUBLIC+KEY+FOR+TESTS+00000000="
		return writeResult, nil
	})
	handle("set_config", func(p map[string]any) (any, error) {
		maps.Copy(m.config, p)
		return writeResult, nil
	})
	handle("get_status", func(map[string]any) (any, error) {
		return map[string]any{"peers": []any{}, "server": map[string]any{
			"initialization": m.config["initialization"], "log": "", "rx_bytes": 0, "status": m.status, "tunnel_ip": "", "tx_bytes": 0,
		}}, nil
	})
	handle("start", func(map[string]any) (any, error) { m.status = 1; return writeResult, nil })
	handle("stop", func(map[string]any) (any, error) { m.status = 0; return writeResult, nil })
	handle("get_peer_list", func(map[string]any) (any, error) { return map[string]any{"peers": m.peers}, nil })
	handle("add_peer", m.addPeer)
	handle("set_peer", m.setPeer)
	handle("remove_peer", m.removePeer)
	handle("generate_peer", m.generatePeer)
	handle("get_setting", func(map[string]any) (any, error) { return m.setting, nil })
	handle("set_setting", func(p map[string]any) (any, error) {
		maps.Copy(m.setting, p)
		return writeResult, nil
	})
	handle("get_route_list", func(map[string]any) (any, error) { return m.routes.result(), nil })
	handle("add_route", func(p map[string]any) (any, error) { m.routes.add(p); return writeResult, nil })
	handle("set_route", func(p map[string]any) (any, error) { return writeResult, m.routes.set(p) })
	handle("remove_route", func(p map[string]any) (any, error) { return writeResult, m.routes.remove(p) })
}

func (m *WireGuardServer) addPeer(p map[string]any) (any, error) {
	m.nextID++
	peer := maps.Clone(p)
	peer["peer_id"] = m.nextID
	peer["client_ip"] = m.unusedClientIP()
	peer["deprecated"] = 0
	m.peers = append(m.peers, peer)
	return map[string]any{"peer_id": m.nextID}, nil
}

func (m *WireGuardServer) unusedClientIP() string {
	for n := 2; ; n++ {
		ip := fmt.Sprintf("10.0.0.%d", n)
		if !slices.ContainsFunc(m.peers, func(peer map[string]any) bool { return peer["client_ip"] == ip }) {
			return ip
		}
	}
}

func (m *WireGuardServer) peerIndex(p map[string]any) (int, error) {
	id := toInt(p["peer_id"])
	i := slices.IndexFunc(m.peers, func(peer map[string]any) bool { return toInt(peer["peer_id"]) == id })
	if i < 0 {
		return 0, ErrInvalidParams
	}
	return i, nil
}

func (m *WireGuardServer) setPeer(p map[string]any) (any, error) {
	i, err := m.peerIndex(p)
	if err != nil {
		return nil, err
	}
	maps.Copy(m.peers[i], p)
	return writeResult, nil
}

func (m *WireGuardServer) removePeer(p map[string]any) (any, error) {
	i, err := m.peerIndex(p)
	if err != nil {
		return nil, err
	}
	m.peers = slices.Delete(m.peers, i, i+1)
	return writeResult, nil
}

func (m *WireGuardServer) generatePeer(p map[string]any) (any, error) {
	i, err := m.peerIndex(p)
	if err != nil {
		return nil, err
	}
	peer := m.peers[i]
	return map[string]any{
		"address":              peer["client_ip"],
		"allowed_ips":          peer["allowed_ips"],
		"dns":                  peer["dns"],
		"domain_end_point":     fmt.Sprintf("router.example.com:%v", m.config["port"]),
		"ip_end_point":         fmt.Sprintf("203.0.113.1:%v", m.config["port"]),
		"persistent_keepalive": peer["persistent_keepalive"],
		"private_key":          "MADE+UP+PEER+PRIVATE+KEY+FOR+TESTS+000000000=",
		"public_key":           m.config["public_key"],
		"presharedkey":         peer["presharedkey"],
		"mtu":                  peer["mtu"],
	}, nil
}
