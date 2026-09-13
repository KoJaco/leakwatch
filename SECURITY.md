# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.1.x   | Yes       |

leakwatch requires **Go 1.27 or newer** and the runtime `goroutineleak` pprof
profile. Security fixes apply only to supported release lines on supported Go
versions.

## Reporting a Vulnerability

Please report security issues **privately** — do not open a public GitHub issue
for vulnerabilities.

1. Use [GitHub Security Advisories](https://github.com/KoJaco/leakwatch/security/advisories/new) to report a vulnerability on this repository.
2. Include a clear description, reproduction steps, and impact assessment.

We will acknowledge reports as quickly as possible and coordinate disclosure
and a fix before public release.

## Operational Security

leakwatch is a diagnostics library, not an authentication system. When deploying
it, keep the following in mind:

### Debug HTTP handler

`Watcher.ServeDebug` and `export/http.Register` expose **full goroutine stack
traces** (function names, file paths, line numbers). Treat these endpoints as
sensitive operational data.

- Bind to loopback (`127.0.0.1`) or a private network unless access is restricted.
- Never expose on a public HTTP mux without protection. Use
  `httpexport.WithMiddleware` for auth, IP allowlists, or reverse-proxy checks.

### CLI `inspect`

`leakwatch inspect` fetches profile data over HTTP. By default it accepts
**localhost URLs only**. Pass `--allow-remote` only for trusted URLs — the CLI
can reach internal services if misused (SSRF risk in automation).

### Profile fetch limits

The library bounds HTTP response size and applies fetch timeouts by default.
Override with `WithMaxProfileBytes` and `WithHTTPTimeout` only when you
understand the memory and timeout implications.

## Security Updates

Security fixes are released as patch versions on the supported line (e.g.
`v0.1.1`). Breaking security-related behaviour changes are documented in
[CHANGELOG.md](CHANGELOG.md).
