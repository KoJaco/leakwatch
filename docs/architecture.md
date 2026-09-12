# Architecture

leakwatch is a library-first goroutine leak detector for Go 1.27+. It samples the
runtime's `goroutineleak` pprof profile on a schedule, clusters leaked goroutines
by stack fingerprint, retains a bounded history of observations, and interprets
that history into export-ready leak state.

## Four core concepts

### Profile

An immutable point-in-time snapshot of what the runtime reported at sample time `T`.
A profile answers "how many goroutines are leaked, and what do their stacks look
like right now?" — not whether a leak is growing or how long it has persisted.

Defined in `internal/profile`. Collected by `collector.Collector`.

### Fingerprint

A stable machine identity for a leak cluster. Two goroutines with the same
normalized stack belong to the same cluster and share a fingerprint.

- `ID` — short hash used as `leak_id` in metrics and URLs (e.g. `7f31c2a9`)
- `Key` — canonical string representation used to generate the ID

Fingerprints are **not** human-readable attribution. See [fingerprinting.md](fingerprinting.md).

### Observation

A timestamped set of cluster counts — raw historical data. Each sample appends one
observation to `History`.

Defined in `internal/domain`. Recorded by `observer.Observer`.

### Leak

The current interpreted state of a cluster across observations: count, first/last
seen, persistence, status (`new`, `growing`, `persistent`, `recurring`, `resolved`),
and human-readable site attribution.

Produced by pure `analysis.Analyze`. Downstream exporters read `analysis.Snapshot`,
never observations directly.

## Pipeline

```
Scheduler
    │
    ▼
Observer.Sample(ctx)
    │
    ├── Collector.Collect(ctx)          → profile.Profile
    ├── Fingerprinter.Fingerprint(prof) → []fingerprint.LeakCluster
    ├── History.Record(observation)     → append to ring buffer
    └── (on read) Analyze(observations) → []analysis.Leak
                                              │
                                              ▼
                                        analysis.Snapshot
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    ▼                         ▼                         ▼
            export/prometheus         export/http              export/json
            (metrics scrape)          (/debug/leaks)           (CLI, tooling)
```

### Sample cycle

1. **Scheduler** fires on `interval` (with optional `jitter` and `SampleGate`).
2. **Collector** fetches raw pprof bytes from `/debug/pprof/goroutineleak` and
   parses them into a `Profile`.
3. **Fingerprinter** normalizes stacks, clusters goroutines, and assigns each
   cluster a fingerprint and leak site.
4. **History** appends an `Observation` (timestamp + clusters). Older entries
   are dropped when capacity is exceeded (default: 256).
5. **Analyze** (called lazily on `Snapshot()`) interprets the full observation
   history into current `Leak` records with status and growth.

### Read path

Exporters call `Watcher.Snapshot()`, which delegates to `Observer.Snapshot()`.
Analysis is pure and recomputed on each read — there is no separate mutable
leak registry.

## Leak status

`Analyze` returns only leaks **present in the latest observation**. Clusters
that disappear are **resolved** — they are omitted from `Snapshot` and their
Prometheus time series drop off on the next scrape.

| Status | Meaning |
|--------|---------|
| `new` | First sample for this `leak_id` |
| `growing` | Net count increase over `DefaultGrowthWindow` (30 minutes) |
| `persistent` | Seen in ≥2 samples with stable count (zero growth rate) |
| `recurring` | Present → absent for ≥1 sample → present again |
| `resolved` | Not in latest observation (implicit; not emitted in `Snapshot`) |

Classification priority (first match wins): `recurring` → `new` → `growing` →
`persistent`.

Growth rates are derived at query time via `GrowthFor(observations, leakID,
window)` — not stored on `Leak`. `RankBySeverity` sorts active leaks for
display and alerting using count, persistence, and growth over the default
window.

## Package boundaries

