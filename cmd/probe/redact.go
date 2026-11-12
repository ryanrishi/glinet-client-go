package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"regexp"
	"strings"
)

var (
	sensitiveKey = regexp.MustCompile(`^(key|key\d+|.*_key\d*|.*password.*|passwd|.*psk.*|.*private.*|.*preshared.*|.*token.*|.*secret.*|sid|sn|sn_bak|.*serial.*|device_?id|.*ddns.*|askpass|ta)$`)
	publicKey    = regexp.MustCompile(`public_?key$`)
	hostKey      = regexp.MustCompile(`^(end_?point|domain)$`)
	hostnameKey  = regexp.MustCompile(`^(hostname|alias)$`)
	macPattern   = regexp.MustCompile(`\b[0-9A-Fa-f]{2}([:-])[0-9A-Fa-f]{2}(?:[:-][0-9A-Fa-f]{2}){4}\b`)
	ipv4Pattern  = regexp.MustCompile(`(^|[^0-9.])((?:\d{1,3}\.){3}\d{1,3})($|[^0-9.])`)
	ipv6Pattern  = regexp.MustCompile(`[0-9A-Fa-f]{0,4}(?::[0-9A-Fa-f]{0,4}){2,7}`)
	publicDNS    = map[string]bool{
		"1.1.1.1": true, "1.0.0.1": true, "8.8.8.8": true, "8.8.4.4": true,
		"9.9.9.9": true, "149.112.112.112": true, "208.67.222.222": true, "208.67.220.220": true,
		"2606:4700:4700::1111": true, "2606:4700:4700::1001": true,
		"2001:4860:4860::8888": true, "2001:4860:4860::8844": true,
	}
	_, cgnat, _ = net.ParseCIDR("100.64.0.0/10")
)

type redactor struct {
	secrets map[string]string
	exact   map[string]string
	fakes   map[string]string
	counter map[string]int
}

func newRedactor() *redactor {
	return &redactor{secrets: map[string]string{}, exact: map[string]string{}, fakes: map[string]string{}, counter: map[string]int{}}
}

func (r *redactor) next(label string) int {
	r.counter[label]++
	return r.counter[label]
}

func (r *redactor) addSecret(value, label string) {
	if len(value) < 4 {
		return
	}
	if _, ok := r.secrets[value]; ok {
		return
	}
	r.secrets[value] = fmt.Sprintf("redacted-%s-%d", label, r.next(label))
}

func (r *redactor) addHost(s string) {
	host := s
	if h, _, err := net.SplitHostPort(s); err == nil {
		host = h
	}
	if host == "" || net.ParseIP(host) != nil || !strings.Contains(host, ".") {
		return
	}
	if _, ok := r.secrets[host]; !ok {
		r.secrets[host] = fmt.Sprintf("host-%d.example.com", r.next("host"))
	}
}

// addValue maps a whole string value to a stable fake value.
// Values with 4 or more characters are also replaced inside other strings, such as log lines.
func (r *redactor) addValue(value, label string, fake func(n int) string) {
	if value == "" {
		return
	}
	if _, ok := r.exact[value]; ok {
		return
	}
	f := fake(r.next(label))
	r.exact[value] = f
	if len(value) >= 4 {
		r.secrets[value] = f
	}
}

func fakeName(label string) func(n int) string {
	return func(n int) string { return fmt.Sprintf("%s-%d", label, n) }
}

func fakePublicKey(n int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("fake-wireguard-public-key-%d", n)))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func isSensitive(key string) bool {
	k := strings.ToLower(key)
	return sensitiveKey.MatchString(k) && !publicKey.MatchString(k)
}

