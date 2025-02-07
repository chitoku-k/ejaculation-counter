# syntax = docker/dockerfile:1
FROM golang:1.23.6 AS base
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

FROM scratch AS production
ARG PORT=80
ENV PORT=$PORT
ENV GIN_MODE=release
COPY --link --from=build /lib/x86_64-linux-gnu/ld-linux-x86-64.* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib/x86_64-linux-gnu/libc.so* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib/x86_64-linux-gnu/libresolv.so* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib64 /lib64
COPY --link --from=build /usr/src/reactor /reactor
COPY --link --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --link --from=build /usr/share/zoneinfo /usr/share/zoneinfo
EXPOSE $PORT
CMD ["/reactor"]
