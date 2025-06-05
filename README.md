# pet-two-phase-commit

[![build](https://github.com/Sugar-pack/orders-manager/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/Sugar-pack/orders-manager/actions/workflows/build.yml)
[![CodeQL](https://github.com/Sugar-pack/orders-manager/actions/workflows/codeql.yml/badge.svg)](https://github.com/Sugar-pack/orders-manager/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/Sugar-pack/orders-manager/branch/main/graph/badge.svg?token=VEXDJ58WWI)](https://codecov.io/gh/Sugar-pack/orders-manager)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Go](https://golang.org/dl/)
- [protoc](https://github.com/protocolbuffers/protobuf/releases)

## Development

```bash
go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go-grpc_out=. api/api.proto
```
## Launch

```bash
docker-compose up --build -d --remove-orphans
```

### Tracing

#### UI

After successfull launch tracing UI will be available on address http://localhost:16686/

### Known problems
if api-service didn't start, just restart it
```bash
docker-compose up -d
```

### config filename

Service uses **config.yml** as a configuration file name. It can not be overriden. Probably, should be passed as a command line argument.
