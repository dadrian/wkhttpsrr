# Repository Guidelines
 
## Project Structure & Module Organization
- Language: Go (module: `github.com/dadrian/wkhttpsrr`, Go 1.24).
- Reference spec: `draft-ietf-tls-wkech-08.txt` (kept at repo root for context).
- Suggested layout as the code grows:
  - `cmd/wkhttpsrr/`: main entrypoint(s) and CLI.
  - `internal/`: private packages for app logic.
  - `pkg/`: public packages (if intended for reuse).
  - `test/` or package-local `_test.go` files for unit tests.
  - `docs/`: protocol notes, design docs.
 
## Build, Test, and Development Commands
- Build: `go build ./...` — compiles all packages.
- Run (example): `go run ./cmd/wkhttpsrr` — runs the CLI entrypoint.
- Test: `go test ./...` — runs unit tests across modules.
- Vet: `go vet ./...` — reports suspicious constructs.
- Format: `gofmt -s -w .` and `goimports -w .` — canonical formatting/imports.
 
## Coding Style & Naming Conventions
- Formatting: enforce `gofmt` and `goimports`; no stylistic deviations.
- Indentation: tabs (Go default). Line length: prefer <120 chars.
- Names: packages lower_snake (short, domain-specific), exported identifiers use `CamelCase` with doc comments.
- Errors: return wrapped errors with context (`fmt.Errorf("...: %w", err)`). Avoid panics in library code.
 
## Testing Guidelines
- Framework: standard `testing` package; table-driven tests preferred.
- Files: name tests `*_test.go`; example tests use `ExampleXxx` where helpful.
- Coverage: target ≥ 80% for core packages; run `go test -cover ./...`.
- Determinism: avoid network/time in unit tests; inject interfaces or use fakes.
 
## Commit & Pull Request Guidelines
- Commits: concise, imperative subject (“Add parser for SRR records”); reference issues (`#123`) in body.
- Conventional tags optional (feat, fix, docs); keep logical changes isolated.
- PRs: include purpose, approach, and tradeoffs; link spec sections when relevant (e.g., WKECH references); add tests and docs; include usage examples for new commands in `cmd/wkhttpsrr`.
 
## Security & Configuration Tips
- No secrets in code or history; prefer env vars or config files ignored by VCS.
- Validate and sanitize all network-facing inputs; prefer constant-time comparisons for sensitive fields.
- Reproducibility: pin tool versions in CI; document flags in `README.md` as they are introduced.
