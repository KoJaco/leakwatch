# Metrics

## Prometheus metrics

| Metric | Labels | Description |
|--------|--------|-------------|
| `goroutine_leak_clusters` | — | Number of active leak clusters |
| `goroutine_leak_count` | `leak_id` | Goroutine count per cluster |
| `goroutine_leak_first_seen_timestamp` | `leak_id` | First observation unix timestamp |
| `goroutine_leak_growth_rate` | `leak_id` | Growth per minute (derived at scrape) |

<!-- TODO: expand with query examples -->
