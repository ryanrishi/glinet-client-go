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
)

const (
	Version          = "v0.1.0"
	defaultBaseUrl   = "http://192.168.8.1"
	defaultUserAgent = "glinet-client-go" + "/" + Version
)

type Client struct {
	BaseURL   *url.URL
	UserAgent string
	common    service // Reuse a single struct instead of allocating one for each service on the heap.

	username string
	password []byte

	mu     sync.Mutex
	sid    string
	nextID atomic.Int64

	// services
	AdGuard *AdGuardService
	Digest  *DigestService
	LAN     *LANService
	System  *SystemService
}

type service struct {
	client *Client
}

func NewClient(username string, password []byte) (*Client, error) {
	return NewClientWithHost(defaultBaseUrl, username, password)
}

func NewClientWithHost(host string, username string, password []byte) (*Client, error) {
	c, err := newClient(host)
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
func NewClientUnauthenticated() *Client {
	c, _ := newClient(defaultBaseUrl)
	return c
}

func newClient(host string) (*Client, error) {
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}

	baseUrl, err := url.Parse(strings.TrimRight(host, "/") + "/rpc")
	if err != nil {
		return nil, fmt.Errorf("glinet: parse host %q: %w", host, err)
	}

	c := &Client{BaseURL: baseUrl, UserAgent: defaultUserAgent}
	c.common.client = c

	c.AdGuard = (*AdGuardService)(&c.common)
	c.Digest = (*DigestService)(&c.common)
	c.LAN = (*LANService)(&c.common)
	c.System = (*SystemService)(&c.common)

	return c, nil
}

func (c *Client) GetSid() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sid
}

// Call calls method of module on the router and decodes the result into result.
// A nil params sends an empty object. A nil result discards the response.
// If the session is not valid, Call logs in again one time and repeats the call.
func (c *Client) Call(ctx context.Context, module, method string, params, result any) error {
	if params == nil {
		params = struct{}{}
	}

	sid := c.GetSid()
	err := c.rpc(ctx, "call", []any{sid, module, method, params}, result)

	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != CodeAccessDenied || c.username == "" {
		return err
	}

	if err := c.login(ctx, sid); err != nil {
		return err
	}

	return c.rpc(ctx, "call", []any{c.GetSid(), module, method, params}, result)
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

	c.sid = res.Sid
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

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("glinet: %s: unexpected HTTP status %s", method, res.Status)
	}

	var rpcRes rpcResponse
	if err := json.NewDecoder(res.Body).Decode(&rpcRes); err != nil {
		return fmt.Errorf("glinet: %s: decode response: %w", method, err)
	}

	if rpcRes.Error != nil {
		return rpcRes.Error
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(rpcRes.Result, result); err != nil {
		return fmt.Errorf("glinet: %s: decode result: %w", method, err)
	}

	return nil
}
