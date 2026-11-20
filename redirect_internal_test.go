package glinet

import (
	"net/http"
	"testing"
)

func TestCheckRedirect(t *testing.T) {
	request := func(rawURL string) *http.Request {
		req, err := http.NewRequest(http.MethodPost, rawURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		return req
	}

	tests := []struct {
		name    string
		via     []string
		to      string
		allowed bool
	}{
		{"HTTP to HTTPS on another port", []string{"http://192.0.2.1/rpc"}, "https://192.0.2.1:8443/rpc", true},
		{"same host other letter case", []string{"http://Router.example/rpc"}, "https://router.example/rpc", true},
		{"another host", []string{"http://192.0.2.1/rpc"}, "http://198.51.100.1/rpc", false},
		{"HTTPS to HTTP", []string{"https://192.0.2.1/rpc"}, "http://192.0.2.1/rpc", false},
		{"HTTP to HTTPS to HTTP", []string{"http://192.0.2.1/rpc", "https://192.0.2.1/rpc"}, "http://192.0.2.1/rpc", false},
		{"10 redirects", []string{"http://192.0.2.1/rpc", "http://192.0.2.1/1", "http://192.0.2.1/2", "http://192.0.2.1/3", "http://192.0.2.1/4", "http://192.0.2.1/5", "http://192.0.2.1/6", "http://192.0.2.1/7", "http://192.0.2.1/8", "http://192.0.2.1/9"}, "http://192.0.2.1/10", false},
	}
	for _, tt := range tests {
		var via []*http.Request
		for _, u := range tt.via {
			via = append(via, request(u))
		}
		err := checkRedirect(request(tt.to), via)
		if (err == nil) != tt.allowed {
			t.Errorf("%s: error = %v, want allowed %v", tt.name, err, tt.allowed)
		}
	}
}
