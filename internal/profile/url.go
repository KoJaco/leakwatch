package profile

import (
	"net"
	"net/url"
	"strings"
)

// IsLocalhostURL reports whether urlStr points at a loopback host.
func IsLocalhostURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	host := u.Hostname()
	if host == "" {
		return false
	}

	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
