# workers

Worker pool leak generator for integration testing and golden fixture recording.

## Run

```sh
go run -tags integration ./test/integration/workers
```

Starts pprof on `LEAKWATCH_PPROF_ADDR` (default `:6060`).

## Environment

| Variable | Default |
|----------|---------|
| `LEAKWATCH_PPROF_ADDR` | `:6060` |

## Record fixture

Wait for a GC cycle so leaked goroutines appear in the profile (retry curl if empty).

```sh
curl -o internal/fingerprint/testdata/workers/leak.pb.gz \
  http://127.0.0.1:6060/debug/pprof/goroutineleak
```

Regenerate expected clusters:

```sh
WRITE_GOLDEN=1 go test ./internal/fingerprint -run TestWriteGoldenFixtures/workers
```
