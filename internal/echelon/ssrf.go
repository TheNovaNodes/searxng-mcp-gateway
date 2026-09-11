package echelon

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// IsDisallowedIP checks if an IP belongs to loopback, private, link-local,
// carrier-grade NAT, multicast, or unspecified addresses.
func IsDisallowedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}

// ValidateTargetURL validates the URL scheme (only http/https) and ensures
// the host does not resolve to private, loopback, or internal networks unless allowPrivate is true.
func ValidateTargetURL(rawURL string, allowPrivate bool) (*url.URL, error) {
	cleanURL := strings.TrimSpace(rawURL)
	if cleanURL == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	parsed, err := url.Parse(cleanURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme %q: only http and https are permitted", parsed.Scheme)
	}

	host := parsed.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing host in URL")
	}

	if allowPrivate {
		return parsed, nil
	}

	// Check if host is direct IP address
	if ip := net.ParseIP(host); ip != nil {
		if IsDisallowedIP(ip) {
			return nil, fmt.Errorf("SSRF protection: access to restricted address %s is prohibited", ip.String())
		}
		return parsed, nil
	}

	// Resolve hostname to check destination IPs
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resolver := net.DefaultResolver
	ips, err := resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %q: %w", host, err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("SSRF protection: no IP addresses found for host %q", host)
	}

	for _, ip := range ips {
		if IsDisallowedIP(ip) {
			return nil, fmt.Errorf("SSRF protection: host %q resolves to restricted address %s", host, ip.String())
		}
	}

	return parsed, nil
}
