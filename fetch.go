// Copyright 2025 David Adrian
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wkhttpsrr

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"time"
)

// Resolver allows injecting DNS resolution.
type Resolver interface {
	LookupIP(ctx context.Context, host string) (v4 []netip.Addr, v6 []netip.Addr, err error)
}

// defaultResolver uses net.Resolver.
type defaultResolver struct{}

func (defaultResolver) LookupIP(ctx context.Context, host string) ([]netip.Addr, []netip.Addr, error) {
	r := net.Resolver{}
	// Use Go's default; separate A and AAAA
	ips, err := r.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, nil, err
	}
	var v4, v6 []netip.Addr
	for _, ip := range ips {
		if a, ok := netip.AddrFromSlice(ip); ok {
			if a.Is4() {
				v4 = append(v4, a)
			} else if a.Is6() {
				v6 = append(v6, a)
			}
		}
	}
	return v4, v6, nil
}

// Fetch retrieves and parses the well-known origin-svcb JSON for the origin.
func Fetch(ctx context.Context, origin Origin, opts ...Option) (*Document, error) {
	cfg := options{resolver: defaultResolver{}}
	for _, o := range opts {
		o(&cfg)
	}

	url := fmt.Sprintf("https://%s/.well-known/origin-svcb", origin.Host)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: origin.Host, InsecureSkipVerify: cfg.insecureSkipVerify},
	}

	// If caller provided IPs, connect directly to the first one.
	// TODO(dadrian): Connect to all?
	if len(origin.Addrs) > 0 {
		d := &net.Dialer{Timeout: cfg.timeout}
		transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
			addr := net.JoinHostPort(origin.Addrs[0].String(), portOr443(origin.Host))
			return d.DialContext(ctx, network, addr)
		}
	}

	client := &http.Client{Transport: transport, Timeout: timeoutOrDefault(cfg.timeout)}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	// Ensure Host header matches the origin host
	req.Host = origin.Host

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	doc, err := Parse(body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func portOr443(host string) string {
	_, p, err := net.SplitHostPort(host)
	if p == "0" || err != nil {
		return "443"
	}
	return p
}

// Resolve resolves host to A/AAAA using the configured resolver.
func Resolve(ctx context.Context, host string, r Resolver) (v4 []netip.Addr, v6 []netip.Addr, err error) {
	if r == nil {
		r = defaultResolver{}
	}
	return r.LookupIP(ctx, host)
}

// options and Option pattern for Fetch.
type options struct {
	resolver           Resolver
	ipv4               []netip.Addr
	ipv6               []netip.Addr
	timeout            time.Duration
	insecureSkipVerify bool
}

// Option configures Fetch behavior.
type Option func(*options)

// WithResolver overrides DNS resolution.
func WithResolver(r Resolver) Option { return func(o *options) { o.resolver = r } }

// WithIPs sets explicit IPs for connecting, bypassing DNS.
func WithIPs(v4, v6 []netip.Addr) Option {
	return func(o *options) {
		if len(v4) > 0 {
			o.ipv4 = append([]netip.Addr{}, v4...)
		}
		if len(v6) > 0 {
			o.ipv6 = append([]netip.Addr{}, v6...)
		}
	}
}

// WithTimeout sets a per-request timeout.
func WithTimeout(d time.Duration) Option { return func(o *options) { o.timeout = d } }

// WithInsecureSkipVerify controls TLS verification (not recommended).
func WithInsecureSkipVerify(b bool) Option { return func(o *options) { o.insecureSkipVerify = b } }

func timeoutOrDefault(d time.Duration) time.Duration {
	if d <= 0 {
		return 10 * time.Second
	}
	return d
}
