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

- Unit tests live alongside the package they test.
- Golden fixtures go in `internal/fingerprint/testdata/`.
- Integration leak generators go in `test/integration/`.
