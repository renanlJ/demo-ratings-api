# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-11-24

### Added

- `GET /api/v1/ratings` and `GET /api/v1/ratings/{productId}` endpoints backed
  by deterministic in-memory fixtures.
- `/healthz`, `/readyz` and `/metrics` platform endpoints (`FAIL_READINESS`
  env var forces readiness failure for governance scenarios).
- Structured JSON logging (`log/slog`) with `request_id` and `trace_id`
  correlation fields; W3C `traceparent` propagation.
- Prometheus text-format metrics (standard library only, zero external
  dependencies).
- Multi-stage distroless Dockerfile, Kubernetes manifests
  (`deploy/kubernetes`) and Backstage catalog descriptor (`catalog-info.yaml`).
- GitHub Actions CI pipeline (`.github/workflows/ci.yml`).
