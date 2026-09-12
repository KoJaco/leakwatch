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
A live gRPC integration generator is deferred to phase 7.

## Expected cluster

- **ID:** `6760f943`
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
