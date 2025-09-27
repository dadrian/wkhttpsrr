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
    "testing"
)

func TestParseAndValidate_MinimalServiceMode(t *testing.T) {
    js := []byte(`{
        "regeninterval": 3600,
        "endpoints": [ {} ]
    }`)
    doc, err := Parse(js)
    if err != nil {
        t.Fatalf("Parse: %v", err)
    }
    if err := Validate(doc); err != nil {
        t.Fatalf("Validate: %v", err)
    }
    if doc.RegenInterval != 3600 {
        t.Fatalf("regeninterval got %d", doc.RegenInterval)
    }
    if len(doc.Endpoints) != 1 {
        t.Fatalf("endpoints len=%d", len(doc.Endpoints))
    }
}

func TestParseAndValidate_ServiceModeWithParams(t *testing.T) {
    js := []byte(`{
        "regeninterval": 3600,
        "endpoints": [
            {
                "priority": 1,
                "params": {
                    "alpn": ["h2","http/1.1"],
                    "ech": "QUJD",
                    "ipv4hint": ["192.0.2.1"],
                    "ipv6hint": ["2001:db8::1"]
                }
            }
        ]
    }`)
    doc, err := Parse(js)
    if err != nil {
        t.Fatalf("Parse: %v", err)
    }
    if err := Validate(doc); err != nil {
        t.Fatalf("Validate: %v", err)
    }
}

func TestValidate_AliasMixedInMultiEndpointFails(t *testing.T) {
    js := []byte(`{
        "regeninterval": 1000,
        "endpoints": [
            {"alias": "cdn1.example.com"},
            {"params": {"alpn":["h2"]}}
        ]
    }`)
    doc, err := Parse(js)
    if err != nil {
        t.Fatalf("Parse: %v", err)
    }
    if err := Validate(doc); err == nil {
        t.Fatalf("Validate expected error, got nil")
    }
}

func TestParseOrigin(t *testing.T) {
    cases := []struct{
        in string
        host string
        port int
    }{
        {"example.com", "example.com", 0},
        {"example.com:8443", "example.com", 8443},
        {"https://example.com", "example.com", 0},
        {"https://example.com:444", "example.com", 444},
    }
    for _, tc := range cases {
        got, err := ParseOrigin(tc.in)
        if err != nil {
            t.Fatalf("ParseOrigin(%q): %v", tc.in, err)
        }
        if got.Host != tc.host || got.Port != tc.port {
            t.Fatalf("ParseOrigin(%q) => (%q,%d)", tc.in, got.Host, got.Port)
        }
    }
}

