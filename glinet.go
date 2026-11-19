// Package glinet is a client for the JSON-RPC API of GL.iNet routers with v4 firmware.
//
// NewClient and NewClientWithHost log in to the router. Each service field of Client calls one
// module of the router, for example Client.System calls the system module.
// Client.Call calls a module method that has no typed method. The router errors are *RPCError,
// and an HTTP status that is not 200 is *HTTPError.
//
// The package glinettest has a fake router for tests.
package glinet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Version is the version of this client. It is in the default User-Agent.
const (
	Version          = "v1.0.0"
	defaultBaseURL   = "http://192.168.8.1"
	defaultUserAgent = "glinet-client-go" + "/" + Version

	// defaultTimeout is the longest wait of the web UI for a method of this client: the answer
	// of netmode set_mode (NetworkModeService.SetMode).
	defaultTimeout = 60 * time.Second
)

// Client calls the JSON-RPC API of a GL.iNet router.
type Client struct {
	// BaseURL is the address of the router, for example http://192.168.8.1.
	BaseURL *url.URL
	// UserAgent is the User-Agent header of each request.
	UserAgent string
	common    service // Reuse a single struct instead of allocating one for each service on the heap.

	httpClient *http.Client

	username string
	password []byte

	mu     sync.Mutex
	sid    string
	nextID atomic.Int64

	wgAddMu sync.Mutex

	// services
	AccessControl   *AccessControlService
	AdGuardHome     *AdGuardHomeService
	Bark            *BarkService
	Cable           *CableService
	Clients         *ClientsService
	DDNS            *DDNSService
	Digest          *DigestService
	DNS             *DNSService
	Fan             *FanService
	Firewall        *FirewallService
	IGMP            *IGMPService
	IPv6            *IPv6Service
	LAN             *LANService
	LED             *LEDService
	LocalAccess     *LocalAccessService
	MultiWAN        *MultiWANService
	Network         *NetworkService
	QoS             *QoSService
	SwitchButton    *SwitchButtonService
	NetworkMode     *NetworkModeService
	Repeater        *RepeaterService
	ParentalControl *ParentalControlService
	System          *SystemService
	Tailscale       *TailscaleService
	Tethering       *TetheringService
	Timer           *TimerService
	VPNPolicy       *VPNPolicyService
	Wifi            *WifiService
	WireGuardClient *WireGuardClientService
	WireGuardServer *WireGuardServerService
	ZeroTier        *ZeroTierService
}

type service struct {
	client *Client
}

// Option sets a value of a client that a constructor makes.
type Option func(*Client)

// WithHTTPClient sets the HTTP client that sends the requests to the router. The Timeout of
// httpClient applies, and a zero Timeout gives no time limit. Without this option, or with a nil
// httpClient, each request (also the login) has a time limit of 60 s, the longest time that the
// web UI waits for the answer of a method of this client (NetworkModeService.SetMode). The
// deadline of the context of a call also applies.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// NewClient returns a client for the default host and logs in.
func NewClient(username string, password []byte, opts ...Option) (*Client, error) {
	return NewClientWithHost(defaultBaseURL, username, password, opts...)
}

// NewClientWithHost returns a client for host and logs in.
func NewClientWithHost(host string, username string, password []byte, opts ...Option) (*Client, error) {
	c, err := newClient(host, opts)
	if err != nil {
		return nil, err
	}

	c.username = username
	c.password = bytes.Clone(password)

	if err := c.login(context.Background(), ""); err != nil {
		return nil, err
	}

	return c, nil
}

// NewClientUnauthenticated returns a client for the default host that has no session.
// Only the Digest service is usable without a session.
func NewClientUnauthenticated(opts ...Option) *Client {
	c, _ := newClient(defaultBaseURL, opts)
	return c
}

