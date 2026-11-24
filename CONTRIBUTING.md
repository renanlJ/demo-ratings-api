# Contributing

## Development setup

Requires Go >= 1.27. No external dependencies — the service intentionally uses
only the Go standard library.

```sh
go build ./...
go test ./...
```

`make` targets wrap the same commands (`make test`, `make coverage`,
`make docker-build`) but a local `make` is not required; run the underlying
commands directly.

## Conventions

- Format code with `gofmt` and keep `go vet` clean before pushing.
- Keep the service dependency-free. New `go.mod` requirements need a strong
  justification and maintainer discussion.
- Commits: imperative mood, concise subject line (e.g. `Add p99 latency bucket`).
- Update `CHANGELOG.md` under `Unreleased` for user-visible changes.

## Pull requests

- CI must be green: `gofmt -l`, `go vet`, `go test ./...`, coverage, Trivy
  image scan.
- Keep PRs focused; link related issues.
