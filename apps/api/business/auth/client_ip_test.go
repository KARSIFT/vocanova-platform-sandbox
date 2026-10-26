package auth

import "testing"

func TestClientIPUsesNormalizedProxyAddressOrPeer(t *testing.T) {
	for _, tc := range []struct{ name, peer, real, want string }{
		{"normalized proxy", "172.18.0.2:8080", "198.51.100.7", "198.51.100.7"},
		{"IPv6 proxy", "172.18.0.2:8080", "2001:db8::7", "2001:db8::7"},
		{"no header IPv4", "198.51.100.8:1234", "", "198.51.100.8"},
		{"no header IPv6", "[2001:db8::8]:1234", "", "2001:db8::8"},
		{"bare IPv6", "2001:db8::9", "", "2001:db8::9"},
		{"invalid forwarded chain", "198.51.100.8:1234", "203.0.113.1, 198.51.100.9", "198.51.100.8"},
		{"malformed header", "198.51.100.8:1234", "attacker-key", "198.51.100.8"},
		{"unknown peer is bounded", "arbitrary-untrusted-value", "", "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClientIP(tc.peer, tc.real); got != tc.want {
				t.Fatalf("got %q; want %q", got, tc.want)
			}
		})
	}
}
