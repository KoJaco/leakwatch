# basic

Minimal channel-leak generator for integration testing and E2E tests.

## Run

```sh
go run -tags integration ./test/integration/basic
```

Starts pprof on `LEAKWATCH_PPROF_ADDR` (default `:6060`).

## Environment

| Variable | Default |
|----------|---------|
| `LEAKWATCH_PPROF_ADDR` | `:6060` |

## E2E

Covered by `TestBasicLeak` in [`test/e2e`](../e2e/).