func newClient(host string, opts []Option) (*Client, error) {
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}

	baseURL, err := url.Parse(strings.TrimRight(host, "/") + "/rpc")
	if err != nil {
		return nil, fmt.Errorf("glinet: parse host %q: %w", host, err)
	}

	c := &Client{BaseURL: baseURL, UserAgent: defaultUserAgent, httpClient: &http.Client{Timeout: defaultTimeout}}
	c.common.client = c
	for _, opt := range opts {
		opt(c)
	}

	c.AccessControl = (*AccessControlService)(&c.common)
	c.AdGuardHome = (*AdGuardHomeService)(&c.common)
	c.Bark = (*BarkService)(&c.common)
	c.Cable = (*CableService)(&c.common)
	c.Clients = (*ClientsService)(&c.common)
	c.Digest = (*DigestService)(&c.common)
	c.DDNS = (*DDNSService)(&c.common)
	c.DNS = (*DNSService)(&c.common)
	c.Fan = (*FanService)(&c.common)
	c.Firewall = (*FirewallService)(&c.common)
	c.IGMP = (*IGMPService)(&c.common)
	c.IPv6 = (*IPv6Service)(&c.common)
	c.LAN = (*LANService)(&c.common)
	c.LED = (*LEDService)(&c.common)
	c.LocalAccess = (*LocalAccessService)(&c.common)
	c.MultiWAN = (*MultiWANService)(&c.common)
	c.Network = (*NetworkService)(&c.common)
	c.QoS = (*QoSService)(&c.common)
	c.SwitchButton = (*SwitchButtonService)(&c.common)
	c.NetworkMode = (*NetworkModeService)(&c.common)
	c.Repeater = (*RepeaterService)(&c.common)
	c.ParentalControl = (*ParentalControlService)(&c.common)
	c.System = (*SystemService)(&c.common)
	c.Tailscale = (*TailscaleService)(&c.common)
	c.Tethering = (*TetheringService)(&c.common)
	c.Timer = (*TimerService)(&c.common)
	c.VPNPolicy = (*VPNPolicyService)(&c.common)
	c.Wifi = (*WifiService)(&c.common)
	c.WireGuardClient = (*WireGuardClientService)(&c.common)
	c.WireGuardServer = (*WireGuardServerService)(&c.common)
	c.ZeroTier = (*ZeroTierService)(&c.common)

	return c, nil
}

// SessionID returns the current session ID.
func (c *Client) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sid
}

// Call calls method of module on the router and decodes the result into result.
// A nil params sends an empty object. A nil result discards the response.
// A result object with a non-zero err_code is an *RPCError. An HTTP status that is not 200 is an *HTTPError.
// If the session is not valid, Call logs in again one time and repeats the call.
func (c *Client) Call(ctx context.Context, module, method string, params, result any) error {
	if params == nil {
		params = struct{}{}
	}

	sid := c.SessionID()
	err := c.call(ctx, sid, module, method, params, result)

	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != CodeAccessDenied || c.username == "" {
		return err
	}

	if err := c.login(ctx, sid); err != nil {
		return err
	}

	return c.call(ctx, c.SessionID(), module, method, params, result)
}

func (c *Client) call(ctx context.Context, sid, module, method string, params, result any) error {
	err := c.rpc(ctx, "call", []any{sid, module, method, params}, result)

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		httpErr.Module, httpErr.Method = module, method
	}
	return err
}

// login gets a new session unless another caller already replaced staleSid.
func (c *Client) login(ctx context.Context, staleSid string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sid != staleSid {
		return nil
	}

	res, err := c.Digest.Login(ctx, c.username, c.password)
	if err != nil {
		return fmt.Errorf("glinet: login: %w", err)
	}

	c.sid = res.SID
	return nil
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
}

func (c *Client) rpc(ctx context.Context, method string, params, result any) error {
	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: c.nextID.Add(1), Method: method, Params: params})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return &HTTPError{StatusCode: res.StatusCode, Method: method}
	}

	var rpcRes rpcResponse
	if err := json.NewDecoder(res.Body).Decode(&rpcRes); err != nil {
		return fmt.Errorf("glinet: %s: decode response: %w", method, err)
	}

	if rpcRes.Error != nil {
		return rpcRes.Error
	}

	var resErr struct {
		Code    int    `json:"err_code"`
		Message string `json:"err_msg"`
	}
	if json.Unmarshal(rpcRes.Result, &resErr) == nil && resErr.Code != 0 {
		return &RPCError{Code: resErr.Code, Message: resErr.Message}
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(rpcRes.Result, result); err != nil {
		return fmt.Errorf("glinet: %s: decode result: %w", method, err)
	}

	return nil
}
