# End-to-end integration tests

Live tests that spawn leak generators as subprocesses, wire a `leakwatch.Watcher`
to each generator's pprof endpoint, and assert on `Snapshot()` output.

## Run locally

Requires Go 1.27+ with goroutine leak detection enabled.

```sh
go test -tags integration -count=1 -timeout 5m ./test/e2e/...
```

## Test matrix

| Test | Generator | Assertions |
|------|-----------|------------|
| `TestBasicLeak` | `basic` | ≥1 leak, count ≥10, status `new` |
| `TestChannelsStableLeakID` | `channels` | stable `leak_id` across ≥2 observations, status `persistent` |
| `TestNetHTTPLLeakAfterRequests` | `nethttp` | leaks after HTTP requests, count growth, positive growth rate |
| `TestGRPCLeak` | `grpc` | ≥1 leak with gRPC frames in projected stack |

## Environment variables

Generators read these env vars (set automatically by the harness):

| Variable | Default | Used by |
|----------|---------|---------|
| `LEAKWATCH_PPROF_ADDR` | `:6060` | all generators |
| `LEAKWATCH_HTTP_ADDR` | `:8080` | `nethttp` |
| `LEAKWATCH_GRPC_ADDR` | `:50051` | `grpc` |

## Timeouts

- Profile readiness: 60s (handles GC delay before leaks appear in pprof)
- Leak appearance: 30s
- Second observation: 15s
- Watcher sample interval: 500ms (zero jitter)

## CI

Integration tests run in CI via:

```sh
go test -tags integration -count=1 -timeout 5m ./test/...
```

Unit tests (`go test ./...`) exclude these tests via the `integration` build tag.
