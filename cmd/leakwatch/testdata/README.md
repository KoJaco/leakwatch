# CLI test fixture

`leak.pb` is a minimal Go 1.27 `goroutineleak` count profile with three leaked
goroutines sharing a channel-leak stack (`main.leak` blocked on `runtime.chanrecv1`).

Regenerate:

```sh
WRITE_FIXTURE=1 go test ./cmd/leakwatch -run TestWriteFixtureProfile
```

The profile is serialized with `github.com/google/pprof/profile` (raw protobuf,
same format the runtime writes).
