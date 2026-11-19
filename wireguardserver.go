package glinet

import "context"

// WireGuardServerService calls the wg-server module: the WireGuard server of the router and its
// peers (the clients that connect to it).
type WireGuardServerService service

const wgServerModule = "wg-server"

// WireGuardServerConfig is the configuration of the WireGuard server.
//
// Initialization is false until the server has keys; then the web UI shows only an "initialize"
// button, which calls Initialize. AddressV4 and AddressV6 are the tunnel addresses of the server
// with a prefix length. Port is the UDP listen port. PrivateKey is a secret; PrivateKey and
// PublicKey are "" before the initialization.
type WireGuardServerConfig struct {
	Initialization bool   `json:"initialization"`
	AddressV4      string `json:"address_v4"`
	AddressV6      string `json:"address_v6"`
	IPv6Enable     bool   `json:"ipv6_enable"`
	LocalAccess    bool   `json:"local_access"`
	Port           int    `json:"port"`
	PrivateKey     string `json:"private_key"`
	PublicKey      string `json:"public_key"`
}

// WireGuardServerStatus is the status of the WireGuard server. Server.Status is 0 stopped,
// 1 running or 2 starting (from the web UI code; 0 verified).
//
// Not verified on a router: the items of Peers, as the test router had no peers.
type WireGuardServerStatus struct {
	Server WireGuardServerState `json:"server"`
	Peers  []map[string]any     `json:"peers"`
}

// WireGuardServerState is the server part of WireGuardServerStatus.
type WireGuardServerState struct {
	Status         int    `json:"status"`
	Initialization bool   `json:"initialization"`
	TunnelIP       string `json:"tunnel_ip"`
	Log            string `json:"log"`
	RxBytes        int64  `json:"rx_bytes"`
	TxBytes        int64  `json:"tx_bytes"`
}

// WireGuardServerPeer is a peer of the WireGuard server: a client that can connect.
//
// Not verified on a router: the test router had no peers, so the fields come from the web UI code.
// ID is the number that add_peer returns. ClientIP is the tunnel address of the client, which the
// router gives. Deprecated 1 marks a peer that the web UI shows struck through and only lets the
// user remove. PresharedKey is a secret; the web UI generates it when it is empty and
// PresharedKeyEnable is true. AllowedIPs has at most 5 items in the web UI.
type WireGuardServerPeer struct {
	ID                  int64  `json:"peer_id"`
	Name                string `json:"name"`
	ClientIP            string `json:"client_ip"`
	Deprecated          int    `json:"deprecated"`
	AllowedIPs          string `json:"allowed_ips"`
	DNS                 string `json:"dns"`
	PresharedKeyEnable  bool   `json:"presharedkey_enable"`
	PresharedKey        string `json:"presharedkey"`
	PersistentKeepalive int    `json:"persistent_keepalive"`
	MTU                 int    `json:"mtu"`
}

// WireGuardServerPeerConfig is the configuration file of a peer for the client device, the result
// of generate_peer. The web UI shows it as text and as a QR code. PrivateKey and PresharedKey are
// secrets. DomainEndpoint uses the DDNS name of the router and IPEndpoint its WAN IP.
//
// Not verified on a router: the fields come from the web UI code.
type WireGuardServerPeerConfig struct {
	Address             string `json:"address"`
	AllowedIPs          string `json:"allowed_ips"`
	DNS                 string `json:"dns"`
	DomainEndpoint      string `json:"domain_end_point"`
	IPEndpoint          string `json:"ip_end_point"`
	ListenPort          int    `json:"listen_port"`
	PersistentKeepalive int    `json:"persistent_keepalive"`
	PrivateKey          string `json:"private_key"`
	PublicKey           string `json:"public_key"`
	PresharedKey        string `json:"presharedkey"`
	MTU                 int    `json:"mtu"`
}

// WireGuardServerSetting is the setting of the WireGuard server. ClientToClient lets the peers
// reach each other. See WireGuardClientSetting for the other fields.
type WireGuardServerSetting struct {
	ClientToClient bool `json:"client_to_client"`
	LocalAccess    bool `json:"local_access"`
	Masq           bool `json:"masq"`
	MTU            int  `json:"mtu"`
}

