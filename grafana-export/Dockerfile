# syntax = docker/dockerfile:1
FROM golang:1.25.1 AS base
WORKDIR /usr/src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go \
    go mod download
COPY . ./

FROM base AS build
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags='-s -w'

FROM base AS dev
COPY --from=golangci/golangci-lint /usr/bin/golangci-lint /usr/bin

FROM mcr.microsoft.com/playwright:v1.52.0-noble
COPY --link --from=build /usr/src/grafana-export /usr/local/bin/grafana-export
CMD ["/usr/local/bin/grafana-export"]
