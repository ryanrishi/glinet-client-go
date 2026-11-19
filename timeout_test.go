package glinet_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ryanrishi/glinet-client-go"
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

func TestCallContextDeadline(t *testing.T) {
	srv := newHangingServer(t)
	c := glinet.NewClientUnauthenticated()
	c.BaseURL, _ = url.Parse(srv.URL + "/rpc")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Digest.Challenge(ctx, "root")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}
