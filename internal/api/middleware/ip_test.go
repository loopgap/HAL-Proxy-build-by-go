package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsTrustedProxy(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		trustedProxies []string
		want           bool
	}{
		{
			name:           "localhost IPv4 should NOT be trusted",
			remoteAddr:     "127.0.0.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			want:           false,
		},
		{
			name:           "localhost IPv6 should NOT be trusted",
			remoteAddr:     "[::1]:12345",
			trustedProxies: []string{"10.0.0.1"},
			want:           false,
		},
		{
			name:           "localhost string should NOT be trusted",
			remoteAddr:     "localhost:12345",
			trustedProxies: []string{"10.0.0.1"},
			want:           false,
		},
		{
			name:           "actual proxy should be trusted when in list",
			remoteAddr:     "10.0.0.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			want:           true,
		},
		{
			name:           "IP in CIDR range should be trusted",
			remoteAddr:     "10.0.0.5:12345",
			trustedProxies: []string{"10.0.0.0/24"},
			want:           true,
		},
		{
			name:           "IP not in CIDR range should NOT be trusted",
			remoteAddr:     "192.168.1.1:12345",
			trustedProxies: []string{"10.0.0.0/24"},
			want:           false,
		},
		{
			name:           "wildcard proxy should be trusted",
			remoteAddr:     "192.168.1.1:12345",
			trustedProxies: []string{"*"},
			want:           true,
		},
		{
			name:           "random external IP should NOT be trusted",
			remoteAddr:     "203.0.113.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			want:           false,
		},
		{
			name:           "empty trusted proxies should NOT trust localhost",
			remoteAddr:     "127.0.0.1:12345",
			trustedProxies: []string{},
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTrustedProxy(tt.remoteAddr, tt.trustedProxies)
			if got != tt.want {
				t.Errorf("isTrustedProxy(%q, %v) = %v, want %v", tt.remoteAddr, tt.trustedProxies, got, tt.want)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		trustedProxies []string
		xff            string
		xri            string
		want           string
	}{
		{
			name:           "localhost connection should NOT trust X-Forwarded-For",
			remoteAddr:     "127.0.0.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			xff:            "203.0.113.1",
			xri:            "",
			want:           "127.0.0.1", // Should return localhost IP, not X-Forwarded-For
		},
		{
			name:           "localhost IPv6 should NOT trust X-Forwarded-For",
			remoteAddr:     "[::1]:12345",
			trustedProxies: []string{"10.0.0.1"},
			xff:            "203.0.113.1",
			xri:            "",
			want:           "::1", // Should return localhost IP, not X-Forwarded-For
		},
		{
			name:           "trusted proxy connection should trust X-Forwarded-For",
			remoteAddr:     "10.0.0.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			xff:            "203.0.113.1",
			xri:            "",
			want:           "203.0.113.1",
		},
		{
			name:           "trusted proxy with multiple X-Forwarded-For returns first",
			remoteAddr:     "10.0.0.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			xff:            "203.0.113.1, 198.51.100.1, 192.0.2.1",
			xri:            "",
			want:           "203.0.113.1",
		},
		{
			name:           "untrusted proxy should NOT trust X-Forwarded-For",
			remoteAddr:     "192.168.1.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			xff:            "203.0.113.1",
			xri:            "",
			want:           "192.168.1.1", // Should return actual remote addr, not X-Forwarded-For
		},
		{
			name:           "wildcard trusted should accept X-Forwarded-For",
			remoteAddr:     "192.168.1.1:12345",
			trustedProxies: []string{"*"},
			xff:            "203.0.113.1",
			xri:            "",
			want:           "203.0.113.1",
		},
		{
			name:           "no X-Forwarded-For falls back to remote addr",
			remoteAddr:     "192.168.1.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			xff:            "",
			xri:            "",
			want:           "192.168.1.1",
		},
		{
			name:           "X-Real-IP used when X-Forwarded-For not trusted",
			remoteAddr:     "127.0.0.1:12345",
			trustedProxies: []string{"10.0.0.1"},
			xff:            "",
			xri:            "203.0.113.1",
			want:           "203.0.113.1", // X-Real-IP is still used from localhost
		},
		{
			name:           "CIDR trusted proxy should trust X-Forwarded-For",
			remoteAddr:     "10.0.0.5:12345",
			trustedProxies: []string{"10.0.0.0/24"},
			xff:            "203.0.113.1",
			xri:            "",
			want:           "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			got := getClientIP(req, tt.trustedProxies)
			if got != tt.want {
				t.Errorf("getClientIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitIPs(t *testing.T) {
	tests := []struct {
		name string
		xff  string
		want []string
	}{
		{
			name: "single IP",
			xff:  "192.0.2.1",
			want: []string{"192.0.2.1"},
		},
		{
			name: "multiple IPs",
			xff:  "192.0.2.1, 198.51.100.1, 203.0.113.1",
			want: []string{"192.0.2.1", " 198.51.100.1", " 203.0.113.1"},
		},
		{
			name: "single IP with spaces",
			xff:  " 192.0.2.1 ",
			want: []string{" 192.0.2.1 "},
		},
		{
			name: "multiple IPs with spaces",
			xff:  " 192.0.2.1 , 198.51.100.1 ",
			want: []string{" 192.0.2.1 ", " 198.51.100.1 "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitIPs(tt.xff)
			if len(got) != len(tt.want) {
				t.Errorf("splitIPs(%q) len = %d, want %d", tt.xff, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitIPs(%q)[%d] = %q, want %q", tt.xff, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTrimIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{
			name: "no trim needed",
			ip:   "192.0.2.1",
			want: "192.0.2.1",
		},
		{
			name: "leading space",
			ip:   " 192.0.2.1",
			want: "192.0.2.1",
		},
		{
			name: "trailing space",
			ip:   "192.0.2.1 ",
			want: "192.0.2.1",
		},
		{
			name: "both spaces",
			ip:   " 192.0.2.1 ",
			want: "192.0.2.1",
		},
		{
			name: "empty string",
			ip:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimIP(tt.ip)
			if got != tt.want {
				t.Errorf("trimIP(%q) = %q, want %q", tt.ip, got, tt.want)
			}
		})
	}
}
