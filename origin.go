package wkhttpsrr

import (
	"fmt"
	"net/netip"
	"strings"
)

// Origin identifies a backend origin host and port, with optional IPs
// to connect to directly (bypassing DNS) when provided.
type Origin struct {
	Host string // host or host:port

	Addrs []netip.Addr
}

func ParseIPList(s string) ([]netip.Addr, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	addrs := make([]netip.Addr, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		a, err := netip.ParseAddr(p)
		if err != nil {
			return nil, fmt.Errorf("invalid IP %q: %w", p, err)
		}
		addrs = append(addrs, a)
	}
	return addrs, nil
}
