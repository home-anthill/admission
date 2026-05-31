# AGENTS.md

This file provides guidance to coding agents when working in this repository.

## Project Overview

Go microservice for device and sensor registration in the home-anthill IoT
platform. It exposes REST APIs with Gin, communicates with downstream services
over gRPC and HTTP, and uses MongoDB for persistence.

The module is `admission` and the Go baseline is `1.26.3`.

## Build & Development Commands

```bash
make build          # Generate protobuf code, run vet+lint, compile to ./build/admission
make run            # Hot-reload dev server via air
make test           # Generate protobuf code, run vet+lint, then run all tests with coverage
make proto          # Regenerate gRPC/protobuf code from .proto files
make lint           # Run staticcheck
make vet            # Run go vet + shadow analysis
make check          # Run govulncheck for security vulnerabilities
make deps           # Install dev tools, update modules, and tidy go.mod/go.sum
```

Run a single test:

```bash
ENV=testing go test -v -race ./... -run TestName
```

Integration and DB-backed tests require MongoDB running locally as a replica set
on port `27017`. The test database is `api-server-test`, selected when
`ENV=testing`. For local development, use the MongoDB replica set from
`../sharded-mongodb-compose/`.

`protoc` must be installed separately for `make proto`, `make build`, and
`make test`; `make deps` installs the Go protobuf plugins but not the `protoc`
binary.

## Environment Configuration

Copy `.env_template` to `.env` and customize:

- `HTTP_PORT` - Gin server port (default: `8099`)
- `GRPC_TLS` - enable TLS for gRPC dial (default: `false`; set to `true` in prod)
- `MONGODB_URL` - MongoDB connection string, for example `mongodb://localhost:27017/`
- `GRPC_URL` - downstream gRPC Registration service address, for example `localhost:50051`
- `HTTP_SENSOR_*` - HTTP sensor service details: base URL, port, and API paths
- `LOG_FOLDER` - directory for log files, created if missing
- `CERT_FOLDER_PATH` - path to TLS certs when `GRPC_TLS=true`
- `API_TOKEN_HASH_SECRET` - mandatory HMAC secret/pepper used to look up `profiles.apiTokenHash`; startup rejects values shorter than 32 characters

See `.env_template` for all variables and defaults.

## Architecture

- `api/` - HTTP handlers (`register.go`, `keepalive.go`) and gRPC proto definitions under `api/grpc/register/`
- `customerrors/` - standardized HTTP and gRPC error types
- `db/` - MongoDB connection setup, collection accessors, startup index creation, and transaction support
- `grpcutil/` - gRPC TLS config and secure dial-option builder
- `httputil/` - HTTP client helpers with a 10-second timeout and 64 KiB response body read cap
- `initialization/` - app startup, logger setup, env loading, Gin router config, and route wiring
- `models/` - persisted structures: `Device`, `Feature`, `Spec`, `Profile`, and `KeepAlive`
- `testutils/` - DB test helpers for collection lifecycle and document insertion
- `utils/` - generic helpers such as slice mapping/filtering, validation error formatting, and API-token hashing
- `validators/` - custom request validations, including feature spec validation
- `integration_tests/` - Ginkgo/Gomega BDD tests with mocked HTTP/gRPC servers

### HTTP Endpoints

- `POST /admission/register` - register a device with up to 16 features. A feature can be a `controller` or `sensor` and must include a validated `spec`.
- `GET /admission/keepalive` - health check endpoint. Returns `{"message":"ok"}`.

### Registration Flow

1. Bind and validate `DeviceRegisterReq`, including feature and spec rules.
2. Hash the supplied `apiToken` with `API_TOKEN_HASH_SECRET`.
3. Find the owning profile by `profiles.apiTokenHash`.
4. Check for an existing device by MAC before downstream side effects; duplicate MACs return `409`.
5. Build a `models.Device` with generated device/feature UUIDs and default `Name` set to the MAC address.
6. Call downstream gRPC `Registration.Register()` for controller features.
7. Call the HTTP sensor registration service for sensor features.
8. Insert the device and link it to the profile in a MongoDB transaction.
9. Return `DeviceRegisterRes` with public fields only.

### Feature Specs

Feature requests include `SpecReq`:

- `bool` specs must not include `min`, `max`, `step`, or `list`.
- `int` and `float` specs require finite `min`, `max`, and positive `step`; `max` must be greater than `min`, and `step` must divide the range.
- `list` specs require a non-empty list, reject numeric range fields, cap list items at 20, and require unique item values.

Custom spec validation is registered in `initialization.RegisterRoutes` through
`validators.RegisterValidations`.

## Key Patterns

- **Dependency injection**: handlers are constructor-injected with logger, MongoDB client, and validator. See `initialization/server.go:RegisterRoutes`.
- **Context propagation**: handlers use `c.Request.Context()` for DB queries, transactions, and gRPC calls.
- **Error handling**: domain errors are wrapped with `customerrors.ErrorWrapper` and returned as structured JSON with consistent HTTP status codes.
- **Validation**: struct tags use `go-playground/validator`; cross-field spec rules live in `validators/`.
- **Environment-driven config**: all runtime config comes from `.env`; `ENV=testing` switches to the test database and Gin TestMode.
- **gRPC**: controller registration uses a 5-second deadline per feature. TLS is controlled by `GRPC_TLS` and certs from `CERT_FOLDER_PATH`.
- **HTTP**: sensor registration uses shared helpers with timeout and capped response reads.
- **Profile API token lookup**: raw `apiToken` values are never queried directly. Hash with `API_TOKEN_HASH_SECRET` and query `profiles.apiTokenHash`; the secret must match `api-server`.
- **Feature fan-out limit**: `DeviceRegisterReq.Features` is capped at 16 before profile lookup or downstream calls.
- **MongoDB indexes**: startup creates unique indexes for `profiles.apiTokenHash` and `devices.mac`. Existing duplicate production data must be cleaned before rollout.
- **Duplicate registration**: duplicate MAC registration returns `409` before downstream calls. If the MAC belongs to another profile, the response remains generic and the device is not attached to the requester.

## Recent Changes

See `CHANGELOG.md` for release history.

Current highlights:

- **5.0.0**: adds device feature spec support and expands test coverage.
- **4.0.0**: adds default device `name`, hardens profile token lookup, enforces MongoDB uniqueness, improves context propagation, caps HTTP response reads, and splits `grpcutil/` and `httputil/` out of `utils/`.

## CI/CD

GitHub Actions (`.github/workflows/docker-image.yml`) runs `make deps`,
`make vet`, `make lint`, `make test`, then builds and pushes the Docker image to
`docker.io/ks89/admission`. It triggers on pushes to `master`, `develop`, and
`ft**` branches, and on pull requests.

The Docker image uses a hardened runtime base, runs as non-root user UID `65534`,
and includes a pre-created `/logs` directory.
