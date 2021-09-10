# syntax = docker/dockerfile:experimental
FROM golang:1.17.1-buster AS build
WORKDIR /usr/src
COPY go.mod go.sum /usr/src/
RUN --mount=type=cache,target=/go \
    go mod download
COPY . /usr/src/
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags='-s -w'

FROM scratch
ENV GIN_MODE release
ENV PORT 80
COPY --from=build /usr/src/cloudflare-exporter /cloudflare-exporter
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 80
CMD ["/cloudflare-exporter"]
