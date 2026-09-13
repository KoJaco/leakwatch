# workers

Golden fixture for worker-pool goroutines blocked on per-worker channels.

## Files

| File | Description |
|------|-------------|
| `leak.pb.gz` | Recorded `goroutineleak` profile |
| `clusters.json` | Expected fingerprint output |

## Generator

Recorded from [`test/integration/workers`](../../../test/integration/workers/main.go) — 40 `workerLoop` goroutines blocked on `<-ch`.

## Expected cluster

- **ID:** `34314923ff52ce48`
- **Count:** 40
- **Location:** `main.workerLoop` in `main.go:14`

## Record fixture

```sh
go run -tags integration ./test/integration/workers
# Wait for GC; retry curl if the profile is empty
curl -o internal/fingerprint/testdata/workers/leak.pb.gz \
  http://127.0.0.1:6060/debug/pprof/goroutineleak
```

## Regenerate clusters.json

```sh
WRITE_GOLDEN=1 go test ./internal/fingerprint -run TestWriteGoldenFixtures/workers
```

## Run golden test

```sh
go test ./internal/fingerprint -run TestGoldenFixtures/workers
```
