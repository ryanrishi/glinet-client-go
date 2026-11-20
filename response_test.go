package glinet_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

// newRawClient returns a client without a session whose requests get body as the response.
func newRawClient(t *testing.T, body string) *glinet.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)

	c := glinet.NewClientUnauthenticated()
	c.BaseURL, _ = url.Parse(srv.URL + "/rpc")
	return c
}

func TestCallResponseShapes(t *testing.T) {
	success := []string{
		`{"id":1,"jsonrpc":"2.0","result":[]}`,
		`{"id":1,"jsonrpc":"2.0","result":{}}`,
		`{"id":1,"jsonrpc":"2.0","result":null}`,
		`{"id":1,"jsonrpc":"2.0","result":{"enabled":true}}`,
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":0,"err_msg":"ok"}}`,
	}
	for _, body := range success {
		if err := newRawClient(t, body).Call(context.Background(), "test", "method", nil, nil); err != nil {
			t.Errorf("%s: error = %v, want nil", body, err)
		}
	}

	rpcErrors := map[string]glinet.RPCError{
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":-1,"err_msg":"parameter missing"}}`:         {Code: -1, Message: "parameter missing"},
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":20002001,"err_msg":"Parameter not found"}}`: {Code: 20002001, Message: "Parameter not found"},
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":-1}}`:                                       {Code: -1},
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":-1,"err_msg":null}}`:                        {Code: -1},
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":-1,"err_msg":{}}}`:                          {Code: -1, Message: "{}"},
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":-1,"err_msg":["a"]}}`:                       {Code: -1, Message: `["a"]`},
		`{"id":1,"jsonrpc":"2.0","result":{"err_code":1,"err_msg":5}}`:                            {Code: 1, Message: "5"},
		`{"id":1,"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params"}}`:             {Code: -32602, Message: "Invalid params"},
	}
	for body, want := range rpcErrors {
		err := newRawClient(t, body).Call(context.Background(), "test", "method", nil, nil)
		var rpcErr *glinet.RPCError
		if !errors.As(err, &rpcErr) || *rpcErr != want {
			t.Errorf("%s: error = %#v, want *RPCError %+v", body, err, want)
		}
	}

	noResult := []string{
		`{"id":1,"jsonrpc":"2.0"}`,
		`{}`,
		`null`,
		`{"id":1,"jsonrpc":"2.0","error":null}`,
	}
	for _, body := range noResult {
		err := newRawClient(t, body).Call(context.Background(), "test", "method", nil, nil)
		var rpcErr *glinet.RPCError
		if err == nil || errors.As(err, &rpcErr) || !strings.Contains(err.Error(), "no result and no error") {
			t.Errorf("%s: error = %v, want no result and no error", body, err)
		}
	}
}

// TestCallRouterResponses sends each recorded router response to Call. Only the responses with an
// error or a non-zero err_code are errors.
func TestCallRouterResponses(t *testing.T) {
	files, err := filepath.Glob("testdata/fixtures/*/*.json")
	if err != nil {
		t.Fatal(err)
	}

	responses := 0
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var f struct {
			Response map[string]json.RawMessage `json:"response"`
		}
		if err := json.Unmarshal(data, &f); err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		_, hasResult := f.Response["result"]
		_, hasError := f.Response["error"]
		if !hasResult && !hasError {
			continue
		}
		responses++

		var resErr struct {
			Code int `json:"err_code"`
		}
		_ = json.Unmarshal(f.Response["result"], &resErr)
		wantErr := hasError || resErr.Code != 0

		body, _ := json.Marshal(f.Response)
		err = newRawClient(t, string(body)).Call(context.Background(), "test", "method", nil, nil)
		var rpcErr *glinet.RPCError
		if wantErr && !errors.As(err, &rpcErr) {
			t.Errorf("%s: error = %v, want an *RPCError", file, err)
		}
		if !wantErr && err != nil {
			t.Errorf("%s: error = %v, want nil", file, err)
		}
	}
	if responses == 0 {
		t.Fatal("no router responses found")
	}
}

func TestLoginWithoutSessionID(t *testing.T) {
	for name, result := range map[string]string{
		"empty object": `{}`,
		"empty sid":    `{"username":"root","sid":""}`,
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if strings.Contains(string(body), `"method":"challenge"`) {
					_, _ = io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":{"alg":1,"salt":"abcdefgh","nonce":"n"}}`)
					return
				}
				_, _ = io.WriteString(w, `{"jsonrpc":"2.0","id":2,"result":`+result+`}`)
			}))
			t.Cleanup(srv.Close)

			_, err := glinet.NewClientWithHost(srv.URL, "root", []byte("password"))
			if err == nil || !strings.Contains(err.Error(), "no session ID") {
				t.Errorf("error = %v, want no session ID", err)
			}
		})
	}
}

func TestServerNilResult(t *testing.T) {
	c, srv := newClient(t)
	srv.Handle("test", "write", func(json.RawMessage) (any, error) { return nil, nil })

	var res json.RawMessage
	if err := c.Call(context.Background(), "test", "write", nil, &res); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if string(res) != `[]` {
		t.Errorf("result = %s, want []", res)
	}
}

func TestRedirectToSameHost(t *testing.T) {
	router := glinettest.NewServer(t)
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, router.URL+"/rpc", http.StatusTemporaryRedirect)
	}))
	t.Cleanup(front.Close)

	c, err := glinet.NewClientWithHost(front.URL, router.Username, []byte(router.Password))
	if err != nil {
		t.Fatalf("NewClientWithHost: %v", err)
	}
	if _, err := c.System.GetTimezoneConfig(context.Background()); err != nil {
		t.Errorf("GetTimezoneConfig: %v", err)
	}
}

func TestRedirectToOtherHostRefused(t *testing.T) {
	var reached atomic.Bool
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached.Store(true)
	}))
	t.Cleanup(other.Close)
	otherURL, _ := url.Parse(other.URL)
	otherURL.Host = "localhost:" + otherURL.Port()

	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, otherURL.String()+"/rpc", http.StatusTemporaryRedirect)
	}))
	t.Cleanup(front.Close)

	_, err := glinet.NewClientWithHost(front.URL, "root", []byte("password"))
	if err == nil || !strings.Contains(err.Error(), "redirect to another host refused") {
		t.Errorf("error = %v, want redirect to another host refused", err)
	}
	if reached.Load() {
		t.Error("the request reached the other host")
	}
}
