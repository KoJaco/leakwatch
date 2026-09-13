# nethttp

Golden fixture for per-request goroutine leaks in an HTTP handler.

## Files

| File | Description |
|------|-------------|
| `leak.pb.gz` | Recorded `goroutineleak` profile |
| `clusters.json` | Expected fingerprint output |

## Generator

Recorded from [`test/integration/nethttp`](../../../test/integration/nethttp/main.go). Trigger leaks with HTTP requests before capture:

```sh
curl http://127.0.0.1:8080/
```

## Expected cluster

- **ID:** `897206aadc79b7e6`
- **Count:** varies with request count (fixture recorded with 5)
- **Location:** `main.main.func1.1` in `main.go:16`

## Record fixture

```sh
go run -tags integration ./test/integration/nethttp
curl http://127.0.0.1:8080/
curl -o internal/fingerprint/testdata/nethttp/leak.pb.gz \
  http://127.0.0.1:6060/debug/pprof/goroutineleak
```

## Regenerate clusters.json

```sh
WRITE_GOLDEN=1 go test ./internal/fingerprint -run TestWriteGoldenFixtures/nethttp
```

## Run golden test

```sh
go test ./internal/fingerprint -run TestGoldenFixtures/nethttp
```
