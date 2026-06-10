# Session Log: Security Dependency Upgrades — CVE Remediation
Date: 2026-06-10
Owner: Jeff Kopko
Related Plan: docs/plans/2026-06-10_Jeff-Kopko_security-dep-upgrades.md

## Milestones
- [x] Dependencies upgraded
- [x] Build + vet + tests pass
- [x] CHANGELOG updated
- [ ] PR created + merged
- [ ] Dependabot alerts confirmed closed

## Commentary Stream
### 09:00
- What I'm doing: Identified 15 Dependabot alerts; 8 fixable now, 3 docker CVEs have no upstream patch, 4 already resolved
- Why: GitHub security notifications for FortinetCloudCSE org
- Notes: docker/docker v29.x not yet on Go module proxy — cannot bump past v28.5.2 today

### Implementation (2026-06-10)
- `go get go.opentelemetry.io/otel*@v1.43.0 go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@v0.62.0` — upgraded otel family; go.mod go directive bumped to 1.25.0 (required by otel v1.43.0)
- `go get github.com/ulikunitz/xz@v0.5.15` — fixed CVE-2025-58058
- `go get golang.org/x/crypto@v0.53.0` — fixed CVE-2025-47914 and CVE-2025-58181
- `go get github.com/docker/docker@v28.5.2+incompatible` — partial fix for CVE-2026-34040 and CVE-2026-33997; 3 CVEs remain unpatched upstream
- `go mod tidy` — pruned unused deps, added new transitive deps (containerd/errdefs, moby/sys/atomicwriter)
- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — fortihugorunner/tests: PASS (0.003s); all other packages have no test files

## Commands (high-level)
- `go get go.opentelemetry.io/otel/...@v1.43.0` — upgrade otel family
- `go get github.com/ulikunitz/xz@v0.5.15` — fix CVE-2025-58058
- `go get golang.org/x/crypto@v0.53.0 golang.org/x/net@latest` — fix x/crypto CVEs
- `go mod tidy` — prune unused indirect deps
- `go build ./...` — verify compile
- `go vet ./...` — static analysis
- `go test ./...` — run test suite

## Dead-ends / Rejected Options
- docker/docker v29.x upgrade: not yet available on module proxy under `+incompatible` path (max is v28.5.2)
- ulikunitz/xz v0.6.x: alpha releases only, not suitable for production

## Risks & Mitigations
- Risk: otel v1.43.0 API breakage
  - Mitigation: `go build ./...` will catch compile errors immediately
- Risk: docker v28 API surface changes breaking existing code
  - Mitigation: `go vet` + build verify