// GetConfig returns the configuration of the WireGuard server.
func (s *WireGuardServerService) GetConfig(ctx context.Context) (*WireGuardServerConfig, error) {
	var res WireGuardServerConfig
	if err := s.client.Call(ctx, wgServerModule, "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Initialize makes the keys of the server (generate_key). The web UI calls it once, while
// GetConfig has Initialization false.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) Initialize(ctx context.Context) error {
	return s.client.Call(ctx, wgServerModule, "generate_key", nil, nil)
}

// SetConfig sets the configuration of the WireGuard server. It sends address_v4, address_v6 and
// port; private_key only when it is not empty; and ipv6_enable only when AddressV6 is not empty.
// Initialization, LocalAccess and PublicKey are not sent. An empty AddressV6 is sent as "".
//
// The web UI of firmware 4.8.3 sends address_v4, address_v6, port and private_key, each also when
// it is "", and no ipv6_enable. Its private key input shows "Auto Generate" when it is empty; what
// the router does with private_key "" is not known, so SetConfig does not send it. Firmware 4.7.0
// sent private_key only when it was not empty, and ipv6_enable and address_v6 only while IPv6 was
// enabled on the router.
//
// The web UI lets the user change the configuration only while the server is stopped. It accepts
// a private IPv4 address with a prefix length up to 30 (default /24), an IPv6 prefix length up to
// 126 (default /64) and a port 1-65535. It shows the error 20001024 as an invalid private key.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) SetConfig(ctx context.Context, config WireGuardServerConfig) error {
	params := map[string]any{"address_v4": config.AddressV4, "address_v6": config.AddressV6, "port": config.Port}
	if config.PrivateKey != "" {
		params["private_key"] = config.PrivateKey
	}
	if config.AddressV6 != "" {
		params["ipv6_enable"] = config.IPv6Enable
	}
	return s.client.Call(ctx, wgServerModule, "set_config", params, nil)
}

// GetStatus returns the status of the WireGuard server.
func (s *WireGuardServerService) GetStatus(ctx context.Context) (*WireGuardServerStatus, error) {
	var res WireGuardServerStatus
	if err := s.client.Call(ctx, wgServerModule, "get_status", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ListPeers returns the peers of the WireGuard server.
func (s *WireGuardServerService) ListPeers(ctx context.Context) ([]WireGuardServerPeer, error) {
	var res struct {
		Peers []WireGuardServerPeer `json:"peers"`
	}
	if err := s.client.Call(ctx, wgServerModule, "get_peer_list", nil, &res); err != nil {
		return nil, err
	}

	return res.Peers, nil
}

// AddPeer adds a peer and returns its ID, which the router returns as peer_id. peer.ID,
// peer.ClientIP and peer.Deprecated are not sent. The router gives the client address and the
// keys.
//
// Not verified on a router: the parameters come from the web UI code. It sends name,
// presharedkey_enable, presharedkey, persistent_keepalive and mtu, with "" for a
// PersistentKeepalive or MTU of 0, as the web UI does; and allowed_ips and dns only when they are
// not empty, as the web UI does. What the router does with "" is not known. The web UI requires
// Name (1-64 characters) and accepts PersistentKeepalive 1-65535 and MTU 576-65535.
func (s *WireGuardServerService) AddPeer(ctx context.Context, peer WireGuardServerPeer) (int64, error) {
	var res struct {
		PeerID int64 `json:"peer_id"`
	}
	if err := s.client.Call(ctx, wgServerModule, "add_peer", wgServerPeerParams(peer), &res); err != nil {
		return 0, err
	}

	return res.PeerID, nil
}

// SetPeer sets the peer peer.ID. It sends the fields of AddPeer and peer_id.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) SetPeer(ctx context.Context, peer WireGuardServerPeer) error {
	params := wgServerPeerParams(peer)
	params["peer_id"] = peer.ID
	return s.client.Call(ctx, wgServerModule, "set_peer", params, nil)
}

// RemovePeer removes the peer peerID.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) RemovePeer(ctx context.Context, peerID int64) error {
	params := struct {
		PeerID int64 `json:"peer_id"`
	}{peerID}
	return s.client.Call(ctx, wgServerModule, "remove_peer", params, nil)
}

// RemoveAllPeers removes every peer with wg-server remove_peer {"all": true}, as the "remove all"
// button of the web UI of firmware 4.8.3 does.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) RemoveAllPeers(ctx context.Context) error {
	return s.client.Call(ctx, wgServerModule, "remove_peer", removeAllParams, nil)
}

