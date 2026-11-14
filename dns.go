package glinet

import "context"

// DNSService calls the dns module.
type DNSService service

// GetHosts returns the hosts file of the router, for example
// "127.0.0.1 localhost\n\n::1     localhost ip6-localhost ip6-loopback\n".
func (s *DNSService) GetHosts(ctx context.Context) (string, error) {
	var res struct {
		Content string `json:"content"`
	}
	if err := s.client.Call(ctx, "dns", "get_host", nil, &res); err != nil {
		return "", err
	}

	return res.Content, nil
}

// SetHosts replaces the hosts file with content. Keep the localhost lines of GetHosts in it.
// The router does not check the lines and keeps content as it is, with comments, blank lines and
// "\r\n", but adds "\n" when content does not end with one.
// In the test, the DNS server of the router did not answer with a new name within 30 s.
func (s *DNSService) SetHosts(ctx context.Context, content string) error {
	params := struct {
		Content string `json:"content"`
	}{Content: content}
	return s.client.Call(ctx, "dns", "set_host", params, nil)
}
