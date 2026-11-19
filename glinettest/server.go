// Package glinettest provides a fake GL.iNet v4 router for tests.
//
// The server implements the challenge, login and call JSON-RPC methods.
// Each module keeps its state in memory. Tests can read and change the state,
// replace a method with Handle, and make all sessions expire with ExpireSessions.
package glinettest

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/GehirnInc/crypt/md5_crypt"
)

// NonceTTL is how long a challenge nonce is valid for login.
const NonceTTL = 1000 * time.Millisecond

// Error is a JSON-RPC error object. A Method returns it to send that error to the client.
// Err is the detail that some router methods add, for example "invalid mac".
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     string `json:"err,omitempty"`
}

func (e *Error) Error() string {
	if e.Err != "" {
		return fmt.Sprintf("rpc error %d: %s: %s", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

// Errors that the router returns.
var (
	ErrAccessDenied   = &Error{Code: -32000, Message: "Access denied"}
	ErrMethodNotFound = &Error{Code: -32601, Message: "Method not found"}
	ErrInvalidParams  = &Error{Code: -32602, Message: "Invalid params"}
)

// ErrInternalServerError makes the server answer with HTTP status 500 and no JSON-RPC response,
// as the router does when some methods fail. A Method can return it.
var ErrInternalServerError = errors.New("internal server error")

// Method handles one module method. params is the fourth element of the call params.
// A returned *Error goes to the client as is. Other errors become -32603.
type Method func(params json.RawMessage) (any, error)

// Server is a fake router. It serves the JSON-RPC API at URL.
type Server struct {
	// URL is the base URL of the server, for example http://127.0.0.1:12345.
	URL string
	// Username is the user that can log in.
	Username string
	// Password is the password of Username.
	Password string
	// HashMethod is the hash-method of the challenge: "" (MD5, firmware 4.7.0) or "sha256".
	HashMethod string

	System          *System
	AccessControl   *AccessControl
	AdGuardHome     *AdGuardHome
	Bark            *Bark
	Cable           *Cable
	Clients         *Clients
	DDNS            *DDNS
	DNS             *DNS
	Fan             *Fan
	Firewall        *Firewall
	IGMP            *IGMP
	IPv6            *IPv6
	LAN             *LAN
	LED             *LED
	LocalAccess     *LocalAccess
	MultiWAN        *MultiWAN
	Network         *Network
	SwitchButton    *SwitchButton
	NetworkMode     *NetworkMode
	Repeater        *Repeater
	ParentalControl *ParentalControl
	Tailscale       *Tailscale
	Tethering       *Tethering
	Timer           *Timer
	VPNPolicy       *VPNPolicy
	Wifi            *Wifi
	WireGuardClient *WireGuardClient
	WireGuardServer *WireGuardServer
	ZeroTier        *ZeroTier

	srv *httptest.Server

	mu        sync.Mutex
	methods   map[string]Method
	sessions  map[string]struct{}
	salt      string
	challenge challenge
}

type challenge struct {
	nonce  string
	issued time.Time
}

// NewServer starts a fake router with the user "root" and the password "password".
// The server stops when the test ends.
func NewServer(t testing.TB) *Server {
	s := &Server{
		Username: "root",
		Password: "password",
		methods:  map[string]Method{},
		sessions: map[string]struct{}{},
		salt:     rand.Text()[:8],
	}

	s.System = newSystem()
	s.System.register(s)
	s.AccessControl = newAccessControl()
	s.AccessControl.register(s)
	s.AdGuardHome = newAdGuardHome()
	s.AdGuardHome.register(s)
	s.Bark = newBark()
	s.Bark.register(s)
	s.Cable = newCable()
	s.Cable.register(s)
	s.Clients = newClients()
	s.Clients.register(s)
	s.DDNS = newDDNS()
	s.DDNS.register(s)
	s.DNS = newDNS()
	s.DNS.register(s)
	s.Fan = newFan()
	s.Fan.register(s)
	s.Firewall = newFirewall()
	s.Firewall.register(s)
	s.IGMP = newIGMP()
	s.IGMP.register(s)
	s.IPv6 = newIPv6()
	s.IPv6.register(s)
	s.LAN = newLAN()
	s.LAN.register(s)
	s.LED = newLED()
	s.LED.register(s)
	s.LocalAccess = newLocalAccess()
	s.LocalAccess.register(s)
	s.MultiWAN = newMultiWAN()
	s.MultiWAN.register(s)
	s.Network = newNetwork()
	s.Network.register(s)
	s.SwitchButton = newSwitchButton()
	s.SwitchButton.register(s)
	s.NetworkMode = newNetworkMode()
	s.NetworkMode.register(s)
	s.Repeater = newRepeater()
	s.Repeater.register(s)
	s.ParentalControl = newParentalControl()
	s.ParentalControl.register(s)
	s.Tailscale = newTailscale()
	s.Tailscale.register(s)
	s.Tethering = newTethering()
	s.Tethering.register(s)
	s.Timer = newTimer()
	s.Timer.register(s)
	s.VPNPolicy = newVPNPolicy()
	s.VPNPolicy.register(s)
	s.Wifi = newWifi()
	s.Wifi.register(s)
	s.WireGuardClient = newWireGuardClient()
	s.WireGuardClient.register(s)
	s.WireGuardServer = newWireGuardServer()
	s.WireGuardServer.register(s)
	s.ZeroTier = newZeroTier()
	s.ZeroTier.register(s)

	s.srv = httptest.NewServer(http.HandlerFunc(s.serveHTTP))
	s.URL = s.srv.URL
	t.Cleanup(s.srv.Close)

	return s
}

// Handle sets the method that serves module.method. It replaces an existing method.
func (s *Server) Handle(module, method string, m Method) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.methods[module+"."+method] = m
}

// Handler returns the method that serves module.method, or nil if there is none.
// A test can call it from a method that it sets with Handle.
func (s *Server) Handler(module, method string) Method {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.methods[module+"."+method]
}

// ExpireSessions makes all session IDs invalid. The next call with an old session ID gets -32000.
func (s *Server) ExpireSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()
	clear(s.sessions)
}

type request struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type response struct {
	ID      json.RawMessage `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  any             `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/rpc" {
		http.NotFound(w, r)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.dispatch(req)
	if errors.Is(err, ErrInternalServerError) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := response{ID: req.ID, JSONRPC: "2.0", Result: result}
	if err != nil {
		rpcErr, ok := err.(*Error)
		if !ok {
			rpcErr = &Error{Code: -32603, Message: err.Error()}
		}
		res.Result = nil
		res.Error = rpcErr
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *Server) dispatch(req request) (any, error) {
	switch req.Method {
	case "challenge":
		return s.handleChallenge(req.Params)
	case "login":
		return s.handleLogin(req.Params)
	case "call":
		return s.handleCall(req.Params)
	default:
		return nil, ErrMethodNotFound
	}
}

func (s *Server) handleChallenge(params json.RawMessage) (any, error) {
	var p struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.challenge = challenge{nonce: rand.Text(), issued: time.Now()}

	result := map[string]any{"alg": 1, "salt": s.salt, "nonce": s.challenge.nonce}
	if s.HashMethod != "" {
		result["hash-method"] = s.HashMethod
	}
	return result, nil
}

func (s *Server) handleLogin(params json.RawMessage) (any, error) {
	var p struct {
		Username string `json:"username"`
		Hash     string `json:"hash"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, ErrInvalidParams
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.challenge
	s.challenge = challenge{}
	if c.nonce == "" || time.Since(c.issued) > NonceTTL || p.Username != s.Username {
		return nil, ErrAccessDenied
	}

	passwordHash, err := md5_crypt.New().Generate([]byte(s.Password), []byte("$1$"+s.salt))
	if err != nil {
		return nil, err
	}
	want := md5.New()
	if s.HashMethod == "sha256" {
		want = sha256.New()
	}
	want.Write([]byte(s.Username + ":" + passwordHash + ":" + c.nonce))
	if p.Hash != hex.EncodeToString(want.Sum(nil)) {
		return nil, ErrAccessDenied
	}

	sid := rand.Text()
	s.sessions[sid] = struct{}{}

	return map[string]any{"username": s.Username, "sid": sid}, nil
}

func (s *Server) handleCall(params json.RawMessage) (any, error) {
	var p []json.RawMessage
	if err := json.Unmarshal(params, &p); err != nil || len(p) < 3 {
		return nil, ErrInvalidParams
	}

	var sid, module, method string
	for i, dst := range []*string{&sid, &module, &method} {
		if err := json.Unmarshal(p[i], dst); err != nil {
			return nil, ErrInvalidParams
		}
	}

	args := json.RawMessage(`{}`)
	if len(p) > 3 {
		args = p[3]
	}

	s.mu.Lock()
	_, ok := s.sessions[sid]
	m := s.methods[module+"."+method]
	s.mu.Unlock()

	if !ok {
		return nil, ErrAccessDenied
	}
	if m == nil {
		return nil, ErrMethodNotFound
	}

	return m(args)
}
