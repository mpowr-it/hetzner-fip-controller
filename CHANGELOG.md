# Changelog
All notable changes to this project will be documented in this file.

This changelog follows [Semantic Versioning](https://semver.org/) and the
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/) conventions.

Older upstream changes from the original `cbeneke/hcloud-fip-controller`
project have intentionally been removed. This fork starts with
`v1.0.0-RC.1`, representing the first maintained and enhanced release under
the `mpowr-it/hetzner-fip-controller` project.

---

## [v1.0.0-RC.1] – 2025-11-17
### Added
- Initial maintained fork published under `mpowr/hetzner-fip-controller`.
- GitHub Actions pipeline for CI:
  - Go build & tests on all branches.
  - Docker build validation on branch pushes.
- GitHub Actions Release pipeline:
  - Multi-arch Docker builds (amd64, arm64).
  - Docker Hub push with semantic version tagging.
  - Automatic `latest` tag on release.

### Changed
- Removed legacy upstream CHANGELOG and all non-maintained version history.
- Replaced upstream DockerHub integration with a new namespace
  (`mpowr/hetzner-fip-controller`).
- Updated repository structure and documentation to reflect active maintenance.

### Fixed
- Eliminated hard process termination caused by transient Hetzner Cloud API
  `503 Service Unavailable` errors (replaced `log.Fatal` crash behavior with
  error logging and retry-friendly flow).

### Improved
- Modernized codebase scaffolding to support future enhancements such as:
  - health/readiness endpoints,
  - Prometheus metrics,
  - improved retry/backoff logic,
  - future controller-runtime migration.

---

## [Upcoming]
(Planned changes for the next release)

### Added
- Health and readiness HTTP endpoints for liveness/readiness probes.
- Retry/backoff subsystem for all Hetzner API calls.
- Prometheus metrics endpoint (`/metrics`).

### Changed
- Align Dockerfile with current Go version (1.22+) and multi-stage build improvements.

---

