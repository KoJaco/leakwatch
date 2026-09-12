# nethttp

HTTP handler leak generator for integration testing and golden fixture recording.

## Run

```sh
go run -tags integration ./test/integration/nethttp
```

App listens on `:8080`. pprof on `http://127.0.0.1:6060/debug/pprof/goroutineleak`.

## Record fixture

Trigger leaks, then capture after GC:

```sh
curl http://127.0.0.1:8080/
curl -o internal/fingerprint/testdata/nethttp/leak.pb.gz \
  http://127.0.0.1:6060/debug/pprof/goroutineleak
```

Regenerate expected clusters:

```sh
WRITE_GOLDEN=1 go test ./internal/fingerprint -run TestWriteGoldenFixtures/nethttp
```
