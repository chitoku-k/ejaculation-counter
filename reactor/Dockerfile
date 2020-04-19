FROM golang:1.14-stretch as build
WORKDIR /usr/src
COPY . /usr/src
RUN CGO_ENABLED=0 go build -ldflags='-s -w'

FROM scratch
ENV GIN_MODE release
ENV PORT 80
COPY --from=build /usr/src/reactor /reactor
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 80
CMD ["/reactor"]
