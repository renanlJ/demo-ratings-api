# syntax=docker/dockerfile:1
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ratings-api ./cmd/ratings-api

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.title="demo-ratings-api" \
      org.opencontainers.image.description="Product ratings service - Demo Commerce Platform" \
      org.opencontainers.image.source="https://github.com/renanlJ/demo-ratings-api" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /out/ratings-api /ratings-api
EXPOSE 8083
USER nonroot
ENTRYPOINT ["/ratings-api"]
