# grpc

gRPC leak generator for E2E tests and manual fixture recording.

## Run

```sh
go run -tags integration ./test/integration/grpc
```

Starts gRPC on `LEAKWATCH_GRPC_ADDR` (default `:50051`) and pprof on
`LEAKWATCH_PPROF_ADDR` (default `:6060`).

Each `Leak` RPC blocks the handler goroutine with gRPC frames visible in the
leaked stack.

## Environment

| Variable | Default |
|----------|---------|
| `LEAKWATCH_GRPC_ADDR` | `:50051` |
| `LEAKWATCH_PPROF_ADDR` | `:6060` |

## E2E

Covered by `TestGRPCLeak` in [`test/e2e`](../e2e/).

The golden fixture under `internal/fingerprint/testdata/grpc/` remains
**synthetic** (representative stack frames). Live re-recording is optional.
