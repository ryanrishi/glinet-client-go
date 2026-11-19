package glinet

import (
	"context"
	"errors"
	"slices"
)

// WireGuardClientService calls the wg-client module of firmware 4.7.0: the WireGuard client
// configurations ("peers") in groups, the tunnel, and the settings and route rules of a peer.
// Firmware 4.9.0 has the module vpn-client instead, which this service does not call. Firmware
// 4.8.3 has both modules: wg-client still has the group and peer reads, but get_status,
// get_setting and get_route_list return -32601 (method not found); the tunnel status is in
// vpn-client get_status. The writes were not tested on 4.8.3.
//
// The router allows one WireGuard client tunnel at a time. The web UI does not let the user edit
// or remove the peer of the running tunnel, or remove its group.
type WireGuardClientService service

const wgClientModule = "wg-client"

// WireGuard client group types.
const (
	WireGuardGroupProvider = 1 // a preset of a VPN service provider; the router gets its peers from the provider
	WireGuardGroupCustom   = 2 // a group of the user ("manual" in the web UI); only these can be renamed or removed
	WireGuardGroupFromApp  = 3 // peers from the GL.iNet app; the web UI shows it only when it has peers
)

// WireGuardGroup is a group of WireGuard client peers.
//
// ID is a number that the router gives the group. It did not change in 5 days on the test router.
// AuthType, Procedure, Username and Password are for the account of a provider group; Password is
// a secret. Their values are not documented.
type WireGuardGroup struct {
	ID        int    `json:"group_id"`
	Name      string `json:"group_name"`
	Type      int    `json:"group_type"`
	PeerCount int    `json:"peer_count"`
	AuthType  int    `json:"auth_type"`
	Procedure int    `json:"procedure"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

// WireGuardPeer is a WireGuard client configuration in a group.
//
// ID is a number that the router gives the peer. It did not change in 5 days on the test router.
// PrivateKey and PresharedKey are secrets.
//
//   - AddressV4 and AddressV6 are the tunnel addresses of the router, with a prefix length, for
//     example "10.0.0.2/32". IPv6Enable enables AddressV6.
//   - AllowedIPs and DNS are lists in one string, separated by commas, for example
//     "0.0.0.0/0, ::/0".
//   - Endpoint is "host:port" of the server; PublicKey is the key of the server.
//   - ListenPort, PersistentKeepalive and MTU are 0 when not set. The router leaves some empty
//     fields out of a read, for example ListenPort or AddressV6.
type WireGuardPeer struct {
	ID                  int    `json:"peer_id"`
	Name                string `json:"name"`
	AddressV4           string `json:"address_v4"`
	AddressV6           string `json:"address_v6"`
	IPv6Enable          bool   `json:"ipv6_enable"`
	PrivateKey          string `json:"private_key"`
	PublicKey           string `json:"public_key"`
	PresharedKeyEnable  bool   `json:"presharedkey_enable"`
	PresharedKey        string `json:"preshared_key"`
	AllowedIPs          string `json:"allowed_ips"`
	Endpoint            string `json:"end_point"`
	DNS                 string `json:"dns"`
	ListenPort          int    `json:"listen_port"`
	PersistentKeepalive int    `json:"persistent_keepalive"`
	MTU                 int    `json:"mtu"`
}

// WireGuardGroupPeers is a group with its peers, an item of the result of get_all_config_list.
// Firmware 4.7.0 does not send Type, so it is 0 there.
type WireGuardGroupPeers struct {
	ID       int             `json:"group_id"`
	Name     string          `json:"group_name"`
	Type     int             `json:"group_type"`
	AuthType int             `json:"auth_type"`
	Username string          `json:"username"`
	Password string          `json:"password"`
	Peers    []WireGuardPeer `json:"peers"`
}

// WireGuardClientStatus is the status of the WireGuard client tunnel.
//
// Status is 0 stopped, 1 connected or 2 connecting (from the web UI code; 1 verified). GroupID,
// PeerID and Name are those of the peer of the tunnel. Domain and Port are the endpoint of the
// server, and IPv4 is the tunnel address of the router. The meaning of Proxy is not known (true
// on the test router while connected).
type WireGuardClientStatus struct {
	Status  int    `json:"status"`
	GroupID int    `json:"group_id"`
	PeerID  int    `json:"peer_id"`
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Port    int    `json:"port"`
	IPv4    string `json:"ipv4"`
	Proxy   bool   `json:"proxy"`
	Log     string `json:"log"`
	RxBytes int64  `json:"rx_bytes"`
	TxBytes int64  `json:"tx_bytes"`
}

// WireGuardClientSetting is the setting of a peer. LocalAccess lets the VPN server side reach the
// LAN of the router; Masq enables masquerading (NAT) in the tunnel. MTU 0 means not set; the
// router leaves it out of the read of some peers. The meaning of DefaultMetric is not known (0 on
// the test router).
type WireGuardClientSetting struct {
	DefaultMetric int  `json:"default_metric"`
	LocalAccess   bool `json:"local_access"`
	Masq          bool `json:"masq"`
	MTU           int  `json:"mtu"`
}

type wgPeerRef struct {
	GroupID int `json:"group_id"`
	PeerID  int `json:"peer_id"`
}

// GetStatus returns the status of the WireGuard client tunnel.
func (s *WireGuardClientService) GetStatus(ctx context.Context) (*WireGuardClientStatus, error) {
	var res WireGuardClientStatus
	if err := s.client.Call(ctx, wgClientModule, "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ListGroups returns the groups.
func (s *WireGuardClientService) ListGroups(ctx context.Context) ([]WireGuardGroup, error) {
	var res struct {
		Groups []WireGuardGroup `json:"groups"`
	}
	if err := s.client.Call(ctx, wgClientModule, "get_group_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Groups, nil
}

// ListAllPeers returns all groups with their peers in one call.
func (s *WireGuardClientService) ListAllPeers(ctx context.Context) ([]WireGuardGroupPeers, error) {
	var res struct {
		ConfigList []WireGuardGroupPeers `json:"config_list"`
	}
	if err := s.client.Call(ctx, wgClientModule, "get_all_config_list", nil, &res); err != nil {
		return nil, err
	}

	return res.ConfigList, nil
}

// ListPeers returns the peers of the group groupID.
func (s *WireGuardClientService) ListPeers(ctx context.Context, groupID int) ([]WireGuardPeer, error) {
	var res struct {
		Peers []WireGuardPeer `json:"peers"`
	}
	params := struct {
		GroupID int `json:"group_id"`
	}{groupID}
	if err := s.client.Call(ctx, wgClientModule, "get_config_list", params, &res); err != nil {
		return nil, err
	}

	return res.Peers, nil
}

// FindPeer returns the peer peerID of the group groupID. It returns nil and no error when there is
// no such peer.
func (s *WireGuardClientService) FindPeer(ctx context.Context, groupID, peerID int) (*WireGuardPeer, error) {
	peers, err := s.ListPeers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	for _, p := range peers {
		if p.ID == peerID {
			return &p, nil
		}
	}

	return nil, nil
}

// AddGroup adds a custom group with the name name and returns its ID. The router gives the ID; the
// web UI does not read the result of add_group, but takes the last group of the next list. So
// AddGroup reads the list before and after the write and returns the ID of the new group. Calls of
// AddGroup and AddPeer on one Client run one at a time; a write by another client between the two
// reads gives an error.
//
// Not verified on a router: the parameters come from the web UI code. The web UI adds a group
// with a default name and then renames it with set_group. It accepts names of 1-64 characters
// that do not start with "." and have no "~".
func (s *WireGuardClientService) AddGroup(ctx context.Context, name string) (int, error) {
	s.client.wgAddMu.Lock()
	defer s.client.wgAddMu.Unlock()

	before, err := s.ListGroups(ctx)
	if err != nil {
		return 0, err
	}

	params := struct {
		Name string `json:"group_name"`
	}{name}
	if err := s.client.Call(ctx, wgClientModule, "add_group", params, nil); err != nil {
		return 0, err
	}

	after, err := s.ListGroups(ctx)
	if err != nil {
		return 0, err
	}
	ids := newIDs(groupIDs(before), groupIDs(after))
	if len(ids) != 1 {
		return 0, errors.New("glinet: wg-client add_group: the group list does not have exactly one new group")
	}

	return ids[0], nil
}

// RenameGroup sets the name of the custom group groupID.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) RenameGroup(ctx context.Context, groupID int, name string) error {
	params := struct {
		GroupID int    `json:"group_id"`
		Name    string `json:"group_name"`
	}{groupID, name}
	return s.client.Call(ctx, wgClientModule, "set_group", params, nil)
}

// RemoveGroup removes the custom group groupID and its peers. The web UI does not offer it for
// the group of the running tunnel.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) RemoveGroup(ctx context.Context, groupID int) error {
	params := struct {
		GroupID int `json:"group_id"`
	}{groupID}
	return s.client.Call(ctx, wgClientModule, "remove_group", params, nil)
}

// AddPeer adds peer to the group groupID and returns the ID of the new peer. peer.ID is not sent.
// Like AddGroup, it reads the peers before and after the write, as the web UI does not read the
// result of add_config.
//
// Not verified on a router: the parameters come from the web UI code. It sends the fields of
// SetPeer without peer_id, and listen_port, persistent_keepalive and mtu only when they are not 0,
// as the web UI does for a new peer.
func (s *WireGuardClientService) AddPeer(ctx context.Context, groupID int, peer WireGuardPeer) (int, error) {
	s.client.wgAddMu.Lock()
	defer s.client.wgAddMu.Unlock()

	before, err := s.ListPeers(ctx, groupID)
	if err != nil {
		return 0, err
	}

	if err := s.client.Call(ctx, wgClientModule, "add_config", wgPeerParams(groupID, peer), nil); err != nil {
		return 0, err
	}

	after, err := s.ListPeers(ctx, groupID)
	if err != nil {
		return 0, err
	}
	ids := newIDs(peerIDs(before), peerIDs(after))
	if len(ids) != 1 {
		return 0, errors.New("glinet: wg-client add_config: the peer list does not have exactly one new peer")
	}

	return ids[0], nil
}

// SetPeer sets the configuration of the peer peer.ID in the group groupID. The web UI shows the
// peer of the running tunnel as read only, so a write to it can stop or restart the tunnel.
//
// Not verified on a router: the parameters come from the web UI code. Like the web UI, it sends
// name, address_v4, address_v6, ipv6_enable, private_key, public_key, allowed_ips, end_point and
// presharedkey_enable; preshared_key only when PresharedKeyEnable is true; dns only when it is not
// empty; and listen_port, persistent_keepalive and mtu, with "" for 0. The web UI sends "" for
// these three when the user empties the input, and does not send an input that the user did not
// fill, so SetPeer can remove a value. What the router does with "", and whether it keeps a field
// that is not sent, is not known. The web UI requires Name (1-64 characters), AddressV4, Endpoint,
// PrivateKey and PublicKey (44 characters of base64), PersistentKeepalive 1-65535 and MTU 576-65535.
func (s *WireGuardClientService) SetPeer(ctx context.Context, groupID int, peer WireGuardPeer) error {
	params := wgPeerParams(groupID, peer)
	params["peer_id"] = peer.ID
	params["listen_port"] = numberOrEmpty(peer.ListenPort)
	params["persistent_keepalive"] = numberOrEmpty(peer.PersistentKeepalive)
	params["mtu"] = numberOrEmpty(peer.MTU)
	return s.client.Call(ctx, wgClientModule, "set_config", params, nil)
}

// RemovePeer removes the peer peerID from the group groupID. The web UI does not offer it for the
// peer of the running tunnel.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) RemovePeer(ctx context.Context, groupID, peerID int) error {
	return s.client.Call(ctx, wgClientModule, "remove_config", wgPeerRef{groupID, peerID}, nil)
}

// RemoveAllPeers removes all peers of the group groupID. The web UI does not offer it for the
// group of the running tunnel.
//
// Not verified on a router: the parameters come from the web UI code. For a provider group the
// web UI also sends delete_key, which this method does not send.
func (s *WireGuardClientService) RemoveAllPeers(ctx context.Context, groupID int) error {
	params := struct {
		GroupID int `json:"group_id"`
	}{groupID}
	return s.client.Call(ctx, wgClientModule, "clear_config_list", params, nil)
}

// Start starts the tunnel with the peer peerID of the group groupID. The status is 2 (connecting)
// until the tunnel is up. The web UI does not let the user start the WireGuard client while the
// OpenVPN client runs in the proxy modes VPNProxyGlobal, VPNProxyDomain, VPNProxyMAC or
// VPNProxyVLAN.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) Start(ctx context.Context, groupID, peerID int) error {
	return s.client.Call(ctx, wgClientModule, "start", wgPeerRef{groupID, peerID}, nil)
}

// Stop stops the tunnel.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) Stop(ctx context.Context) error {
	return s.client.Call(ctx, wgClientModule, "stop", nil, nil)
}

// GetSetting returns the setting of the peer peerID of the group groupID.
func (s *WireGuardClientService) GetSetting(ctx context.Context, groupID, peerID int) (*WireGuardClientSetting, error) {
	var res WireGuardClientSetting
	if err := s.client.Call(ctx, wgClientModule, "get_setting", wgPeerRef{groupID, peerID}, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetSetting sets the setting of the peer peerID of the group groupID. The web UI changes only the
// setting of the peer of the tunnel, and sends all fields of the read with the changed values.
// It accepts an MTU of 68-65535; it sends MTU 0 when the field is empty.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) SetSetting(ctx context.Context, groupID, peerID int, setting WireGuardClientSetting) error {
	params := struct {
		wgPeerRef
		WireGuardClientSetting
	}{wgPeerRef{groupID, peerID}, setting}
	return s.client.Call(ctx, wgClientModule, "set_setting", params, nil)
}

// ListRoutes returns the route rules of the peer peerID of the group groupID.
func (s *WireGuardClientService) ListRoutes(ctx context.Context, groupID, peerID int) (*VPNRoutes, error) {
	var res VPNRoutes
	if err := s.client.Call(ctx, wgClientModule, "get_route_list", wgPeerRef{groupID, peerID}, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// AddRoute adds a route rule to the peer peerID of the group groupID. route.ID is not sent.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) AddRoute(ctx context.Context, groupID, peerID int, route VPNRoute) error {
	route.ID = 0
	return s.client.Call(ctx, wgClientModule, "add_route", wgRouteParams(groupID, peerID, route), nil)
}

// SetRoute sets the route rule route.ID of the peer peerID of the group groupID.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) SetRoute(ctx context.Context, groupID, peerID int, route VPNRoute) error {
	return s.client.Call(ctx, wgClientModule, "set_route", wgRouteParams(groupID, peerID, route), nil)
}

// RemoveRoute removes the route rule route.ID of the peer peerID of the group groupID. Only
// route.ID and route.Dest (for the address family) are used.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardClientService) RemoveRoute(ctx context.Context, groupID, peerID int, route VPNRoute) error {
	params := map[string]any{"group_id": groupID, "peer_id": peerID, "rule_id": route.ID, "route_flag": routeFlag(route.Dest)}
	return s.client.Call(ctx, wgClientModule, "remove_route", params, nil)
}

func wgRouteParams(groupID, peerID int, route VPNRoute) map[string]any {
	p := routeParams(route)
	p["group_id"] = groupID
	p["peer_id"] = peerID
	return p
}

func wgPeerParams(groupID int, peer WireGuardPeer) map[string]any {
	p := map[string]any{
		"group_id":            groupID,
		"name":                peer.Name,
		"address_v4":          peer.AddressV4,
		"address_v6":          peer.AddressV6,
		"ipv6_enable":         peer.IPv6Enable,
		"private_key":         peer.PrivateKey,
		"public_key":          peer.PublicKey,
		"allowed_ips":         peer.AllowedIPs,
		"end_point":           peer.Endpoint,
		"presharedkey_enable": peer.PresharedKeyEnable,
	}
	if peer.PresharedKeyEnable {
		p["preshared_key"] = peer.PresharedKey
	}
	if peer.DNS != "" {
		p["dns"] = peer.DNS
	}
	if peer.ListenPort != 0 {
		p["listen_port"] = peer.ListenPort
	}
	if peer.PersistentKeepalive != 0 {
		p["persistent_keepalive"] = peer.PersistentKeepalive
	}
	if peer.MTU != 0 {
		p["mtu"] = peer.MTU
	}
	return p
}

func groupIDs(groups []WireGuardGroup) []int {
	ids := make([]int, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	return ids
}

func peerIDs(peers []WireGuardPeer) []int {
	ids := make([]int, len(peers))
	for i, p := range peers {
		ids[i] = p.ID
	}
	return ids
}

func newIDs(before, after []int) []int {
	var ids []int
	for _, id := range after {
		if !slices.Contains(before, id) {
			ids = append(ids, id)
		}
	}
	return ids
}