func (r *redactor) collect(v any) {
	switch t := v.(type) {
	case map[string]any:
		_, hasMAC := t["mac"]
		for k, val := range t {
			s, ok := val.(string)
			lk := strings.ToLower(k)
			if ok && lk == "ssid" {
				r.addValue(s, "ssid", fakeName("ssid"))
			}
			if ok && (hostnameKey.MatchString(lk) || (lk == "name" && hasMAC)) {
				r.addValue(s, "hostname", fakeName("hostname"))
			}
			if ok && publicKey.MatchString(lk) {
				r.addValue(s, "public-key", fakePublicKey)
			}
			if ok && isSensitive(k) {
				r.addSecret(s, strings.ToLower(k))
			}
			if ok && hostKey.MatchString(strings.ToLower(k)) {
				r.addHost(s)
			}
			r.collect(val)
		}
	case []any:
		for _, val := range t {
			r.collect(val)
		}
	}
}

func (r *redactor) redact(v any) any {
	return r.walk("", v)
}

func (r *redactor) walk(key string, v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = r.walk(k, val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = r.walk(key, val)
		}
		return out
	case string:
		if t == "" {
			return t
		}
		if isSensitive(key) {
			r.addSecret(t, strings.ToLower(key))
			if fake, ok := r.secrets[t]; ok {
				return fake
			}
			return "redacted"
		}
		if fake, ok := r.exact[t]; ok {
			return fake
		}
		return r.redactString(t)
	default:
		return v
	}
}

func (r *redactor) redactString(s string) string {
	if strings.Contains(s, "PRIVATE KEY") {
		return "redacted-private-key-pem"
	}
	for _, real := range sortedKeys(r.secrets) {
		s = strings.ReplaceAll(s, real, r.secrets[real])
	}
	s = macPattern.ReplaceAllStringFunc(s, r.fakeMAC)
	s = ipv4Pattern.ReplaceAllStringFunc(s, func(m string) string {
		sub := ipv4Pattern.FindStringSubmatch(m)
		return sub[1] + r.fakeIPv4(sub[2]) + sub[3]
	})
	s = ipv6Pattern.ReplaceAllStringFunc(s, r.fakeIPv6)
	return s
}

func (r *redactor) redactText(b []byte) []byte {
	for _, real := range sortedKeys(r.secrets) {
		b = bytes.ReplaceAll(b, []byte(real), []byte(r.secrets[real]))
	}
	return b
}

func (r *redactor) mapped(real, label string, fake func(n int) string) string {
	key := label + ":" + strings.ToLower(real)
	if f, ok := r.fakes[key]; ok {
		return f
	}
	f := fake(r.next(label))
	r.fakes[key] = f
	return f
}

func (r *redactor) fakeMAC(m string) string {
	if strings.EqualFold(m, "00:00:00:00:00:00") || strings.EqualFold(m, "ff:ff:ff:ff:ff:ff") {
		return m
	}
	sep := string(m[2])
	f := r.mapped(m, "mac", func(n int) string {
		return fmt.Sprintf("02:00:00:00:%02x:%02x", n/256, n%256)
	})
	f = strings.ReplaceAll(f, ":", sep)
	if m != strings.ToLower(m) {
		f = strings.ToUpper(f)
	}
	return f
}

func (r *redactor) fakeIPv4(s string) string {
	ip := net.ParseIP(s).To4()
	if ip == nil || publicDNS[s] || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() ||
		ip.IsMulticast() || ip.IsUnspecified() || ip.Equal(net.IPv4bcast) || cgnat.Contains(ip) || ip[0] == 255 {
		return s
	}
	return r.mapped(s, "ipv4", func(n int) string { return fmt.Sprintf("203.0.113.%d", n) })
}

func (r *redactor) fakeIPv6(s string) string {
	ip := net.ParseIP(s)
	if ip == nil || ip.To4() != nil || strings.Count(s, ":") < 2 || publicDNS[s] ||
		ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() {
		return s
	}
	switch {
	case ip.IsLinkLocalUnicast():
		return r.mapped(s, "ipv6ll", func(n int) string { return fmt.Sprintf("fe80::%x", n) })
	case ip.IsPrivate():
		return r.mapped(s, "ipv6ula", func(n int) string { return fmt.Sprintf("fd00::%x", n) })
	default:
		return r.mapped(s, "ipv6", func(n int) string { return fmt.Sprintf("2001:db8::%x", n) })
	}
}