| Path | Visibility | Responsibility |
|------|------------|----------------|
| `leakwatch` (root) | Public | `Watcher` API, options, wiring |
| `export/prometheus` | Public | Prometheus collector |
| `export/http` | Public | JSON debug HTTP handler |
| `export/json` | Public | Snapshot JSON encoding |
| `internal/scheduler` | Private | Periodic sampling with jitter and gate |
| `internal/collector` | Private | Profile fetch + parse orchestration |
| `internal/profile` | Private | Domain types, pprof source and parser |
| `internal/fingerprint` | Private | Stack normalization, clustering, IDs |
| `internal/observer` | Private | Sample orchestration, history |
| `internal/domain` | Private | `Observation` type |
| `internal/analysis` | Private | Pure leak interpretation |
| `internal/pipeline` | Private | Point-in-time profile → snapshot (CLI) |
| `cmd/leakwatch` | Binary | CLI (`analyze`, `inspect`, `version`) |

**Do not import `internal/` from outside this module.** The CLI is the only
non-library consumer and may import `internal/` directly.

### Interface policy

Interfaces exist only where multiple implementations are expected:

- `collector.Collector` — pprof endpoint, fixture profiles, test doubles
- `fingerprint.Fingerprinter` — swappable clustering algorithm

Everything else is a concrete type or pure function: `History`, `Analyze`,
`Scheduler`, `GrowthFor`, `DetectRecurrence`.

## Implementation order

The pipeline dependencies define the build sequence. Each phase should land as a
focused PR with unit tests; see [pr-checklist.md](pr-checklist.md).

| Phase | Package(s) | Delivers |
|-------|------------|----------|
| 1. Profile parsing | `internal/profile` | Fetch + parse `goroutineleak` pprof into `Profile` |
| 2. Fingerprinting | `internal/fingerprint` | Normalize, cluster, assign stable IDs and leak sites |
| 3. Analysis | `internal/analysis` | `Analyze`, `GrowthFor`, `DetectRecurrence`, status classification |
| 4. Growth metrics | `export/prometheus` | Wire `goroutine_leak_growth_rate` from history |
| 5. CLI | `cmd/leakwatch` | `analyze` (offline profile), `inspect` (live endpoint) |
| 6. Golden tests | `internal/fingerprint/testdata/` | Recorded profiles → expected clusters |
| 7. End-to-end tests | `test/integration/` | Leak generators → live watcher → assertions |

Phases 1–3 unblock a working library. Phases 4–5 complete observability and
tooling. Phases 6–7 validate correctness against real runtime output.

## Testing strategy

Three layers, each with a distinct purpose:

### Unit tests

Live alongside the package under test (`*_test.go`). Cover pure functions
(`Analyze`, `NormalizeStack`, `History.Record`, parser edge cases) with
synthetic or fixture data. Fast; run on every CI push.

### Golden / fixture tests

Recorded `goroutineleak` profiles in `internal/fingerprint/testdata/{channels,nethttp,grpc,workers}/`.
Assert that fingerprinting produces expected cluster counts, IDs, and leak sites.
Decouple fingerprint logic from live runtime timing.

### Integration leak generators

Programs in `test/integration/{basic,channels,nethttp,grpc}/` (build tag
`integration`). Each intentionally leaks goroutines in a realistic pattern.
Run manually or in CI to produce live profiles:

```sh
go run -tags integration ./test/integration/basic
```

These are **generators**, not assertions. They exist to produce realistic leak
profiles for fixture recording and manual inspection.

### End-to-end tests (planned)

Wire integration generators to leakwatch in CI:

1. Start a leak generator with `net/http/pprof` on a known port.
2. Create a `Watcher` pointed at its `/debug/pprof/goroutineleak` endpoint.
3. Sample until the expected leak clusters appear.
4. Assert on `Snapshot()`: cluster count, `leak_id` stability across samples,
   status transitions, and growth direction.

E2E tests belong in `test/integration/` or a top-level `test/e2e/` package
with the `integration` build tag. They depend on phases 1–3 being complete and
should be added in phase 7 of the implementation order.

```sh
# Planned CI invocation
go test -tags integration -count=1 -timeout 5m ./test/...
```

## Related documentation

- [sampling.md](sampling.md) — scheduler, jitter, sample gate, GC cost
- [fingerprinting.md](fingerprinting.md) — ID vs Key vs LeakSite, normalization, clustering
- [limitations.md](limitations.md) — runtime detector boundaries, metric caveats
- [metrics.md](metrics.md) — Prometheus metrics and query examples
- [pr-checklist.md](pr-checklist.md) — per-PR build and review checklist
