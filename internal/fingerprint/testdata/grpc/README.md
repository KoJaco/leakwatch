# grpc

Golden fixture for gRPC-style handler stack frames.

## Files

| File | Description |
|------|-------------|
| `leak.pb.gz` | Synthetic `goroutineleak` profile |
| `clusters.json` | Expected fingerprint output |

## Generator

**Synthetic** — built by `TestWriteSyntheticFixtures` in `golden_test.go` using
gRPC-like stack frames (`google.golang.org/grpc.(*Server).handleStream`, etc.).

A live gRPC leak generator lives in [`test/integration/grpc`](../../../test/integration/grpc/main.go)
and is exercised by [`test/e2e/e2e_test.go`](../../../test/e2e/e2e_test.go).

## Expected cluster

- **ID:** `e3a195b8d37bc2d4`
- **Count:** 8
- **Location:** `google.golang.org/grpc.(*Server).handleStream` in `server.go:100`

## Regenerate fixture and clusters.json

```sh
WRITE_SYNTHETIC=1 go test ./internal/fingerprint -run TestWriteSyntheticFixtures/grpc
WRITE_GOLDEN=1 go test ./internal/fingerprint -run TestWriteGoldenFixtures/grpc
```

## Run golden test

```sh
go test ./internal/fingerprint -run TestGoldenFixtures/grpc
```
