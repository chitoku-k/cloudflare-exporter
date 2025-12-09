# syntax = docker/dockerfile:1
FROM golang:1.25.5 AS base
WORKDIR /usr/src
COPY go.mod go.sum /usr/src/
RUN --mount=type=cache,target=/go \
    go mod download
COPY . /usr/src/

FROM base AS build
ARG VERSION=v0.0.0-dev
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w -X main.version=$VERSION"

FROM base AS dev
COPY --from=golangci/golangci-lint /usr/bin/golangci-lint /usr/bin

FROM scratch
ARG PORT=80
ENV PORT=$PORT
ENV GIN_MODE=release
COPY --link --from=build /lib/x86_64-linux-gnu/ld-linux-x86-64.* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib/x86_64-linux-gnu/libc.so* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib/x86_64-linux-gnu/libresolv.so* /lib/x86_64-linux-gnu/
COPY --link --from=build /lib64 /lib64
COPY --link --from=build /usr/src/cloudflare-exporter /cloudflare-exporter
COPY --link --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE $PORT
RUN ["/cloudflare-exporter", "--version"]
CMD ["/cloudflare-exporter"]
