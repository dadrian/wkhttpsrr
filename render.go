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

import "fmt"

// SvcParams is a placeholder for future typed svcparam handling.
type SvcParams map[string]any

// Record represents a DNS HTTPS/SVCB record in presentation-like form.
type Record struct {
    Owner    string
    TTL      uint32
    Priority uint16
    Target   string // "." for origin
    Params   SvcParams
}

// ZoneConfig controls rendering for future DNS adapters.
type ZoneConfig struct {
    TTL       uint32
    OwnerPort int
}

// ToZone converts a Document into a set of HTTPS/SVCB records for the given origin.
// Note: Minimal scaffolding; full RFC9460 rendering will be added later.
func ToZone(origin string, doc *Document, cfg ZoneConfig) ([]Record, error) {
    if doc == nil {
        return nil, fmt.Errorf("nil document")
    }
    // Basic mapping: each ServiceMode endpoint becomes one record.
    var out []Record
    for _, ep := range doc.Endpoints {
        if ep.Alias != "" {
            // AliasMode handling TBD - left to DNS provider adapter
            continue
        }
        pri := uint16(1)
        if ep.Priority != nil {
            if *ep.Priority < 0 || *ep.Priority > 65535 {
                return nil, fmt.Errorf("priority out of range")
            }
            pri = uint16(*ep.Priority)
        }
        target := ep.Target
        if target == "" {
            target = "."
        }
        out = append(out, Record{
            Owner:    origin,
            TTL:      cfg.TTL,
            Priority: pri,
            Target:   target,
            Params:   SvcParams(ep.Params),
        })
    }
    return out, nil
}

