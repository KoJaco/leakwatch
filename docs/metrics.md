# Metrics

leakwatch exposes goroutine leak state via Prometheus metrics, an HTTP debug
handler, and JSON export. This page documents the Prometheus metrics and
example queries.

## Prometheus metrics

Register the collector against a `Watcher` (or any `SnapshotProvider`):

```go
import (
    promexport "github.com/KoJaco/leakwatch/export/prometheus"
    "github.com/prometheus/client_golang/prometheus"
)

reg := prometheus.NewRegistry()
reg.MustRegister(promexport.NewCollector(watcher))
```

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `goroutine_leak_clusters` | Gauge | — | Number of active leak clusters in the latest snapshot |
| `goroutine_leak_count` | Gauge | `leak_id` | Goroutine count per cluster |
| `goroutine_leak_first_seen_timestamp` | Gauge | `leak_id` | Unix timestamp of first observation for this cluster |
| `goroutine_leak_growth_rate` | Gauge | `leak_id` | Growth per minute, derived from observation history at scrape time |

All metrics are gauges — they reflect current state, not cumulative totals.

### Label cardinality

`leak_id` cardinality equals the number of distinct leak clusters (normalized stack
fingerprints), not the number of leaked goroutines. Each active cluster produces
three labeled metrics. Resolved clusters are removed from the snapshot and their
time series disappear on the next scrape.

Monitor cardinality in deployments with highly variable stack traces (e.g. leaks
at different call sites per request before normalization stabilizes).

## Grafana dashboard

A starter dashboard is provided at `grafana/dashboards/goroutine-leaks.json`.
Import it and point at your Prometheus datasource.

Panels:

- **Active Leak Clusters** — `goroutine_leak_clusters`
- **Leak Count by ID** — `goroutine_leak_count`

## Query examples

### Active leak clusters

```promql
goroutine_leak_clusters
```

Alert when any leaks exist:

```promql
goroutine_leak_clusters > 0
```

### Total leaked goroutines

Sum across all clusters:

```promql
sum(goroutine_leak_count)
```

### Top leak clusters by count

```promql
topk(5, goroutine_leak_count)
```

### Growing leaks

Detect clusters with positive growth:

```promql
goroutine_leak_growth_rate > 0
```

Rate of total leaked goroutines increasing (alternative without growth rate):

```promql
sum(goroutine_leak_count) - sum(goroutine_leak_count offset 30m) > 0
```

### New leaks in the last hour

Compare current cluster count to one hour ago:

```promql
goroutine_leak_clusters - goroutine_leak_clusters offset 1h > 0
```

Per-cluster first seen within the last hour:

```promql
goroutine_leak_first_seen_timestamp > (time() - 3600)
```

### Persistent high-count leaks

Clusters with count above a threshold for 30+ minutes (stable, not necessarily growing):

```promql
min_over_time(goroutine_leak_count[30m]) > 10
```

### Leak age

Seconds since first observation:

```promql
time() - goroutine_leak_first_seen_timestamp
```

Alert on leaks older than 24 hours:

```promql
(time() - goroutine_leak_first_seen_timestamp) > 86400
```

## HTTP debug handler

`Watcher.ServeDebug("/debug/leaks")` registers:

| Endpoint | Response |
|----------|----------|
| `GET /debug/leaks` | JSON list of leaks (id, count, status) |
| `GET /debug/leaks/{leak_id}` | JSON detail (site, stack, first/last seen) |

Use this to resolve `leak_id` from metrics to human-readable attribution.

Example:

```sh
curl -s http://localhost:6060/debug/leaks/7f31c2a9 | jq .
```

## JSON export

`export/json.Encode(w, snapshot)` writes an `analysis.Snapshot` as indented JSON.
Used by the CLI and external tooling.

## Metric implementation status

| Metric | Status |
|--------|--------|
| `goroutine_leak_clusters` | Implemented |
| `goroutine_leak_count` | Implemented |
| `goroutine_leak_first_seen_timestamp` | Implemented |
| `goroutine_leak_growth_rate` | Implemented |

## Related documentation

- [limitations.md](limitations.md) — how to interpret metric values
- [fingerprinting.md](fingerprinting.md) — what `leak_id` means
- [architecture.md](architecture.md) — export boundary (`Snapshot`)
- [pr-checklist.md](pr-checklist.md) — phase 4 growth metrics checklist
