# PR checklist

Per-phase checklist for implementation PRs. Each phase should land as a single
focused branch and PR. See [architecture.md](architecture.md) for the full
pipeline context.

## General checklist (every PR)

- [ ] Branch name follows `feat/<phase>` or `docs/<topic>` convention
- [ ] Single logical change per PR
- [ ] `go vet ./...` passes
- [ ] `go test -race -count=1 ./...` passes
- [ ] `golangci-lint run` passes (if installed locally)
- [ ] New code has unit tests; no skipped tests unless explicitly deferred with issue reference
- [ ] CHANGELOG.md updated under `[Unreleased]` if user-visible behaviour changes
- [ ] Relevant doc in `docs/` updated if behaviour or API changes

---

## Phase 1: Profile parsing

**Branch:** `feat/profile-parsing`  
**Packages:** `internal/profile`, `internal/collector`

### Deliverables

- [x] `PProfSource.Fetch` — HTTP GET with context cancellation
- [x] `DefaultParser.Parse` — protobuf `goroutineleak` profile → `profile.Profile`
- [x] Populate `Profile.CapturedAt` and `Goroutine.Stack` frames
- [x] Unit tests with recorded profile bytes (minimal fixture in `testdata/`)

### Tests

- [x] `TestParser_*` — parse valid profile, empty profile, malformed input
- [x] `TestPProfSource_Fetch` — mock HTTP server (or httptest)
- [x] Remove `t.Skip("not implemented")` from `parser_test.go`

### Manual verification

```sh
# Start any integration generator with pprof
go run -tags integration ./test/integration/basic

# Fetch profile manually
curl -o /tmp/leak.pb.gz http://localhost:6060/debug/pprof/goroutineleak
```

---

## Phase 2: Fingerprinting

**Branch:** `feat/fingerprinting`  
**Packages:** `internal/fingerprint`  
**Depends on:** Phase 1

### Deliverables

- [x] `NormalizeStack` — documented rules from [fingerprinting.md](fingerprinting.md)
- [x] `ProjectStack` — application-frame projection
- [x] `ClusterGoroutines` — group by projected stack
- [x] `DefaultFingerprinter.Fingerprint` — orchestrate and assign ID/Key/Site
- [x] Remove `ErrNotImplemented` from fingerprinter

### Tests

- [x] `TestFingerprint_*` per fixture directory (synthetic unit tests; golden `.pb.gz` deferred to phase 6)
- [ ] Golden fixtures recorded in `internal/fingerprint/testdata/{channels,nethttp,grpc,workers}/`
- [x] Remove `t.Skip("not implemented")` from `fingerprint_test.go`

### Fixture recording

```sh
# 1. Run integration generator
go run -tags integration ./test/integration/channels

# 2. Capture profile
curl -o internal/fingerprint/testdata/channels/leak.pb.gz \
  http://localhost:6060/debug/pprof/goroutineleak

# 3. Write expected output alongside fixture (clusters.json or test code)
```

---

## Phase 3: Analysis

**Branch:** `feat/analysis`  
**Packages:** `internal/analysis`, `internal/observer`  
**Depends on:** Phase 2

### Deliverables

- [x] `Analyze` — observations → `[]Leak` with status classification
- [x] `GrowthFor` — rate per minute/hour over configurable window
- [x] `DetectRecurrence` — disappear/reappear detection
- [x] `RankBySeverity` — sort by count × growth × persistence
- [x] Status transitions: `new`, `growing`, `persistent`, `recurring`, `resolved`

### Tests

- [x] Synthetic observation sequences (no runtime dependency)
- [x] `TestHistoryRecord` — ring buffer behaviour
- [x] Remove skips from `analysis_test.go`, `history_test.go`

---

## Phase 4: Growth metrics

**Branch:** `feat/growth-metrics`  
**Packages:** `export/prometheus`  
**Depends on:** Phase 3

### Deliverables

- [ ] Wire `goroutine_leak_growth_rate` from `GrowthFor` at scrape time
- [ ] Pass observation history to collector (or expose via `SnapshotProvider`)

### Tests

- [ ] `collector_test.go` — assert non-zero growth rate with synthetic history
- [ ] Verify metric values with `prometheus/testutil`

---

## Phase 5: CLI

**Branch:** `feat/cli`  
**Packages:** `cmd/leakwatch`  
**Depends on:** Phases 1–3

### Deliverables

- [ ] `leakwatch analyze <profile.pb.gz>` — offline analysis, human-readable output
- [ ] `leakwatch inspect <pprof-url>` — fetch + analyze live endpoint
- [ ] Shared pipeline code with library (no duplicated logic)

### Tests

- [ ] CLI integration test with fixture profile file
- [ ] Exit codes: 0 on success, 1 on error

### Manual verification

```sh
go run ./cmd/leakwatch analyze internal/fingerprint/testdata/channels/leak.pb.gz
go run ./cmd/leakwatch inspect http://localhost:6060/debug/pprof/goroutineleak
```

---

## Phase 6: Golden tests

**Branch:** `feat/golden-fixtures`  
**Depends on:** Phases 1–2

### Deliverables

- [ ] Fixtures for all four patterns: channels, nethttp, grpc, workers
- [ ] CI runs golden tests (no `integration` tag required)
- [ ] README in each `testdata/` subdirectory documenting how the fixture was recorded

### Tests

- [ ] `go test ./internal/fingerprint/...` passes with fixtures
- [ ] `go test ./internal/profile/...` passes with fixtures

---

## Phase 7: End-to-end integration tests

**Branch:** `feat/e2e-tests`  
**Packages:** `test/integration/` or `test/e2e/`  
**Depends on:** Phases 1–3

### Deliverables

- [ ] E2E test harness: start leak generator + pprof server in test
- [ ] Create `Watcher` pointed at test server
- [ ] Sample until expected clusters appear (with timeout)
- [ ] Assert: cluster count, `leak_id` stability, status, growth direction
- [ ] CI job runs `go test -tags integration -timeout 5m ./test/...`

### Test matrix

| Generator | Expected assertion |
|-----------|-------------------|
| `basic` | ≥1 cluster, count ≥10 |
| `channels` | ≥1 cluster, stable `leak_id` across 2+ samples |
| `nethttp` | ≥1 cluster after requests |
| `grpc` | Implement generator first, then assert |

### CI update

Add to `.github/workflows/test.yml`:

```yaml
- name: Integration tests
  run: go test -tags integration -count=1 -timeout 5m ./test/...
```

### Notes

- E2E tests are slow (GC per sample). Use short intervals in test config.
- Generators in `test/integration/` remain useful for manual debugging and fixture recording.
- E2E tests add automated assertions on top of generators.

---

## Documentation PRs

Documentation-only PRs use `docs/<topic>` branches and skip implementation
checklists above. They should still:

- [ ] Cross-link sibling docs
- [ ] Update README.md `Documentation` section if new pages are added
- [ ] Keep [architecture.md](architecture.md) implementation order in sync

---

## Review focus by phase

| Phase | Reviewer should verify |
|-------|------------------------|
| 1 | Parser handles real Go 1.27 `goroutineleak` protobuf format |
| 2 | Same stack → same ID; different stacks → different IDs |
| 3 | Status transitions match definitions in architecture.md |
| 4 | Growth rate matches manual calculation from fixture observations |
| 5 | CLI output matches HTTP debug handler for same snapshot |
| 6 | Fixtures reproducible from documented recording steps |
| 7 | E2E tests fail when fingerprinting/analysis is broken |
