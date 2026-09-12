# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial project scaffold with domain model, pipeline packages, and stub implementations.
- Goroutine leak profile fetch (`PProfSource`) and parse (`DefaultParser`) for Go 1.27 `goroutineleak` pprof profiles.
- Stack fingerprinting: normalization, projection, clustering, and stable `leak_id` generation from goroutine stacks.
- Temporal analysis: `Analyze` interprets observation history into export-ready leaks with status classification (`new`, `growing`, `persistent`, `recurring`), plus `GrowthFor`, `DetectRecurrence`, and `RankBySeverity` helpers.
- Prometheus growth metric: `goroutine_leak_growth_rate` derived from observation history at scrape time via `GrowthFor`.
- CLI: `leakwatch analyze` (offline profile) and `leakwatch inspect` (live pprof URL) with optional `-json` output.
- Golden fixtures: committed `leak.pb.gz` and `clusters.json` for channels, nethttp, grpc (synthetic), and workers leak patterns; offline golden tests in `internal/fingerprint` and `internal/profile`.
