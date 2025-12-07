# syntax = docker/dockerfile:1
FROM golang:1.25.5 AS base
WORKDIR /usr/src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go \
    go mod download
COPY . ./

FROM base AS build
ARG VERSION=v0.0.0-dev
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w -X main.version=$VERSION"

FROM scratch AS production-linux-amd64
ARG DEB_BUILD_MULTIARCH=x86_64-linux-gnu
ARG LD=/lib64/ld-linux-x86-64.so.2

FROM scratch AS production-linux-arm64
ARG DEB_BUILD_MULTIARCH=aarch64-linux-gnu
ARG LD=/lib/ld-linux-aarch64.so.1

FROM production-${TARGETPLATFORM//\//-} AS production
ARG PORT=80
ENV PORT=$PORT
ENV GIN_MODE=release
COPY --link --from=build $LD $LD
COPY --link --from=build /lib/$DEB_BUILD_MULTIARCH/ld-linux-x86-64.* /lib/$DEB_BUILD_MULTIARCH/
COPY --link --from=build /lib/$DEB_BUILD_MULTIARCH/libc.so* /lib/$DEB_BUILD_MULTIARCH/
COPY --link --from=build /lib/$DEB_BUILD_MULTIARCH/libresolv.so* /lib/$DEB_BUILD_MULTIARCH/
COPY --link --from=build /usr/src/reactor /reactor
COPY --link --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --link --from=build /usr/share/zoneinfo /usr/share/zoneinfo
EXPOSE $PORT
RUN ["/reactor", "--version"]
CMD ["/reactor"]
