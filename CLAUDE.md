# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go microservice for device/sensor registration in the home-anthill IoT platform. Exposes REST APIs (Gin framework) and communicates with other services via gRPC. Uses MongoDB for persistence.

## Build & Development Commands

```bash
make build          # Generate proto, run vet+lint, compile to ./build/admission
make run            # Hot-reload dev server via air
make test           # Run integration tests with coverage (requires MongoDB replica set)
make proto          # Regenerate gRPC/protobuf code from .proto files
make lint           # Run staticcheck
make vet            # Run go vet + shadow analysis
make check          # Run govulncheck for security vulnerabilities
make deps           # Install all dev tools (shadow, staticcheck, air, govulncheck, protoc plugins, go-cover-treemap)
```

**Run a single test:**
```bash
ENV=testing go test -v -race ./integration_tests -run TestName
```

**Tests require** MongoDB running locally as a **replica set** (port 27017). The test database is `api-server-test` (selected when `ENV=testing`). For local dev, use the MongoDB replica set from `../sharded-mongodb-compose/`.

## Environment Configuration

Copy `.env_template` to `.env` and customize:
- `HTTP_PORT` — Gin server port (default: 8099)
- `GRPC_TLS` — Enable TLS for gRPC dial (default: false; set to `true` in prod)
- `MONGODB_URL` — MongoDB connection string (e.g., `mongodb://localhost:27017/`)
- `GRPC_URL` — Downstream gRPC Registration service address (e.g., `localhost:50051`)
- `HTTP_SENSOR_*` — HTTP sensor service details (base URL, port, API paths)
- `LOG_FOLDER` — Directory for log files (created if missing)
- `CERT_FOLDER_PATH` — Path to TLS certs when `GRPC_TLS=true`

See `.env_template` for all variables and defaults.

## Architecture

- **api/** — HTTP handlers (`register.go`: device registration; `keepalive.go`: device heartbeat) and gRPC proto definitions (`api/grpc/register/`)
- **db/** — MongoDB connection, collection accessors, startup index creation, transaction support
- **models/** — Data structures: `Device` (sensors/controllers), `Profile` (user account), `KeepAlive` (device heartbeat status)
- **initialization/** — App startup: logger setup, env loading, Gin router config, dependency injection
- **customerrors/** — Standardized error responses and gRPC error types
- **grpcutil/** — gRPC TLS config and secure dial-option builder
- **httputil/** — HTTP client with timeouts (10s default) and capped response reads; `Get` and `Post` helpers
- **utils/** — Validation helpers, generic slice utilities (filter, map)
- **integration_tests/** — Ginkgo/Gomega BDD tests with mocked HTTP/gRPC servers
- **testutils/** — DB test helpers (collection lifecycle, document insertion)

### HTTP Endpoints

- `POST /admission/register` — Register a device with one or more features (`controller` and/or `sensor`). Validates `DeviceRegisterReq`, queries MongoDB, calls downstream gRPC `Registration.Register()` for controllers and the HTTP sensor registration service for sensors.
- `GET /admission/keepalive` — Health check endpoint. Returns `{"message":"ok"}`.

### Request Flow

1. **Register endpoint**: Validates JSON body (`DeviceRegisterReq`) and finds the owning profile by `apiToken`
2. Checks for an existing device by MAC before downstream side effects; duplicate MACs return `409`
3. Calls downstream gRPC service (`Registration.Register()` with 5-second deadline per call) for controller features
4. Calls HTTP sensor service to register sensor features
5. Inserts the device and links it to the profile in a MongoDB transaction
6. Returns `DeviceRegisterRes` (public fields only: UUID, MAC, manufacturer, model, features)

### Key Patterns

- **Dependency injection**: Handlers are constructor-injected with logger, DB client, and validator. See `initialization/server.go:RegisterRoutes`.
- **Context propagation**: Handlers use request context (`c.Request.Context()`) for all DB queries and gRPC calls, respecting client cancellation and timeouts.
- **Error handling**: Errors are wrapped via `customerrors.ErrorWrapper` and returned as structured HTTP JSON responses with consistent status codes.
- **Validation**: Struct tags with `go-playground/validator` (e.g., `validate:"required,uuid4,mac"`). Custom error messages via `utils.GetErrorMessage`.
- **Environment-driven**: All config via `.env` (no hardcoded values). `ENV=testing` switches to test database and Gin TestMode.
- **gRPC**: Calls use per-request 5-second deadline. TLS toggled via `GRPC_TLS` env var; certs from `CERT_FOLDER_PATH` when enabled.
- **HTTP**: Downstream HTTP calls use a shared client with 10-second timeout and 64 KiB response body read cap to prevent goroutine and memory exhaustion.
- **MongoDB indexes**: Startup creates unique indexes for `profiles.apiToken` and `devices.mac`. Existing duplicate production data must be cleaned before rollout because index creation will fail on duplicates.
- **Duplicate registration**: Duplicate MAC registration returns `409` before downstream calls. If the MAC belongs to another profile, the response remains generic and the device is not attached to the requester.

## Recent Refactoring (See `CHANGELOG_CLAUDE.md`)

Recent major changes:
- **Security hardening**: Go baseline upgraded to 1.26.2; `govulncheck ./...` should report no vulnerabilities with that toolchain
- **Uniqueness enforcement**: MongoDB unique indexes for `profiles.apiToken` and `devices.mac`, with duplicate-key races mapped to `409`
- **Duplicate ownership protection**: Cross-profile attempts to register an existing MAC return generic `409` and do not attach the device
- **HTTP response cap**: Downstream HTTP helper response reads are capped at 64 KiB
- **Package split**: `utils/grpc.go` → `grpcutil/grpc.go`, `utils/http.go` → `httputil/http.go` for better organization
- **Context propagation**: Handlers now use request context instead of long-lived background context
- **HTTP timeouts**: Downstream HTTP calls use a 10-second timeout to prevent goroutine leaks
- **Security**: gRPC TLS minimum version set to 1.3; MongoDB URL redacted from logs; input validation tightened (alphanum constraint on feature names)
- **Error handling**: Library functions now return errors instead of panicking; standardized error responses via `ErrorWrapper`
- **Bug fixes**: Timeout allocation per gRPC call (not shared), response status code checks, validation of zero-value bool fields

For details, see `CHANGELOG_CLAUDE.md`.

## CI/CD

GitHub Actions (`.github/workflows/docker-image.yml`): runs `make deps`, `make vet`, `make lint`, `make test`, then builds and pushes Docker image to `docker.io/ks89/admission`. Triggers on pushes to `master`, `develop`, `ft**` branches and PRs.

Docker image uses hardened base (`dhi.io/alpine-base`), runs as non-root user (UID 65534), and includes a pre-created `/logs` directory.
