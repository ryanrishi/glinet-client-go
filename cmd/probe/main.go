package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"
	_ "github.com/GehirnInc/crypt/sha256_crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
)

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

type fixture struct {
	Request  rpcRequest      `json:"request"`
	Response json.RawMessage `json:"response"`
}

type outcome struct {
	name    string
	fixture fixture
	raw     map[string]any
}

type probe struct {
	url     string
	sid     string
	nextID  int
	client  *http.Client
	results map[string]map[string]any
	out     []outcome
}

func main() {
	env := readEnv(".env")
	host := firstNonEmpty(os.Getenv("GLINET_HOST"), env["GLINET_HOST"], "http://192.168.8.1")
	username := firstNonEmpty(os.Getenv("GLINET_USERNAME"), env["GLINET_USERNAME"], "root")
	password := firstNonEmpty(os.Getenv("GLINET_PASSWORD"), env["GLINET_PASSWORD"])
	if password == "" {
		log.Fatal("GLINET_PASSWORD is not set in .env")
	}

	p := &probe{
		url:     strings.TrimRight(host, "/") + "/rpc",
		client:  &http.Client{Timeout: 30 * time.Second},
		results: map[string]map[string]any{},
	}
	r := newRedactor()
	r.addSecret(password, "password")

	if err := p.login(username, password, r); err != nil {
		log.Fatalf("login: %v", err)
	}
	r.addSecret(p.sid, "sid")

	for _, c := range calls {
		params, ok := c.params(p.results)
		if !ok {
			p.out = append(p.out, outcome{name: c.name(), fixture: fixture{
				Request:  rpcRequest{JSONRPC: "2.0", Method: "call", Params: []any{"<sid>", c.module, c.method, nil}},
				Response: json.RawMessage(`{"skipped":"no ID from a previous list response"}`),
			}})
			continue
		}
		p.call(c.module, c.method, params)
		if c.module == "system" && c.method == "get_info" {
			r.collect(p.results[c.name()])
		}
	}

	p.invalidSid()

	model, firmware := "unknown", "unknown"
	if info := p.results["system.get_info"]; info != nil {
		if res, ok := info["result"].(map[string]any); ok {
			model = fmt.Sprint(res["model"])
			firmware = fmt.Sprint(res["firmware_version"])
		}
	}
	dir := filepath.Join("testdata", "fixtures", sanitize(model+"_"+firmware))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}

	for _, o := range p.out {
		if o.raw != nil {
			r.collect(o.raw)
		}
	}
	for _, o := range p.out {
		resp := o.fixture.Response
		if o.raw != nil {
			resp = mustJSON(r.scrub(r.redact(o.raw)))
		}
		f := fixture{Request: o.fixture.Request, Response: resp}
		f.Request.Params = r.scrub(r.redact(roundTrip(f.Request.Params)))
		data, err := json.MarshalIndent(f, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, o.name+".json"), append(data, '\n'), 0o644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("wrote %d fixtures to %s\n", len(p.out), dir)
}

