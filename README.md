# leakwatch

Library-first goroutine leak detection for Go 1.27+, built on the runtime's
goroutine leak profile. A thin CLI supports offline analysis of exported profiles.

**Status:** skeleton — core pipeline structure is in place; parsing, fingerprinting,
and analysis are not yet implemented.

## Requirements

- Go 1.27 or newer.

## Install

```sh
go get github.com/KoJaco/leakwatch@latest
```

## Usage

```go
package main

import (
    "context"
    "time"

    "github.com/KoJaco/leakwatch"
)

func main() {
    watcher := leakwatch.New(
        leakwatch.WithInterval(5 * time.Minute),
        leakwatch.WithJitter(30 * time.Second),
    )

    watcher.Start(context.Background())
    _ = watcher.ServeDebug("/debug/leaks")
}
```

## CLI

```sh
# Analyze an exported goroutine leak profile offline
leakwatch analyze profile.pb.gz

# Fetch and inspect a live pprof endpoint
leakwatch inspect http://localhost:6060/debug/pprof/goroutineleak
```

## Pipeline

```
Scheduler → Observer.Sample()
              ├── Collector  → Profile
              ├── Fingerprinter → []LeakCluster
              ├── History.Record()
              └── Analyze() → []Leak → Snapshot
                                              ↓
                                    Prometheus / HTTP / JSON
```

## Four core concepts

| Concept | Meaning |
|---------|---------|
| **Profile** | Immutable point-in-time runtime snapshot |
| **Fingerprint** | Stable machine identity for a leak cluster |
| **Observation** | Timestamped cluster counts — raw historical data |
| **Leak** | Current interpreted state of a cluster |

## Documentation

| Doc | Topic |
|-----|-------|
| [architecture.md](docs/architecture.md) | Pipeline, package boundaries, implementation order, testing strategy |
| [sampling.md](docs/sampling.md) | Scheduler, jitter, sample gate, GC impact |
| [fingerprinting.md](docs/fingerprinting.md) | ID vs Key vs LeakSite, normalization, clustering |
| [limitations.md](docs/limitations.md) | Runtime detector boundaries, metric interpretation |
| [metrics.md](docs/metrics.md) | Prometheus metrics and query examples |
| [pr-checklist.md](docs/pr-checklist.md) | Per-phase PR checklist for implementation |

## License

MIT — see [LICENSE](LICENSE).
