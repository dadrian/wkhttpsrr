# Development Notes

These notes capture the current design, decisions, and next steps for implementing draft-ietf-tls-wkech-08 in this repository.

## Scope & Goals
- Implement fetch, parse, and validation for `/.well-known/origin-svcb` JSON (Sections 4–6 of the draft).
- Expose a reusable core package at repo root (`github.com/dadrian/wkhttpsrr`).
- Provide a minimal CLI with `fetch` and `validate` subcommands.
- Support callers that provide both hostname and explicit IPs (IPv4/IPv6), and callers that only provide a DNS name (we resolve A/AAAA).
- Keep DNS zone rendering logic in the library (not CLI) for future provider integrations.

## Package Layout
- Root package `wkhttpsrr`:
  - `origin.go`: `Origin`, `ParseOrigin` (accepts `host` or `host:port`; optional `https://` scheme allowed but not required).
  - `fetch.go`: `Fetch(ctx, Origin, ...Option)` with HTTPS-only, SNI set to `Origin.Host`, optional direct-IP dialing, `Resolver` interface, `WithIPs`, `WithTimeout`, `WithInsecureSkipVerify`.
  - `types.go`: `Document`, `Endpoint`, and `MarshalPretty`.
  - `parse.go`: `Parse` and `Validate` per Section 5; endpoint and param validation; priority normalization.
  - `render.go`: scaffolding for `ToZone`, `Record`, `SvcParams`, `ZoneConfig` (kept out of the CLI).
  - `doc.go`: package documentation.
- CLI `cmd/wkhttpsrr`:
  - Subcommands: `fetch`, `validate`.
  - Flags: `--ipv4`, `--ipv6`, `--timeout`, `--insecure` (validate also supports `--file`).

## Key Behaviors & Decisions
- CLI input does not require a scheme: accepts `host` or `host:port`. If a scheme is present, it must be `https`.
- `Fetch` constructs `https://<host[:port]>/.well-known/origin-svcb` and uses a custom `http.Transport` with:
  - TLS SNI = `Origin.Host` and Host header set to the same.
  - Optional direct-IP dialing if IPs are provided (bypassing DNS).
  - Otherwise, default Go resolver is used via `Resolver` interface (pluggable).
- Validation highlights:
  - `regeninterval` must be positive; `endpoints` array must be non-empty.
  - If more than one endpoint is present, no `alias` endpoints allowed (multi-endpoint SHOULD be ServiceMode only).
  - `target`/`alias` charset: lower ASCII letters, digits, `-`, `_`, `.` (per spec guidance).
  - Known params enforced by type/shape: `alpn` ([]string non-empty), `ech` (string, base64 best-effort), `ipv4hint`/`ipv6hint` (IP strings), `port` (number in range). Unknown params accepted as string or []string.
  - Priorities are normalized using carry-forward defaults (first defaults to 1 if unspecified).

## Usage Examples
- Fetch JSON (DNS resolution):
  - `go run ./cmd/wkhttpsrr fetch example.com`
- Fetch JSON (direct IPs, no DNS):
  - `go run ./cmd/wkhttpsrr fetch --ipv4 192.0.2.1,192.0.2.2 example.com:8443`
- Validate via network:
  - `go run ./cmd/wkhttpsrr validate example.com`
- Validate from file:
  - `go run ./cmd/wkhttpsrr validate --file ./testdata/origin.json`

## Testing
- Table-driven tests in `parse_test.go` cover:
  - Minimal service-mode object (`{}` endpoint).
  - ServiceMode with params and types.
  - Alias mixed in a multi-endpoint array (rejected).
  - `ParseOrigin` variations: `host`, `host:port`, optional `https://`.
- Tests are deterministic and avoid network.

## Security Considerations
- HTTPS-only fetch; system roots by default.
- `--insecure` flag disables TLS verification (off by default; discouraged).
- When dialing by IP, SNI and Host header are set to the origin host to preserve WebPKI semantics.

## Known Limitations / TODOs
- Direct-IP dialing currently uses only the first provided/resolved IP; improve to iterate over v4/v6 with backoff and error aggregation.
- `ech` validation is best-effort; add stricter parsing of ECHConfigList (lengths/format) if needed.
- Fully implement RFC9460 HTTPS/SVCB presentation rendering (including AliasMode and quoting) in `ToZone`.
- Add resolver injection to `Fetch` path when not using explicit IPs (interface exists; wire more tests with a fake resolver).
- Add ECH checker interface implementation to probe the backend (as suggested in Section 6) — behind a library interface; optional in CLI.
- Add more validation rules as the draft evolves; consider warnings vs hard errors.
- Consider a retry/backoff policy and caching for `Fetch`.
- Expand tests: edge cases for params, invalid chars, large inputs; golden tests for zone rendering when implemented.

## Open Questions
- TTL policy for `ToZone`: should default be derived from `regeninterval` (e.g., 50–80% of the interval), or left entirely to callers?
- Preferred output for aliasing in DNS adapters: HTTPS AliasMode vs CNAME fallback, behind a policy flag?
- Should CLI support printing a recommended TTL or summary info (without exposing full zone rendering)?

