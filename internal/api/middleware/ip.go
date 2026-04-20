package middleware

import (
	"net"
	"net/http"
	"strings"
)

// isTrustedProxy checks if the remote address is from a trusted proxy.
// NOTE: Localhost connections (127.0.0.1, ::1, localhost) are NOT considered
// trusted because they can be trivially spoofed by an attacker. X-Forwarded-For
// headers from localhost connections should be ignored.
func isTrustedProxy(remoteAddr string, trustedProxies []string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	// Localhost addresses can be spoofed - they should NOT be trusted for header forwarding
	// An attacker can easily spoof their source IP to 127.0.0.1 to bypass rate limiting
	if host == "127.0.0.1" || host == "::1" || host == "localhost" {
		return false
	}

	for _, proxy := range trustedProxies {
		if proxy == "*" || proxy == host {
			return true
		}
		// Check if trusted proxy is a CIDR range
		if strings.Contains(proxy, "/") {
			_, ipnet, err := net.ParseCIDR(proxy)
			if err == nil {
				ip := net.ParseIP(host)
				if ip != nil && ipnet.Contains(ip) {
					return true
				}
			}
		}
	}
	return false
}

func getClientIP(r *http.Request, trustedProxies []string) string {
	// Only trust X-Forwarded-For if request comes from a trusted proxy
	if isTrustedProxy(r.RemoteAddr, trustedProxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			for _, ip := range splitIPs(xff) {
				if ip = trimIP(ip); ip != "" {
					return ip
				}
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := trimIP(xri); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func splitIPs(xff string) []string {
	var ips []string
	start := 0
	for i := 0; i < len(xff); i++ {
		if xff[i] == ',' {
			ips = append(ips, xff[start:i])
			start = i + 1
		}
	}
	ips = append(ips, xff[start:])
	return ips
}

func trimIP(ip string) string {
	return strings.TrimSpace(ip)
}
