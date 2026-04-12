package middleware

import (
	"net"
	"net/http"
)

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, ip := range splitIPs(xff) {
			if ip = trimIP(ip); ip != "" {
				return ip
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
	for len(ip) > 0 && (ip[0] == ' ' || ip[0] == '\t') {
		ip = ip[1:]
	}
	for len(ip) > 0 && (ip[len(ip)-1] == ' ' || ip[len(ip)-1] == '\t') {
		ip = ip[:len(ip)-1]
	}
	return ip
}
