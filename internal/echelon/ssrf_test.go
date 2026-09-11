package echelon

import (
	"testing"
)

func TestValidateTargetURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		allowPrivate bool
		wantErr      bool
	}{
		{"empty url", "", false, true},
		{"invalid scheme ftp", "ftp://example.com/file", false, true},
		{"invalid scheme file", "file:///etc/passwd", false, true},
		{"missing host", "https://", false, true},
		{"loopback ipv4", "http://127.0.0.1:8080/secret", false, true},
		{"loopback ipv4 allowed", "http://127.0.0.1:8080/secret", true, false},
		{"loopback ipv6", "http://[::1]:8080/secret", false, true},
		{"cloud metadata", "http://169.254.169.254/latest/meta-data/", false, true},
		{"private 10.x", "http://10.0.0.1/admin", false, true},
		{"private 192.168.x", "http://192.168.1.1/", false, true},
		{"private 172.16.x", "http://172.16.0.1/", false, true},
		{"unspecified 0.0.0.0", "http://0.0.0.0:80/", false, true},
		{"public valid url", "https://example.com/page", false, false},
		{"public valid http", "http://example.com", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateTargetURL(tt.url, tt.allowPrivate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTargetURL(%q, %v) error = %v, wantErr %v", tt.url, tt.allowPrivate, err, tt.wantErr)
			}
		})
	}
}
