package glinet

import (
	"testing"
	"time"

	"github.com/ryanrishi/glinet-client-go/glinettest"
)

func TestDefaultHTTPClientTimeout(t *testing.T) {
	srv := glinettest.NewServer(t)
	newWithHost := func(opts ...Option) *Client {
		c, err := NewClientWithHost(srv.URL, srv.Username, []byte(srv.Password), opts...)
		if err != nil {
			t.Fatalf("NewClientWithHost: %v", err)
		}
		return c
	}

	tests := map[string]*Client{
		"NewClientWithHost":                            newWithHost(),
		"NewClientWithHost WithHTTPClient(nil)":        newWithHost(WithHTTPClient(nil)),
		"NewClientUnauthenticated":                     NewClientUnauthenticated(),
		"NewClientUnauthenticated WithHTTPClient(nil)": NewClientUnauthenticated(WithHTTPClient(nil)),
	}
	for name, c := range tests {
		if got, want := c.httpClient.Timeout, 60*time.Second; got != want {
			t.Errorf("%s: Timeout = %v, want %v", name, got, want)
		}
	}
}
