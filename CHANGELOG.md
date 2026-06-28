# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-06-27

### Added
- `LookupMany` enriches many emails concurrently while preserving input order.
  A failure on one email is reported in that result's `Err` field instead of
  aborting the batch, and concurrency is configurable via `WithConcurrency`.
- Automatic pagination via Go 1.23 range-over-func iterators: `IterRuns`,
  `IterRunResults`, `IterAllRuns`, and `IterAllResults` stream every page on
  demand.

### Changed
- README examples now use recognizable addresses instead of placeholders.

### Removed
- Dropped a leftover `errors` import keep-alive that no longer served a purpose.

## [0.3.0] - 2026-06-25

### Added
- Initial release of the `Encrata` client covering email intelligence
  (`Lookup`, `Validate`, `Breaches`), monitors, and contact lists.
- Functional options for configuration: `WithBaseURL`, `WithHTTPClient`,
  `WithTimeout`, and `WithMaxRetries`.
- Automatic retries for transient failures (HTTP 429, 500, 502, 503, 504,
  connection errors, and timeouts) using full-jitter exponential backoff with
  `Retry-After` support.
- Typed error hierarchy (`AuthenticationError`, `InsufficientCreditsError`,
  `InvalidRequestError`, `RateLimitError`, `APIConnectionError`, `APIError`)
  inspectable with `errors.As`.

[0.4.0]: https://github.com/Encratahq/encrata-go/releases/tag/v0.4.0
[0.3.0]: https://github.com/Encratahq/encrata-go/releases/tag/v0.3.0
