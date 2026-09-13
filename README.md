# leakwatch

Library-first goroutine leak detection for Go 1.27+, built on the runtime's
goroutine leak profile. A thin CLI supports offline analysis of exported profiles
and live pprof inspection.

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

# JSON output
leakwatch analyze -json profile.pb.gz

# Fetch and inspect a live pprof endpoint (localhost only by default)
leakwatch inspect http://127.0.0.1:6060/debug/pprof/goroutineleak
leakwatch inspect -json http://localhost:6060/debug/pprof/goroutineleak

# Fetch a remote host explicitly
leakwatch inspect --allow-remote https://staging.example/debug/pprof/goroutineleak
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

## Security

The debug HTTP handler (`ServeDebug`) and Prometheus metrics expose goroutine
stack traces and opaque `leak_id` labels. Treat them as sensitive operational
data.

- Bind debug and pprof endpoints to **loopback** (`127.0.0.1`) unless traffic is
  restricted by network policy.
- Never register `ServeDebug` on a public HTTP mux without protection. Wrap the
  handler with middleware:

```go
auth := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w, r) {
        if r.Header.Get("Authorization") != "Bearer "+token {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
_ = watcher.ServeDebug("/debug/leaks", httpexport.WithMiddleware(auth))
```

- `leakwatch inspect` accepts **localhost URLs only** by default. Pass
  `--allow-remote` to fetch other hosts (use only with trusted URLs).
- Profile fetch uses bounded reads and HTTP timeouts (see `WithMaxProfileBytes`
  and `WithHTTPTimeout`).

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