func (p *probe) post(req rpcRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Post(p.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (p *probe) id() int {
	p.nextID++
	return p.nextID
}

func (p *probe) login(username, password string, r *redactor) error {
	chReq := rpcRequest{JSONRPC: "2.0", Method: "challenge", Params: map[string]any{"username": username}, ID: p.id()}
	raw, err := p.post(chReq)
	if err != nil {
		return err
	}
	var ch struct {
		Result struct {
			Alg        int    `json:"alg"`
			Salt       string `json:"salt"`
			Nonce      string `json:"nonce"`
			HashMethod string `json:"hash_method"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &ch); err != nil {
		return fmt.Errorf("challenge: %w", err)
	}
	var chMap map[string]any
	_ = json.Unmarshal(raw, &chMap)
	if res, ok := chMap["result"].(map[string]any); ok {
		res["salt"] = "<salt>"
		res["nonce"] = "<nonce>"
	}
	p.out = append(p.out, outcome{name: "challenge", fixture: fixture{Request: chReq, Response: mustJSON(chMap)}})

	cipher, err := unixCrypt(ch.Result.Alg, password, ch.Result.Salt)
	if err != nil {
		return err
	}
	r.addSecret(cipher, "cipher")

	h := newHash(ch.Result.HashMethod)
	h.Write([]byte(username + ":" + cipher + ":" + ch.Result.Nonce))
	digest := hex.EncodeToString(h.Sum(nil))

	loginReq := rpcRequest{JSONRPC: "2.0", Method: "login", Params: map[string]any{"username": username, "hash": digest}, ID: p.id()}
	raw, err = p.post(loginReq)
	if err != nil {
		return err
	}
	var lr struct {
		Result struct {
			Sid string `json:"sid"`
		} `json:"result"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(raw, &lr); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if lr.Result.Sid == "" {
		return fmt.Errorf("login failed: %s", lr.Error)
	}
	p.sid = lr.Result.Sid
	var loginMap map[string]any
	_ = json.Unmarshal(raw, &loginMap)
	loginReq.Params = map[string]any{"username": username, "hash": "<hash>"}
	p.out = append(p.out, outcome{name: "login", fixture: fixture{Request: loginReq}, raw: loginMap})
	return nil
}

func unixCrypt(alg int, password, salt string) (string, error) {
	var c crypt.Crypt
	var prefix string
	switch alg {
	case 1:
		c, prefix = crypt.MD5, "$1$"
	case 5:
		c, prefix = crypt.SHA256, "$5$"
	case 6:
		c, prefix = crypt.SHA512, "$6$"
	default:
		return "", fmt.Errorf("unknown alg %d", alg)
	}
	return c.New().Generate([]byte(password), []byte(prefix+salt))
}

func newHash(method string) hash.Hash {
	switch strings.ToLower(method) {
	case "sha256":
		return sha256.New()
	case "sha512":
		return sha512.New()
	default:
		return md5.New()
	}
}

func (p *probe) call(module, method string, params map[string]any) {
	name := module + "." + method
	req := rpcRequest{JSONRPC: "2.0", Method: "call", Params: []any{p.sid, module, method, params}, ID: p.id()}
	raw, err := p.post(req)
	req.Params = []any{"<sid>", module, method, params}
	var m map[string]any
	if err == nil {
		err = json.Unmarshal(raw, &m)
	}
	if err != nil {
		m = map[string]any{"transport_error": err.Error()}
	}
	p.results[name] = m
	p.out = append(p.out, outcome{name: name, fixture: fixture{Request: req}, raw: m})
	fmt.Printf("%-45s %s\n", name, status(m))
}

func (p *probe) invalidSid() {
	req := rpcRequest{JSONRPC: "2.0", Method: "call", Params: []any{"invalidsid00000000000000000000000", "system", "get_info", map[string]any{}}, ID: p.id()}
	raw, err := p.post(req)
	var m map[string]any
	if err == nil {
		err = json.Unmarshal(raw, &m)
	}
	if err != nil {
		m = map[string]any{"transport_error": err.Error()}
	}
	p.out = append(p.out, outcome{name: "invalid_sid.system.get_info", fixture: fixture{Request: req}, raw: m})
}

func status(m map[string]any) string {
	if e, ok := m["error"]; ok {
		return fmt.Sprintf("error %v", e)
	}
	if t, ok := m["transport_error"]; ok {
		return fmt.Sprintf("transport error %v", t)
	}
	if res, ok := m["result"].(map[string]any); ok {
		if c, ok := res["err_code"]; ok {
			return fmt.Sprintf("ok (err_code %v %v)", c, res["err_msg"])
		}
	}
	return "ok"
}

func readEnv(path string) map[string]string {
	env := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return env
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
			v = v[1 : len(v)-1]
		}
		env[strings.TrimSpace(k)] = v
	}
	return env
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}

func sanitize(s string) string {
	return regexp.MustCompile(`[^A-Za-z0-9._-]+`).ReplaceAllString(s, "-")
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func roundTrip(v any) any {
	var out any
	_ = json.Unmarshal(mustJSON(v), &out)
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	return keys
}
