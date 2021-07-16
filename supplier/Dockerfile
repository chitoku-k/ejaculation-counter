# syntax = docker/dockerfile:experimental
FROM golang:1.16.6-buster AS build
WORKDIR /usr/src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go \
    go mod download
COPY . ./
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags='-s -w'

FROM scratch AS production
ENV GIN_MODE release
ENV PORT 80
COPY --from=build /usr/src/supplier /supplier
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
EXPOSE 80
CMD ["/supplier"]