// GetPeerConfig returns the configuration file of the peer peerID for the client device
// (generate_peer). The web UI calls it after each add and set of a peer and to export a peer.
//
// Not verified on a router: the parameters come from the web UI code. Whether generate_peer
// makes new keys is not known.
func (s *WireGuardServerService) GetPeerConfig(ctx context.Context, peerID int64) (*WireGuardServerPeerConfig, error) {
	params := struct {
		PeerID int64 `json:"peer_id"`
	}{peerID}
	var res WireGuardServerPeerConfig
	if err := s.client.Call(ctx, wgServerModule, "generate_peer", params, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Start starts the WireGuard server.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) Start(ctx context.Context) error {
	return s.client.Call(ctx, wgServerModule, "start", nil, nil)
}

// Stop stops the WireGuard server.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) Stop(ctx context.Context) error {
	return s.client.Call(ctx, wgServerModule, "stop", nil, nil)
}

// GetSetting returns the setting of the WireGuard server.
func (s *WireGuardServerService) GetSetting(ctx context.Context) (*WireGuardServerSetting, error) {
	var res WireGuardServerSetting
	if err := s.client.Call(ctx, wgServerModule, "get_setting", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetSetting sets the setting of the WireGuard server. It sends all fields; the web UI sends MTU
// 0 when the field is empty.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) SetSetting(ctx context.Context, setting WireGuardServerSetting) error {
	return s.client.Call(ctx, wgServerModule, "set_setting", setting, nil)
}

// ListRoutes returns the route rules of the WireGuard server.
func (s *WireGuardServerService) ListRoutes(ctx context.Context) (*VPNRoutes, error) {
	var res VPNRoutes
	if err := s.client.Call(ctx, wgServerModule, "get_route_list", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// AddRoute adds a route rule to the WireGuard server. route.ID is not sent. It sends dest,
// gateway, metric, mtu and route_flag, also an empty gateway and a metric or mtu of 0, as the web
// UI of firmware 4.8.3 does, and scope only when it is not empty. That web UI does not send scope.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) AddRoute(ctx context.Context, route VPNRoute) error {
	route.ID = 0
	return s.client.Call(ctx, wgServerModule, "add_route", wgServerRouteParams(route), nil)
}

// SetRoute sets the route rule route.ID of the WireGuard server. It sends the fields of AddRoute
// and rule_id, so an empty Gateway, Metric or MTU removes that value.
//
// Not verified on a router: the parameters come from the web UI code. What the router does with
// gateway "" and metric or mtu 0 is not known.
func (s *WireGuardServerService) SetRoute(ctx context.Context, route VPNRoute) error {
	return s.client.Call(ctx, wgServerModule, "set_route", wgServerRouteParams(route), nil)
}

func wgServerRouteParams(route VPNRoute) map[string]any {
	p := routeParams(route)
	p["gateway"] = route.Gateway
	p["metric"] = route.Metric
	p["mtu"] = route.MTU
	return p
}

// RemoveRoute removes the route rule route.ID of the WireGuard server. Only route.ID and
// route.Dest (for the address family) are used.
//
// Not verified on a router: the parameters come from the web UI code.
func (s *WireGuardServerService) RemoveRoute(ctx context.Context, route VPNRoute) error {
	params := map[string]any{"rule_id": route.ID, "route_flag": routeFlag(route.Dest)}
	return s.client.Call(ctx, wgServerModule, "remove_route", params, nil)
}

func wgServerPeerParams(peer WireGuardServerPeer) map[string]any {
	p := map[string]any{
		"name":                 peer.Name,
		"presharedkey_enable":  peer.PresharedKeyEnable,
		"presharedkey":         peer.PresharedKey,
		"persistent_keepalive": numberOrEmpty(peer.PersistentKeepalive),
		"mtu":                  numberOrEmpty(peer.MTU),
	}
	if peer.AllowedIPs != "" {
		p["allowed_ips"] = peer.AllowedIPs
	}
	if peer.DNS != "" {
		p["dns"] = peer.DNS
	}
	return p
}

// numberOrEmpty returns n, or "" for 0: the value that the web UI sends for an empty number input.
func numberOrEmpty(n int) any {
	if n == 0 {
		return ""
	}
	return n
}
