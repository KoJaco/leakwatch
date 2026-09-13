# Contributing to leakwatch

## Development setup

```sh
git clone https://github.com/KoJaco/leakwatch.git
cd leakwatch
go test ./...
```

Requires Go 1.27 or newer.

## Package boundaries

| Path | Visibility | Purpose |
|------|------------|---------|
| `leakwatch` (root) | Public | Minimal watcher API |
| `export/` | Public | Prometheus, HTTP, JSON exporters |
| `internal/` | Private | All pipeline implementation |
| `cmd/leakwatch` | Binary | CLI — imports `internal/` directly |

**Do not import `internal/` from outside this module.**

## Interface policy

Only create interfaces where multiple implementations are expected or mocking is
materially valuable:

- `collector.Collector` — pprof, fixtures, test doubles
- `fingerprint.Fingerprinter` — swappable clustering algorithm

Prefer concrete types and pure functions elsewhere (`History`, `Analyze`, `Scheduler`).

## Running checks

```sh
go vet ./...
go test -race -count=1 ./...
golangci-lint run
```

## Adding tests

Three layers — see [docs/architecture.md](docs/architecture.md#testing-strategy) for detail.

### Unit tests

- Live alongside the package they test (`*_test.go`).
- Cover pure functions and components with synthetic or fixture data.
- Run on every CI push: `go test -count=1 ./...`

### Golden / fixture tests

- Recorded `goroutineleak` profiles in `internal/fingerprint/testdata/`.
- Assert expected cluster counts, IDs, and leak sites.
- No live runtime required.

### Integration leak generators

- Programs in `test/integration/{basic,channels,nethttp,grpc}/` (build tag `integration`).
- Intentionally leak goroutines for manual inspection and fixture recording.

```sh
go run -tags integration ./test/integration/basic
```

### End-to-end tests

Implemented in [`test/e2e/`](../test/e2e/) with the `integration` build tag.
Run in CI with race detection:

```sh
go test -race -tags integration -count=1 -timeout 5m ./test/...
```

See [docs/pr-checklist.md](docs/pr-checklist.md) phase 7.

## Implementation PRs

Follow the per-phase checklist in [docs/pr-checklist.md](docs/pr-checklist.md).
Each implementation phase lands as a single focused PR in dependency order.
