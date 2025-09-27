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

// Package wkhttpsrr provides primitives to fetch, parse, and validate
// origin service binding information as defined in draft-ietf-tls-wkech-08.
//
// It supports fetching from an origin's well-known endpoint over HTTPS,
// parsing/validation of the JSON structure, and utilities for future
// DNS zone rendering and integration.
package wkhttpsrr

