package request

import (
	"net"
	"net/http"
	"strings"
)

// GetRealIP extracts the client's IP address.
// If trustProxyHeaders is true, it checks X-Forwarded-For and X-Real-IP headers.
// Otherwise, it strictly uses RemoteAddr to prevent IP spoofing.
func GetRealIP(r *http.Request, trustProxyHeaders bool) string {
	if trustProxyHeaders {
		// Check X-Forwarded-For first (behind proxy)
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// XFF can contain multiple IPs, the first one is the client
			parts := strings.Split(xff, ",")
			ip := strings.TrimSpace(parts[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}

		// Check X-Real-IP
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			ip := strings.TrimSpace(xri)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	// Sanitize/Validate the IP
	if parsedIP := net.ParseIP(host); parsedIP == nil {
		return "unknown"
	}
	return host
}
