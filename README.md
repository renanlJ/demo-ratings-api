# demo-ratings-api

Product ratings microservice for the **Demo Commerce Platform** — a
polyglot reference workload for devex.io (Backstage-based IDP). Returns
1–5 star averages, counts and per-star distributions per product.

- **Language**: Go 1.27 (standard library only — `net/http`, `log/slog`)
- **Role**: leaf service consumed by `demo-product-api`
- **Part of**: `system: demo-commerce` / `domain: commerce` (see `catalog-info.yaml`)

## API

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | `/api/v1/ratings` | Rating summaries for all products |
| GET | `/api/v1/ratings/{productId}` | Rating summary for one product (404 if unknown) |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness (503 when `FAIL_READINESS=true`) |
| GET | `/metrics` | Prometheus text exposition |

OpenAPI contract: [`docs/api/openapi.yaml`](docs/api/openapi.yaml)

Sample SKUs: `SKU-1001`, `SKU-1002`, `SKU-1003`.

## Configuration

| Env var | Default | Description |
| ------- | ------- | ----------- |
| `PORT` | `8083` | Listen port |
| `LOG_LEVEL` | `info` | `debug`/`info`/`warn`/`error` (JSON logs) |
| `FAIL_READINESS` | `false` | Force `/readyz` → 503 (governance scenario) |

## Develop

```bash
make install      # go mod download (no external deps)
make test         # go test ./...
make coverage     # coverage report
make lint format  # go vet + gofmt check
make run          # go run ./cmd/ratings-api
```

Requires Go 1.27+. No third-party dependencies — `go.mod` is the whole
dependency tree.

## Docker

```bash
make docker-build     # distroless-based image, non-root
docker run -p 8083:8083 demo-ratings-api:0.1.0
```

## Kubernetes

```bash
kubectl kustomize deploy/kubernetes          # render
kubectl apply -k deploy/kubernetes           # deploy (per-repo)
```

The platform-wide deployment lives in `demo-product-api/deploy/platform`.

## Observability

- Structured JSON logs (`log/slog`) with `request_id` and W3C `trace_id`
  parsed from `traceparent`/`X-Request-ID` headers.
- Prometheus counters/histograms on `/metrics`: request count, duration,
  status codes.

## Governance

Scenario recipes (health-check failure, coverage gate, deploy failure,
…) live in `demo-product-api/docs/governance/scenarios.md`.

## CI

`.github/workflows/ci.yml`: lint+tests+coverage → SAST/SCA/secrets →
build → image build+scan → GHCR publish on `v*` tags.
