# Sampling

leakwatch samples the runtime `goroutineleak` profile on a fixed interval. This
page documents the scheduler, sample gate, jitter behaviour, and the GC cost of
each sample.

## Scheduler

The scheduler (`internal/scheduler`) drives periodic sampling. It is started by
`Watcher.Start` and runs until the context is cancelled or `Watcher.Stop` is called.

### Configuration

| Option | Default | Description |
|--------|---------|-------------|
| `WithInterval` | `5m` | Time between sample attempts |
| `WithJitter` | `30s` | Random delay added before each sample |
| `WithSampleGate` | `AlwaysAllow` | Predicate that can skip individual samples |

### Tick loop

On each tick:

1. Wait for `interval` (via `time.Ticker`).
2. If `jitter > 0`, sleep for a pseudo-random duration in `[0, jitter)`.
3. Evaluate `SampleGate`. If it returns `(false, nil)`, skip this tick.
4. Call `Observer.Sample(ctx)`.

If `interval <= 0`, the scheduler returns immediately without sampling.

If `Sample` returns an error, the scheduler **stops** and returns the error.
The current `Watcher.Start` implementation discards this error (`_ = scheduler.Run(...)`).
A future improvement may log and continue on transient fetch/parse failures.

## Jitter

Jitter spreads sample times to avoid synchronized load across replicas. Without
it, N instances started together would all sample (and trigger GC) at the same
instant.

### Current implementation

```go
jitter := time.Duration(int64(s.Jitter) * int64(time.Now().UnixNano()%100) / 100)
```

This scales `Jitter` by `UnixNano() % 100 / 100`, producing a value in
`[0, jitter)`. The distribution is **not uniform** — values cluster toward
lower durations because `UnixNano() % 100` is not a high-quality random source
and the modulo range is small.

### Caveats

- **Not cryptographically random.** Acceptable for load spreading; do not use
  jitter timing for security-sensitive decisions.
- **Correlated across replicas** started at the same wall-clock instant may
  receive similar jitter values within the same second.
- **Zero jitter** is valid and disables the delay entirely.
- **Jitter is applied after the tick**, not added to the interval. Total time
  between samples is `interval + jitter_sleep`, not `interval ± jitter`.

See the brainstorming notes on jitter improvements in a follow-up session.

## Sample gate

`SampleGate` is a function:

```go
type SampleGate func(context.Context) (bool, error)
```

| Return value | Behaviour |
|--------------|-----------|
| `(true, nil)` | Proceed with sample |
| `(false, nil)` | Skip this tick, continue scheduling |
| `(_, err)` | Stop the scheduler, propagate error |

Built-in gates:

- `collector.AlwaysAllow` — default; never skips
- `collector.NeverAllow` — always skips (useful for tests)
- `collector.GateIf(cond)` — permits sampling only when `cond(ctx)` is true

### Use cases

- **Load shedding:** skip samples when request latency or CPU exceeds a threshold.
- **Coordinated blackout:** skip during deployments or maintenance windows.
- **Testing:** use `NeverAllow` or a fixture gate to control sample timing.

The gate is evaluated **per tick**, not per process lifetime. A gate that
returns `false` does not delay the next tick — the scheduler moves on to the
next interval immediately.

## GC impact

Fetching a `goroutineleak` profile is **not a lightweight read**. The runtime
must run a full GC cycle with leak detection enabled before writing the profile.

From `runtime/pprof.writeGoroutineLeak`:

1. Acquire a global goroutine-leak profile lock (serializes concurrent requests).
2. Call `runtime_goroutineLeakGC()` — sets a pending flag and spins until a GC
   cycle completes with goroutine leak detection.
3. Write the profile from goroutines in `_Gleaked` state.

### Implications for leakwatch

| Concern | Guidance |
|---------|----------|
| **GC frequency** | Each sample triggers at least one GC. Default 5m interval is conservative. Sub-minute sampling on latency-sensitive services is not recommended without measurement. |
| **STW pause** | Leak detection runs during GC mark termination (stop-the-world). Frequent sampling increases STW exposure. |
| **Concurrent requests** | Multiple simultaneous `goroutineleak` fetches are serialized by the runtime lock. leakwatch's single scheduler avoids this; avoid external tools polling the same endpoint concurrently. |
| **Gate as throttle** | `SampleGate` is the primary mechanism to reduce GC pressure under load. |

### Frequency guidance

| Interval | Suitable for |
|----------|--------------|
| `1m–5m` | Development, staging, low-traffic services |
| `5m–15m` | Production services with pprof already enabled |
| `15m+` | High-throughput or latency-sensitive production |
| `< 1m` | Not recommended without explicit GC/latency benchmarking |

The runtime may also run leak-detection GC independently of leakwatch if other
callers fetch `goroutineleak` profiles (e.g. manual `curl`, other agents).

## pprof endpoint

The default collector fetches from:

```
http://127.0.0.1:6060/debug/pprof/goroutineleak
```

Override with `WithPProfURL`. The target process must import `net/http/pprof`
(or equivalent) and expose the `goroutineleak` profile (Go 1.27+).

The endpoint supports standard pprof query parameters:

- `?debug=0` (default) — protobuf profile (used by leakwatch)
- `?debug=1` — text format
- `?debug=2` — full goroutine stacks (includes non-leaked goroutines; not used by leakwatch)

## Related documentation

- [architecture.md](architecture.md) — where sampling fits in the pipeline
- [limitations.md](limitations.md) — what the runtime detector reports per sample
- [pr-checklist.md](pr-checklist.md) — phase 1 PR checklist for profile parsing
