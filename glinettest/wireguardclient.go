package glinettest

import (
	"encoding/json"
	"maps"
	"net"
	"slices"
	"strconv"
	"sync"
)

// WireGuardClient serves the wg-client module. The groups, peers, settings and status are those of
// the test router, with made-up names, addresses and keys, until a test or a write changes them.
//
// The reads match the router: get_group_list, get_config_list, get_setting, get_route_list and
// get_status. get_all_config_list has the groups with their peers; the router leaves some empty
// peer fields out of it, which the fake does not.
//
// The writes are assumptions, because they were not verified on a router; they follow the web UI
// code:
//   - add_group adds a group of type 2 with a new group_id; set_group sets its name; remove_group
//     removes it with its peers.
//   - add_config adds a peer with a new peer_id and the sent fields; set_config sets the fields
//     that it gets and keeps the others; remove_config and clear_config_list remove peers. The
//     value "" for listen_port, persistent_keepalive or mtu removes that field.
//   - start sets status 1, the peer of the tunnel, and domain and port from its end_point; stop
//     sets status 0 and keeps the rest.
//   - set_setting sets the fields that it gets. A peer without a setting has
//     {"default_metric":0,"local_access":false,"masq":true}.
//   - The writes return []. An unknown group or peer gives ErrInvalidParams.
//
// It does not check the values and does not stop a write to the peer of the running tunnel.
type WireGuardClient struct {
	mu       sync.Mutex
	status   map[string]any
	groups   []map[string]any
	peers    map[int][]map[string]any
	settings map[peerKey]map[string]any
	routes   map[peerKey]*vpnRoutes
	nextID   int
}

type peerKey struct{ group, peer int }

func newWireGuardClient() *WireGuardClient {
	m := &WireGuardClient{
		peers:    map[int][]map[string]any{},
		settings: map[peerKey]map[string]any{},
		routes:   map[peerKey]*vpnRoutes{},
		nextID:   10000,
	}
	mustUnmarshal(wgClientStatus, &m.status)
	var groups struct {
		Groups []map[string]any `json:"groups"`
	}
	mustUnmarshal(wgClientGroups, &groups)
	for _, g := range groups.Groups {
		delete(g, "peer_count")
	}
	m.groups = groups.Groups
	mustUnmarshal(wgClientPeers, &m.peers)
	var settings []struct {
		Group, Peer int
		Setting     map[string]any
	}
	mustUnmarshal(wgClientSettings, &settings)
	for _, s := range settings {
		m.settings[peerKey{s.Group, s.Peer}] = s.Setting
	}
	return m
}

// Status returns the result of get_status.
func (m *WireGuardClient) Status() json.RawMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return mustMarshal(m.status)
}

// SetStatus sets the result of get_status.
func (m *WireGuardClient) SetStatus(result json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mustUnmarshal(string(result), &m.status)
}

