FROM golang:1.16.4-buster AS dev
WORKDIR /usr/src/app
CMD ["bash", "-c", "go build -o /usr/bin && exec /usr/bin/supplier"]

FROM golang:1.16.4-buster AS build
WORKDIR /usr/src
COPY . /usr/src
RUN CGO_ENABLED=0 go build -ldflags='-s -w'

FROM scratch AS production
ENV GIN_MODE release
ENV PORT 80
COPY --from=build /usr/src/supplier /supplier
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
EXPOSE 80
CMD ["/supplier"]
