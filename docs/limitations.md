# Limitations

leakwatch reports what the Go 1.27+ runtime goroutine leak detector finds, then
interprets that signal over time. It does not perform independent leak detection.

## What the runtime detector catches

The runtime marks a goroutine as leaked (`_Gleaked`) when, after a GC cycle with
leak detection enabled, it is blocked on a concurrency primitive that is
**exclusively unreachable** from the rest of the program.

From `runtime.findGoroutineLeaks`:

> Scans the remaining stackRoots and marks any which are blocked over exclusively
> unreachable concurrency primitives as leaked (deadlocked).

### Detected wait reasons

Leaks are identified when a goroutine is waiting on primitives that the GC can
prove unreachable:

| Wait type | Examples |
|-----------|----------|
| Channel wait | Blocked on `<-ch` or `ch <-` where the channel is unreachable |
| Sync wait | Blocked on `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, etc. where the primitive is unreachable |

The `goroutineleak` pprof profile contains only goroutines in `_Gleaked` state
(plus special handling for `main` blocked on `select{}` — see below).

### Main goroutine exception

If `main` is blocked on `select {}`, it is treated as leaked during analysis
(to preserve reachability of child leaks) but **excluded from the profile output**.
A program with `main` on `select{}` and child goroutines blocked on unreachable
channels will still report the child leaks.

## What it does not catch

| Scenario | Why |
|----------|-----|
| **Goroutines blocked on reachable primitives** | If another goroutine can still send on the channel or unlock the mutex, it is not a leak — the primitive is reachable. |
| **Slow goroutines** | A goroutine doing work (not blocked on a sync primitive) is not leaked, even if it runs forever. |
| **Goroutine count growth without deadlock** | Spawning goroutines that are still runnable or blocked on reachable primitives does not appear in `goroutineleak`. |
| **Leaks that become reachable intermittently** | If a reference to the blocking primitive is temporarily restored, the goroutine may escape detection until the next GC cycle. |
| **Leaks before first GC** | Leak detection requires a GC cycle. Immediately after startup, the profile may be empty even if leaks exist. |
| **Non-Go runtimes / cgo goroutines** | Only Go goroutines tracked by the runtime are considered. |
| **Goroutines in `select` on reachable channels** | `select` waiting on at least one reachable case is not a leak. |

leakwatch inherits all of these boundaries. It cannot report leaks the runtime
does not classify as `_Gleaked`.

## Sampling tradeoffs

Each sample triggers a GC cycle with leak detection (see [sampling.md](sampling.md)).
Tradeoffs:

| Higher frequency | Lower frequency |
|------------------|-----------------|
| Faster detection of new leaks | Lower GC/STW overhead |
| Better growth-rate accuracy | Coarser growth signals |
| More history entries consumed | May miss short-lived leaks within the interval |

### History window

`History` retains the last N observations (default 256). At a 5-minute interval,
this covers ~21 hours. Analysis beyond that window is not possible without
increasing capacity or exporting observations externally.

### Transient leaks

A leak that appears for one sample and disappears the next may be classified as
`recurring` (if it reappears) or may never surface in metrics if it resolves
between samples. Do not rely on leakwatch to catch sub-interval transient leaks
at default sampling rates.

## leak_id generation

`leak_id` is the `fingerprint.Fingerprint.ID` — a hash of the canonical stack
key after normalization and projection. See [fingerprinting.md](fingerprinting.md).

### Properties

- **Stable across samples** for the same normalized stack (within a normalization version).
- **Opaque** — does not encode function names or file paths.
- **Not a stack hash of raw pprof output** — normalization runs first.

### When leak_id changes

| Event | Effect |
|-------|--------|
| Code deployed with different stack layout | New `leak_id` for the same logical bug |
| Normalization rule change in leakwatch | IDs may change; treat as breaking change |
| Different blocking site in the same function | Different `leak_id` (different projected stack) |

Use `leak_id` for time-series continuity within a deployment. Use `LeakSite` for
human investigation.

## Metric interpretation

### `goroutine_leak_clusters`

Count of active leak clusters in the latest snapshot. A rise indicates new leak
**types** (distinct stacks), not necessarily more total leaked goroutines.

### `goroutine_leak_count{leak_id}`

Point-in-time count of leaked goroutines for one cluster. Compare across time to
see growth. A flat line with a high value indicates a **persistent** leak, not
necessarily a growing one.

### `goroutine_leak_first_seen_timestamp{leak_id}`

Unix timestamp of first observation. Reset only when the cluster disappears from
history and reappears (new fingerprint session). Does not reset on process restart
unless history is also reset.

### `goroutine_leak_growth_rate{leak_id}`

Derived at scrape time from observation history via `GrowthFor` over the
default 30-minute window:

- Positive rate → goroutines in this cluster are increasing between samples.
- Zero rate → stable count (persistent leak) or insufficient history.
- Negative rate → count decreasing (possible resolution).

Growth rate depends on sample interval. Compare rates only within the same
`interval` configuration.

### Alerting guidance

| Signal | Suggested interpretation |
|--------|--------------------------|
| `goroutine_leak_clusters > 0` | At least one leak type exists — investigate |
| `goroutine_leak_count` increasing over 30m+ | Growing leak — prioritize |
| `goroutine_leak_count` flat and high | Persistent leak — may be intentional (worker pool bug) or accepted debt |
| Cluster appears then disappears repeatedly | Recurring leak — may indicate lifecycle bug (e.g. per-request leak under load) |

Always correlate with `LeakSite` via `/debug/leaks/{leak_id}` before acting.

## Related documentation

- [sampling.md](sampling.md) — GC cost and frequency guidance
- [fingerprinting.md](fingerprinting.md) — how `leak_id` is derived
- [metrics.md](metrics.md) — PromQL examples
- [architecture.md](architecture.md) — leak status definitions
