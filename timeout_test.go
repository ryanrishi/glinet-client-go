package glinet_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ryanrishi/glinet-client-go"
	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func newHangingServer(t *testing.T) *httptest.Server {
	t.Helper()
	stop := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-stop
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(stop) })
	return srv
}

func TestLoginHTTPClientTimeout(t *testing.T) {
	srv := newHangingServer(t)

	start := time.Now()
	_, err := glinet.NewClientWithHost(srv.URL, "root", []byte("password"),
		glinet.WithHTTPClient(&http.Client{Timeout: 50 * time.Millisecond}))
	if err == nil {
		t.Fatal("expected an error")
	}
	var urlErr *url.Error
	if !errors.As(err, &urlErr) || !urlErr.Timeout() {
		t.Errorf("error %v is not a timeout", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("login took %v", elapsed)
	}
}

func TestLoginContextDeadline(t *testing.T) {
	srv := newHangingServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := glinet.NewClientWithHost(srv.URL, "root", []byte("password"), glinet.WithLoginContext(ctx))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("login took %v", elapsed)
	}
}

func TestCallContextDeadline(t *testing.T) {
	srv := glinettest.NewServer(t)
	stop := make(chan struct{})
	t.Cleanup(func() { close(stop) })
	srv.Handle("test", "hang", func(json.RawMessage) (any, error) {
		<-stop
		return nil, nil
	})

	c, err := glinet.NewClientWithHost(srv.URL, srv.Username, []byte(srv.Password))
	if err != nil {
		t.Fatalf("NewClientWithHost: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := c.Call(ctx, "test", "hang", nil, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestCallContextDeadlineSessionRenewal(t *testing.T) {
	srv := glinettest.NewServer(t)
	target, _ := url.Parse(srv.URL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	var hangChallenge atomic.Bool
	stop := make(chan struct{})
	front := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		if hangChallenge.Load() && strings.Contains(string(body), `"method":"challenge"`) {
			<-stop
			return
		}
		proxy.ServeHTTP(w, r)
	}))
	t.Cleanup(front.Close)
	t.Cleanup(func() { close(stop) })

	c, err := glinet.NewClientWithHost(front.URL, srv.Username, []byte(srv.Password))
	if err != nil {
		t.Fatalf("NewClientWithHost: %v", err)
	}
	hangChallenge.Store(true)
	srv.ExpireSessions()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := c.System.GetTimezoneConfig(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}
