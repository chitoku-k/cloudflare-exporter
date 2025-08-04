cloudflare-exporter
===================

[![][workflow-badge]][workflow-link]

A Prometheus exporter for Cloudflare Load Balancers

## Requirements

- Go
- Cloudflare Load Balancing subscription

## API Access

For Cloudflare API, it is strongly recommended to create a dedicated Account API token or User API token.
Ensure the following permission is granted to the token:

- Account - Load Balancing: Monitors And Pools (Read)

## Installation

### Linux

```sh
$ docker buildx build .
```

```sh
# Port number (required)
export PORT=8080

# TLS certificate and private key (optional; if not specified, exporter is served over HTTP)
export TLS_CERT=/path/to/tls/cert
export TLS_KEY=/path/to/tls/key

# Cloudflare API Token (recommended, optional; either CF_API_TOKEN, or the combination of CF_API_KEY and CF_API_EMAIL is required)
export CF_API_TOKEN=

# Cloudflare API Key and API email (optional; either CF_API_TOKEN, or the combination of CF_API_KEY and CF_API_EMAIL is required)
export CF_API_KEY=
export CF_API_EMAIL=

# Cloudflare Account ID (optional)
export CF_ACCOUNT_ID=
```

### Windows

```powershell
> docker buildx build --file=Dockerfile.windows .
```

```powershell
# Port number (required)
$Env:PORT = 8080

# TLS certificate and private key (optional; if not specified, exporter is served over HTTP)
$Env:TLS_CERT = "\path\to\tls\cert"
$Env:TLS_KEY = "\path\to\tls\key"

# Cloudflare API Token (recommended, optional; either CF_API_TOKEN, or the combination of CF_API_KEY and CF_API_EMAIL is required)
$Env:CF_API_TOKEN = ""

# Cloudflare API Key and API email (optional; either CF_API_TOKEN, or the combination of CF_API_KEY and CF_API_EMAIL is required)
$Env:CF_API_KEY = ""
$Env:CF_API_EMAIL = ""

# Cloudflare Account ID (optional)
$Env:CF_ACCOUNT_ID = ""
```

## Usage

### Linux

```sh
$ ./cloudflare-exporter
```

### Windows

```powershell
> .\cloudflare-exporter.exe
```

## Prometheus Configuration

The Cloudflare exporter needs the name of pools to be passed which can be
configured by relabelling in a similar way to the blackbox exporter.

Example config:

```yaml
scrape_configs:
  - job_name: cloudflare
    static_configs:
      - targets:
        - pool01 # Name of Cloudflare pool to check for.
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: 127.0.0.1:8080 # The cloudflare exporter's real hostname:port.
```

Example result:

```
# HELP cloudflare_origin_health Result of health check
# TYPE cloudflare_origin_health gauge
cloudflare_origin_health{code="200",health_region="NEAS",origin_address="www01.example.com",pool_name="pool01"} 1
cloudflare_origin_health{code="503",health_region="NEAS",origin_address="www02.example.com",pool_name="pool01"} 0
# HELP cloudflare_origin_rtt_seconds RTT to the pool origin
# TYPE cloudflare_origin_rtt_seconds gauge
cloudflare_origin_rtt_seconds{code="200",health_region="NEAS",origin_address="www01.example.com",pool_name="pool01"} 0.0653
cloudflare_origin_rtt_seconds{code="503",health_region="NEAS",origin_address="www02.example.com",pool_name="pool01"} 0
```

### Spec

| Status | Condition                                |
|--------|------------------------------------------|
| 200    | Success.                                 |
| 400    | Target is not specified.                 |
| 500    | Unexpected error calling Cloudflare API. |

[workflow-link]:    https://github.com/chitoku-k/cloudflare-exporter/actions?query=branch:master
[workflow-badge]:   https://img.shields.io/github/actions/workflow/status/chitoku-k/cloudflare-exporter/ci.yml?branch=master&style=flat-square
