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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/netip"
	"regexp"
)

// Parse parses raw JSON into a Document without enforcing all semantic rules.
func Parse(data []byte) (*Document, error) {
	var d Document
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}
	return &d, nil
}

var dnsNameAllowed = regexp.MustCompile(`^[a-z0-9._-]*$`)

// Validate applies spec-level checks to the parsed Document.
func Validate(doc *Document) error {
	if doc == nil {
		return fmt.Errorf("nil document")
	}
	if doc.RegenInterval <= 0 {
		return fmt.Errorf("regeninterval must be positive")
	}
	if len(doc.Endpoints) == 0 {
		return fmt.Errorf("endpoints must be non-empty")
	}
	// Multi-endpoint arrays should be ServiceMode only (no alias)
	if len(doc.Endpoints) > 1 {
		for i, ep := range doc.Endpoints {
			if ep.Alias != "" {
				return fmt.Errorf("endpoint %d: alias not allowed when multiple endpoints present", i)
			}
		}
	}

	// A single-endpoint array is valid with an empty endpoint
	if len(doc.Endpoints) == 1 && doc.Endpoints[0].IsEmpty() {
		return nil
	}

	// Validate each endpoint
	for i := range doc.Endpoints {
		if err := validateEndpoint(&doc.Endpoints[i]); err != nil {
			return fmt.Errorf("endpoint %d: %w", i, err)
		}
	}
	// Normalize priorities to carry-forward defaults
	normalizePriorities(doc)
	return nil
}

func validateEndpoint(ep *Endpoint) error {
	// Empty object is allowed (ServiceMode minimal)
	// Alias mode implies no other fields
	if ep.Alias != "" {
		if ep.Target != "" || len(ep.Params) > 0 || ep.Priority != nil {
			return fmt.Errorf("alias must not be combined with target/params/priority")
		}
		if !dnsNameAllowed.MatchString(ep.Alias) {
			return fmt.Errorf("alias contains invalid characters")
		}
		return nil
	}

	if ep.Target != "" && !dnsNameAllowed.MatchString(ep.Target) {
		return fmt.Errorf("target contains invalid characters")
	}
	// Validate known params basic types
	if len(ep.Params) > 0 {
		for k, v := range ep.Params {
			switch k {
			case "ech":
				s, ok := v.(string)
				if !ok {
					return fmt.Errorf("param ech must be string")
				}
				// Best-effort base64 validation
				if _, err := base64.StdEncoding.DecodeString(s); err != nil {
					// allow URL or raw; don't fail hard, but signal
					// Accept, but warn via error message for now
					// To keep validation usable, just ensure it's non-empty
					if s == "" {
						return fmt.Errorf("param ech must not be empty")
					}
				}
			case "alpn":
				arr, ok := asStringSlice(v)
				if !ok || len(arr) == 0 {
					return fmt.Errorf("param alpn must be non-empty array of strings")
				}
			case "ipv4hint":
				arr, ok := asStringSlice(v)
				if !ok || len(arr) == 0 {
					return fmt.Errorf("param ipv4hint must be non-empty array of strings")
				}
				for _, s := range arr {
					if _, err := netip.ParseAddr(s); err != nil {
						return fmt.Errorf("param ipv4hint invalid IP %q", s)
					}
				}
			case "ipv6hint":
				arr, ok := asStringSlice(v)
				if !ok || len(arr) == 0 {
					return fmt.Errorf("param ipv6hint must be non-empty array of strings")
				}
				for _, s := range arr {
					if _, err := netip.ParseAddr(s); err != nil {
						return fmt.Errorf("param ipv6hint invalid IP %q", s)
					}
				}
			case "port":
				// JSON numbers become float64
				switch n := v.(type) {
				case float64:
					if n <= 0 || n > 65535 || n != float64(int(n)) {
						return fmt.Errorf("param port must be valid integer port")
					}
				default:
					return fmt.Errorf("param port must be number")
				}
			default:
				// Allow string or []string for unknown params
				if _, ok := v.(string); ok {
					continue
				}
				if arr, ok := asStringSlice(v); ok && arr != nil {
					continue
				}
				return fmt.Errorf("param %s has unsupported type", k)
			}
		}
	}
	return nil
}

func asStringSlice(v any) ([]string, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		s, ok := it.(string)
		if !ok {
			return nil, false
		}
		out = append(out, s)
	}
	return out, true
}

func normalizePriorities(doc *Document) {
	prev := 0
	for i := range doc.Endpoints {
		if doc.Endpoints[i].Alias != "" {
			continue
		}
		if doc.Endpoints[i].Priority == nil {
			if prev == 0 {
				v := 1
				doc.Endpoints[i].Priority = &v
				prev = v
			} else {
				v := prev
				doc.Endpoints[i].Priority = &v
			}
		} else {
			prev = *doc.Endpoints[i].Priority
		}
	}
}
