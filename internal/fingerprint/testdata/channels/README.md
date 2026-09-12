# channels

Golden fixture for goroutines blocked on a shared channel receive.

## Files

| File | Description |
|------|-------------|
| `leak.pb.gz` | Recorded `goroutineleak` profile |
| `clusters.json` | Expected fingerprint output |

## Generator

Recorded from [`test/integration/channels`](../../../test/integration/channels/main.go) — 50 goroutines blocked on `<-leak`.

## Expected cluster

- **ID:** `a7db680f`
- **Count:** 50
- **Location:** `main.main.func1` in `main.go:17`

## Record fixture

```sh
go run -tags integration ./test/integration/channels
# Wait for GC; retry curl if the profile is empty
curl -o internal/fingerprint/testdata/channels/leak.pb.gz \
  http://127.0.0.1:6060/debug/pprof/goroutineleak
```

## Regenerate clusters.json

```sh
WRITE_GOLDEN=1 go test ./internal/fingerprint -run TestWriteGoldenFixtures/channels
```

## Run golden test

```sh
go test ./internal/fingerprint -run TestGoldenFixtures/channels
```
