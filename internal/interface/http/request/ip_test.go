package request

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRealIP(t *testing.T) {

	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		trust      bool
		want       string
	}{
		{
			name:       "direct connection (untrusted)",
			remoteAddr: "192.168.1.1:12345",
			headers:    nil,
			trust:      false,
			want:       "192.168.1.1",
		},
		{
			name:       "direct connection (trusted)",
			remoteAddr: "192.168.1.1:12345",
			headers:    nil,
			trust:      true,
			want:       "192.168.1.1",
		},
		{
			name:       "ignored headers when untrusted",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			trust:      false,
			want:       "10.0.0.1",
		},
		{
			name:       "behind proxy with X-Forwarded-For (trusted)",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			trust:      true,
			want:       "203.0.113.50",
		},
		{
			name:       "X-Forwarded-For with multiple IPs (trusted)",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18"},
			trust:      true,
			want:       "203.0.113.50",
		},
		{
			name:       "with X-Real-IP (trusted)",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.60"},
			trust:      true,
			want:       "203.0.113.60",
		},
		{
			name:       "X-Forwarded-For takes precedence (trusted)",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.50",
				"X-Real-IP":       "203.0.113.60",
			},
			trust: true,
			want:  "203.0.113.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := GetRealIP(req, tt.trust)
			if got != tt.want {
				t.Errorf("GetRealIP() = %s, want %s", got, tt.want)
			}
		})
	}
}
