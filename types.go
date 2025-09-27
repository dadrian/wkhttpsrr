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
	"encoding/json"
)

// Document represents the top-level JSON structure from the well-known endpoint.
type Document struct {
	RegenInterval int        `json:"regeninterval"`
	Endpoints     []Endpoint `json:"endpoints"`
}

// Endpoint models either ServiceMode or AliasMode. If Alias is non-empty,
// it is considered AliasMode. An empty object is a valid ServiceMode with
// target="." and no params.
type Endpoint struct {
	Alias    string                     `json:"alias,omitempty"`
	Target   string                     `json:"target,omitempty"`
	Priority *int                       `json:"priority,omitempty"`
	Params   map[string]any             `json:"params,omitempty"`
	raw      map[string]json.RawMessage `json:"-"`
}

func (e *Endpoint) IsEmpty() bool {
	return e.Alias == "" && e.Target == "" && e.Priority == nil && e.Params == nil
}

// UnmarshalJSON retains raw fields to help validation detect empty objects.
func (e *Endpoint) UnmarshalJSON(b []byte) error {
	type Alias Endpoint
	var x Alias
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	*e = Endpoint(x)
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err == nil {
		e.raw = m
	}
	return nil
}

// MarshalPretty returns stable indented JSON for display.
func MarshalPretty(doc *Document) ([]byte, error) {
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return b, nil
}
