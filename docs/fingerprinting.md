# Fingerprinting

Fingerprinting turns a `profile.Profile` into `[]fingerprint.LeakCluster` — groups
of leaked goroutines that share a stable machine identity and a human-readable
leak site.

This document describes the design intent and data model. Normalization rules and
clustering heuristics are subject to iteration; see the brainstorming session on
fingerprinting for open decisions.

## ID vs Key vs LeakSite

Three distinct concepts serve different consumers:

| Concept | Type | Purpose | Example consumer |
|---------|------|---------|------------------|
| **Fingerprint.ID** | `string` | Short stable hash for metrics labels and URLs | Prometheus `leak_id`, `/debug/leaks/{id}` |
| **Fingerprint.Key** | `string` | Canonical representation hashed to produce ID | Deduplication, cross-sample matching |
| **LeakSite** | struct | Human-readable attribution | HTTP debug handler, CLI output |

```go
type Fingerprint struct {
    ID  string // e.g. "7f31c2a9"
    Key string // canonical representation used to generate ID
}

type LeakSite struct {
    Location Location   // top application frame
    Stack    []Frame    // full normalized stack for display
}
```

**Rule:** never use `Fingerprint.ID` as display text. IDs are opaque and may
change if normalization rules are versioned. Always resolve to `LeakSite` for
human output.

**Rule:** `Fingerprint.Key` must be deterministic — same normalized stack always
produces the same key and ID across samples, processes, and builds (modulo
documented normalization exceptions).

## Pipeline stages

Fingerprinting runs in three stages:

```
Profile.Goroutines
    │
    ▼
NormalizeStack(stack)     — strip runtime noise, canonicalize frame text
    │
    ▼
ProjectStack(stack)       — reduce to application-relevant frames
    │
    ▼
ClusterGoroutines(grs)    — group by projected stack similarity
    │
    ▼
[]LeakCluster             — fingerprint ID/Key + LeakSite + count
```

Implemented in `internal/fingerprint/`:

| Function | File | Status |
|----------|------|--------|
| `NormalizeStack` | `normalize.go` | Implemented |
| `ProjectStack` | `projection.go` | Implemented |
| `ClusterGoroutines` | `cluster.go` | Implemented |
| `Fingerprinter.Fingerprint` | `fingerprint.go` | Orchestrates the above |

## Normalization

Normalization makes stacks comparable across samples. Goals:

1. **Stability** — the same logical leak site produces the same key across samples.
2. **Noise reduction** — runtime/internal frames do not dominate clustering.
3. **Predictability** — rules are documented and testable via golden fixtures.

### Planned rules (initial)

| Rule | Rationale |
|------|-----------|
| Strip or collapse `runtime.*` frames below the application entry | Runtime frames vary with scheduler state |
| Normalize function names (package path, no module version) | `example.com/foo/v2.Bar` and `example.com/foo/v3.Bar` may differ — document policy |
| Preserve file + line for application frames | Attribution requires source location |
| Handle inlined frames consistently | Inlining may alias multiple source lines to one location |
| Trim stacks above the blocking primitive | Focus on where the goroutine is stuck, not who spawned it |

Exact rules will be validated against golden fixtures in
`internal/fingerprint/testdata/{channels,nethttp,grpc,workers}/`.

### What normalization does not do

- Merge clusters with different blocking primitives (channel vs mutex wait).
- Deduplicate goroutines that are legitimately stuck at different call sites.
- Guarantee cross-binary stability (rebuilds with different inlining may shift keys).

## Clustering

Two goroutines belong to the same cluster when their **projected, normalized
stacks are identical**.

```
Goroutine A: main → handler → worker → <-ch
Goroutine B: main → handler → worker → <-ch
→ same cluster, count = 2

Goroutine C: main → handler → otherWorker → <-ch
→ different cluster (different projection)
```

### Leak site selection

For each cluster, `LeakSite.Location` is the **topmost application frame** in the
projected stack (closest to the blocking primitive). `LeakSite.Stack` is the full
projected stack for detail views.

### Count

`LeakCluster.Count` is the number of goroutines in the profile with that
fingerprint. This is a point-in-time count, not a cumulative total.

## ID generation

`Fingerprint.ID` is derived from `Fingerprint.Key`:

1. Build canonical key from normalized + projected stack (function:file:line per frame, joined).
2. Hash key (e.g. truncated SHA-256 or FNV) to a short hex string.
3. Use hash as `ID` for metrics and URLs.

Properties:

- **Stable within a normalization version** — same stack → same ID across samples.
- **Opaque** — ID reveals nothing about the stack without a lookup table.
- **Bounded cardinality** — one time series per active cluster, not per goroutine.

If normalization rules change in a future release, IDs may change for the same
logical leak. Document breaking changes in CHANGELOG.

## Goroutine IDs

`profile.Goroutine` intentionally has **no goroutine ID field** in v1. Runtime
goroutine IDs are not stable across samples. Identity is stack-derived only.

## Golden fixtures

Each subdirectory under `internal/fingerprint/testdata/` holds recorded
`goroutineleak` profiles for a leak pattern:

| Directory | Pattern |
|-----------|---------|
| `channels/` | Blocked on channel receive/send |
| `nethttp/` | Per-request goroutine leak in HTTP handler |
| `grpc/` | gRPC handler/worker leak |
| `workers/` | Worker pool goroutine leak |

Fixtures are recorded from `test/integration/` generators (or synthesized for
`grpc/` until phase 7). Golden tests in `golden_test.go` and
`internal/profile/golden_test.go` assert expected cluster count, IDs, and leak
sites. Regenerate with `WRITE_GOLDEN=1 go test ./internal/fingerprint -run TestWriteGoldenFixtures`.

## Related documentation

- [architecture.md](architecture.md) — where fingerprinting fits in the pipeline
- [limitations.md](limitations.md) — runtime detector boundaries affecting stacks
- [metrics.md](metrics.md) — how `leak_id` appears in Prometheus
- [pr-checklist.md](pr-checklist.md) — phase 2 PR checklist
