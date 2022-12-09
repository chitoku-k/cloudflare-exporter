# syntax = docker/dockerfile:experimental
FROM golang:1.19.4-buster AS build
WORKDIR /usr/src
COPY go.mod go.sum /usr/src/
RUN --mount=type=cache,target=/go \
    go mod download
COPY . /usr/src/
RUN --mount=type=cache,target=/go \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags='-s -w'

FROM scratch
ARG PORT=80
ENV PORT=$PORT
ENV GIN_MODE=release
COPY --from=build /usr/src/cloudflare-exporter /cloudflare-exporter
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE $PORT
CMD ["/cloudflare-exporter"]
