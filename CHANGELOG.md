# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- HTTP debug middleware hook (`httpexport.WithMiddleware`) and GET-only enforcement on debug handlers.
- Watcher options: `WithHistoryCapacity`, `WithHTTPTimeout`, `WithMaxProfileBytes`, `WithHTTPClient`, `WithLogger`.
- `Watcher.LastSampleError()` for health checks after transient sample failures.
- Prometheus metric `goroutine_leak_fingerprint_version`.
- CLI `inspect --allow-remote` flag; localhost URLs only by default.
- Dependabot configuration for Go modules and GitHub Actions.
- CI: `govulncheck` and race detector on integration tests.

### Changed

- `leak_id` now uses SHA-256 truncated to 16 hex characters (was 8-char FNV-32).
- Profile parsing uses count-weighted samples instead of expanding goroutine structs.
- Observations use `prof.CapturedAt` when available.
- Scheduler logs and retries transient sample errors with exponential backoff.
- `History` uses `sync.RWMutex` for concurrent reads during sampling.
- Jitter uses uniform random delay via `math/rand/v2`.

### Fixed

- Data race on observation history during concurrent `Snapshot` and `Sample`.
- Unbounded HTTP response reads and missing default fetch timeout.
- Wrong pprof profile types are rejected with `ErrWrongProfileType`.

### Breaking

- `leak_id` format changed; Prometheus series reset on upgrade.
- `inspect` rejects non-localhost URLs unless `--allow-remote` is passed.
- Invalid (non-`goroutineleak`) profiles now error instead of silently mis-parsing.

### Added (initial release)

- Initial project scaffold with domain model, pipeline packages, and stub implementations.
- Goroutine leak profile fetch (`PProfSource`) and parse (`DefaultParser`) for Go 1.27 `goroutineleak` pprof profiles.
- Stack fingerprinting: normalization, projection, clustering, and stable `leak_id` generation from goroutine stacks.
- Temporal analysis: `Analyze` interprets observation history into export-ready leaks with status classification (`new`, `growing`, `persistent`, `recurring`), plus `GrowthFor`, `DetectRecurrence`, and `RankBySeverity` helpers.
- Prometheus growth metric: `goroutine_leak_growth_rate` derived from observation history at scrape time via `GrowthFor`.
- CLI: `leakwatch analyze` (offline profile) and `leakwatch inspect` (live pprof URL) with optional `-json` output.
- Golden fixtures: committed `leak.pb.gz` and `clusters.json` for channels, nethttp, grpc (synthetic), and workers leak patterns; offline golden tests in `internal/fingerprint` and `internal/profile`.
- E2E integration tests: `test/e2e/` with live leak generators, Watcher assertions, and CI job (`-tags integration`); real gRPC generator added.
