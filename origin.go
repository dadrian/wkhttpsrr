package wkhttpsrr

import (
    "fmt"
    "net"
    "net/netip"
    "strconv"
    "strings"
)

// Origin identifies a backend origin host and port, with optional IPs
// to connect to directly (bypassing DNS) when provided.
type Origin struct {
    Host string
    Port int // defaults to 443 if zero

    IPv4 []netip.Addr
    IPv6 []netip.Addr
}

// HostPort returns host:port with a default of 443.
func (o Origin) HostPort() string {
    p := o.Port
    if p == 0 {
        p = 443
    }
    return net.JoinHostPort(o.Host, strconv.Itoa(p))
}

// ParseOrigin parses inputs like:
//   example.com
//   example.com:8443
//   https://example.com
//   https://example.com:8443
// It defaults to HTTPS semantics and port 443; if a scheme is present, it
// must be https.
func ParseOrigin(s string) (Origin, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return Origin{}, fmt.Errorf("empty origin")
    }
    // strip optional scheme
    if i := strings.Index(s, "://"); i >= 0 {
        scheme := strings.ToLower(s[:i])
        if scheme != "https" {
            return Origin{}, fmt.Errorf("unsupported scheme %q (only https)", scheme)
        }
        s = s[i+3:]
    }

    host, portStr, err := net.SplitHostPort(s)
    if err != nil {
        // If missing port, treat whole as host
        if strings.Contains(err.Error(), "missing port in address") {
            host = s
            portStr = ""
        } else {
            return Origin{}, fmt.Errorf("parse origin: %w", err)
        }
    }
    host = strings.TrimSpace(host)
    if host == "" {
        return Origin{}, fmt.Errorf("missing host")
    }
    var port int
    if portStr != "" {
        p, perr := strconv.Atoi(portStr)
        if perr != nil || p <= 0 || p > 65535 {
            return Origin{}, fmt.Errorf("invalid port %q", portStr)
        }
        port = p
    }
    return Origin{Host: host, Port: port}, nil
}

