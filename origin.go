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