func (m *WireGuardClient) register(s *Server) {
	handle := func(method string, f func(p map[string]any) (any, error)) {
		s.Handle("wg-client", method, func(params json.RawMessage) (any, error) {
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
	handle("get_status", func(map[string]any) (any, error) { return m.status, nil })
	handle("get_group_list", m.groupList)
	handle("get_all_config_list", m.allConfigList)
	handle("get_config_list", m.configList)
	handle("add_group", m.addGroup)
	handle("set_group", m.setGroup)
	handle("remove_group", m.removeGroup)
	handle("add_config", m.addConfig)
	handle("set_config", m.setConfig)
	handle("remove_config", m.removeConfig)
	handle("clear_config_list", m.clearConfigList)
	handle("start", m.start)
	handle("stop", func(map[string]any) (any, error) {
		m.status["status"] = 0
		return writeResult, nil
	})
	handle("get_setting", m.getSetting)
	handle("set_setting", m.setSetting)
	handle("get_route_list", func(p map[string]any) (any, error) {
		r, err := m.peerRoutes(p)
		if err != nil {
			return nil, err
		}
		return r.result(), nil
	})
	handle("add_route", m.routeWrite(func(r *vpnRoutes, p map[string]any) error { r.add(p); return nil }))
	handle("set_route", m.routeWrite((*vpnRoutes).set))
	handle("remove_route", m.routeWrite((*vpnRoutes).remove))
}

func (m *WireGuardClient) group(p map[string]any) (map[string]any, error) {
	id := toInt(p["group_id"])
	i := slices.IndexFunc(m.groups, func(g map[string]any) bool { return toInt(g["group_id"]) == id })
	if i < 0 {
		return nil, ErrInvalidParams
	}
	return m.groups[i], nil
}

func (m *WireGuardClient) peerIndex(p map[string]any) (int, int, error) {
	if _, err := m.group(p); err != nil {
		return 0, 0, err
	}
	gid, pid := toInt(p["group_id"]), toInt(p["peer_id"])
	i := slices.IndexFunc(m.peers[gid], func(peer map[string]any) bool { return toInt(peer["peer_id"]) == pid })
	if i < 0 {
		return 0, 0, ErrInvalidParams
	}
	return gid, i, nil
}

func (m *WireGuardClient) groupList(map[string]any) (any, error) {
	groups := make([]map[string]any, len(m.groups))
	for i, g := range m.groups {
		groups[i] = maps.Clone(g)
		groups[i]["peer_count"] = len(m.peers[toInt(g["group_id"])])
	}
	return map[string]any{"groups": groups}, nil
}

func (m *WireGuardClient) allConfigList(map[string]any) (any, error) {
	list := make([]map[string]any, len(m.groups))
	for i, g := range m.groups {
		item := map[string]any{"peers": nonNil(m.peers[toInt(g["group_id"])])}
		for _, k := range []string{"auth_type", "group_id", "group_name", "username", "password"} {
			if v, ok := g[k]; ok {
				item[k] = v
			}
		}
		list[i] = item
	}
	return map[string]any{"config_list": list}, nil
}

func (m *WireGuardClient) configList(p map[string]any) (any, error) {
	g, err := m.group(p)
	if err != nil {
		return nil, err
	}
	return map[string]any{"auth_type": g["auth_type"], "peers": nonNil(m.peers[toInt(g["group_id"])])}, nil
}

func (m *WireGuardClient) newID() int {
	m.nextID++
	return m.nextID
}

func (m *WireGuardClient) addGroup(p map[string]any) (any, error) {
	m.groups = append(m.groups, map[string]any{"auth_type": 0, "group_id": m.newID(), "group_name": p["group_name"], "group_type": 2})
	return writeResult, nil
}

func (m *WireGuardClient) setGroup(p map[string]any) (any, error) {
	g, err := m.group(p)
	if err != nil {
		return nil, err
	}
	g["group_name"] = p["group_name"]
	return writeResult, nil
}

func (m *WireGuardClient) removeGroup(p map[string]any) (any, error) {
	if _, err := m.group(p); err != nil {
		return nil, err
	}
	id := toInt(p["group_id"])
	m.groups = slices.DeleteFunc(m.groups, func(g map[string]any) bool { return toInt(g["group_id"]) == id })
	delete(m.peers, id)
	return writeResult, nil
}

func (m *WireGuardClient) addConfig(p map[string]any) (any, error) {
	if _, err := m.group(p); err != nil {
		return nil, err
	}
	gid := toInt(p["group_id"])
	peer := map[string]any{}
	setPeerFields(peer, p)
	delete(peer, "group_id")
	peer["peer_id"] = m.newID()
	m.peers[gid] = append(m.peers[gid], peer)
	return writeResult, nil
}

func (m *WireGuardClient) setConfig(p map[string]any) (any, error) {
	gid, i, err := m.peerIndex(p)
	if err != nil {
		return nil, err
	}
	fields := maps.Clone(p)
	delete(fields, "group_id")
	setPeerFields(m.peers[gid][i], fields)
	return writeResult, nil
}

func (m *WireGuardClient) removeConfig(p map[string]any) (any, error) {
	gid, i, err := m.peerIndex(p)
	if err != nil {
		return nil, err
	}
	m.peers[gid] = slices.Delete(m.peers[gid], i, i+1)
	return writeResult, nil
}

func (m *WireGuardClient) clearConfigList(p map[string]any) (any, error) {
	if _, err := m.group(p); err != nil {
		return nil, err
	}
	delete(m.peers, toInt(p["group_id"]))
	return writeResult, nil
}

func (m *WireGuardClient) start(p map[string]any) (any, error) {
	gid, i, err := m.peerIndex(p)
	if err != nil {
		return nil, err
	}
	peer := m.peers[gid][i]
	m.status["status"] = 1
	m.status["group_id"] = gid
	m.status["peer_id"] = peer["peer_id"]
	m.status["name"] = peer["name"]
	m.status["ipv4"] = peer["address_v4"]
	endpoint, _ := peer["end_point"].(string)
	host, port, _ := net.SplitHostPort(endpoint)
	m.status["domain"] = host
	m.status["port"], _ = strconv.Atoi(port)
	return writeResult, nil
}

func (m *WireGuardClient) key(p map[string]any) (peerKey, error) {
	if _, _, err := m.peerIndex(p); err != nil {
		return peerKey{}, err
	}
	return peerKey{toInt(p["group_id"]), toInt(p["peer_id"])}, nil
}

func (m *WireGuardClient) getSetting(p map[string]any) (any, error) {
	k, err := m.key(p)
	if err != nil {
		return nil, err
	}
	if s, ok := m.settings[k]; ok {
		return s, nil
	}
	return map[string]any{"default_metric": 0, "local_access": false, "masq": true}, nil
}

func (m *WireGuardClient) setSetting(p map[string]any) (any, error) {
	k, err := m.key(p)
	if err != nil {
		return nil, err
	}
	s, _ := m.getSetting(p)
	setting := maps.Clone(s.(map[string]any))
	maps.Copy(setting, p)
	delete(setting, "group_id")
	delete(setting, "peer_id")
	m.settings[k] = setting
	return writeResult, nil
}

func (m *WireGuardClient) peerRoutes(p map[string]any) (*vpnRoutes, error) {
	k, err := m.key(p)
	if err != nil {
		return nil, err
	}
	if m.routes[k] == nil {
		m.routes[k] = &vpnRoutes{}
	}
	return m.routes[k], nil
}

func (m *WireGuardClient) routeWrite(f func(*vpnRoutes, map[string]any) error) func(map[string]any) (any, error) {
	return func(p map[string]any) (any, error) {
		r, err := m.peerRoutes(p)
		if err != nil {
			return nil, err
		}
		if err := f(r, p); err != nil {
			return nil, err
		}
		return writeResult, nil
	}
}

func mustUnmarshal(data string, v any) {
	if err := json.Unmarshal([]byte(data), v); err != nil {
		panic(err)
	}
}
