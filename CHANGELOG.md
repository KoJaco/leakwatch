# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-09-13

First public release. Requires Go 1.27+ and the runtime `goroutineleak` pprof
profile.

### Added

- Library-first goroutine leak watcher with scheduled sampling, fingerprinting,
  temporal analysis, and export surfaces (Prometheus, HTTP debug, JSON).
- Goroutine leak profile fetch (`PProfSource`) and parse (`DefaultParser`) for
  Go 1.27 `goroutineleak` pprof profiles.
- Stack fingerprinting: normalization, projection, clustering, and stable
  `leak_id` generation (SHA-256, 16 hex chars).
- Temporal analysis: `Analyze` with status classification (`new`, `growing`,
  `persistent`, `recurring`), plus `GrowthFor`, `DetectRecurrence`, and
  `RankBySeverity`.
- Prometheus metrics including `goroutine_leak_growth_rate` and
  `goroutine_leak_fingerprint_version`.
- CLI: `leakwatch analyze` (offline profile) and `leakwatch inspect` (live
  pprof URL) with optional `-json` output and `--allow-remote`.
- HTTP debug middleware hook (`httpexport.WithMiddleware`) and GET-only
  enforcement on debug handlers.
- Watcher options: `WithHistoryCapacity`, `WithHTTPTimeout`, `WithMaxProfileBytes`,
  `WithHTTPClient`, `WithLogger`.
- `Watcher.LastSampleError()` for health checks after transient sample failures.
- Golden fixtures and e2e integration tests for channels, nethttp, grpc, and
  workers leak patterns.
- Dependabot configuration and CI checks (race detector on integration tests,
  govulncheck).
- `SECURITY.md` and launch documentation.

### Changed

- Profile parsing uses count-weighted samples instead of expanding goroutine
  structs.
- Observations use `prof.CapturedAt` when available.
- Scheduler logs and retries transient sample errors with exponential backoff.
- `History` uses `sync.RWMutex` for concurrent reads during sampling.
- Jitter uses uniform random delay via `math/rand/v2`.

### Fixed

- Data race on observation history during concurrent `Snapshot` and `Sample`.
- Unbounded HTTP response reads and missing default fetch timeout.
- Wrong pprof profile types are rejected with `ErrWrongProfileType`.

### Notes

- `leakwatch inspect` accepts localhost URLs only unless `--allow-remote` is
  passed.
- Each sample triggers leak-detection GC — see [docs/sampling.md](docs/sampling.md).
- Debug endpoints expose stack traces — bind to loopback or use middleware.
