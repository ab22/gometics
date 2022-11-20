# Go Metrics

## Pre-Requisites
- Docker & Docker Compose
- `golangci-lint`

## Commands

```sh
# Lint go files.
make lint

# Build and download docker images.
make build

# Start docker containers.
make up

# Stop docker containers.
make down
```

## Endpoints

### Go Metrics

URL: `<none yet>`

### Prometheus

URL: `http://localhost:9090`

### Grafana

URL: `http://localhost:3000/`
Default User: admin
Default Pass: admin
