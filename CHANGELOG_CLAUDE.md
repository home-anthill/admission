# Changelog

## 2026-04-24 Security Hardening

- **Go toolchain vulnerabilities fixed** (`go.mod`, `Dockerfile`): Raised the project Go baseline and Docker builder image from Go 1.26.0/1.26 to Go 1.26.2. `govulncheck ./...` now reports no vulnerabilities.
- **Duplicate MAC ownership protection** (`api/register.go`): Registration now checks for an existing device by MAC before downstream gRPC/HTTP calls. Duplicate registrations return `409` with the existing generic `{"message":"Already registered"}` response.
- **Cross-profile device attachment prevented** (`api/register.go`): If a valid profile token tries to register a MAC already owned by another profile, the service returns `409` and does not add that device to the requesting profile.
- **Duplicate-key race handling** (`api/register.go`): MongoDB duplicate-key errors during insert are now mapped to `409`, covering concurrent registration races that pass the initial read check.
- **MongoDB uniqueness enforced at startup** (`db/database.go`): Startup now creates unique indexes for `profiles.apiTokenHash` and `devices.mac`. Deployment note: existing duplicate data must be cleaned before rollout, otherwise index creation will fail and startup will stop.
- **Profile API token lookup hardened** (`api/register.go`, `utils/api_token_crypto.go`): Registration now hashes the incoming profile token with mandatory `API_TOKEN_HASH_SECRET` and queries `profiles.apiTokenHash` instead of querying plaintext `profiles.apiToken`.
- **Downstream HTTP response reads capped** (`httputil/http.go`): `Get` and `Post` now read at most 64 KiB from downstream response bodies, preventing memory exhaustion from large responses.
- **Regression coverage added** (`integration_tests/register_test.go`, `httputil/http_test.go`): Added tests for cross-profile duplicate registration and capped HTTP response bodies.

## Bug Fixes

- **`Enable: false` rejected by validation** (`api/register.go`): Removed `required` tag on `FeatureReq.Enable` bool field. The `required` validator rejects the zero value (`false`), making it impossible to register a device with `Enable: false`.
- **`GetErrorMessage` panic on unexpected error type** (`utils/validator.go`): Replaced bare type assertion `err.(validator.ValidationErrors)` with `errors.As`, preventing a panic if the error is not a `ValidationErrors`.
- **HTTP response status codes ignored** (`api/register.go`): `registerSensorsViaHTTP` now checks the status code returned by `utils.Post` and the keep-alive call. Previously, a 4xx/5xx from the sensor service was silently treated as success.
- **`io.ReadAll` errors discarded** (`httputil/http.go`): Both `Get` and `Post` now return an error if reading the response body fails.
- **`logger.Sync()` deferred in wrong scope** (`initialization/start.go`): `defer logger.Sync()` was inside `Start()`, so it ran when `Start` returned — not when the program exited. Removed the defer so the caller controls the logger lifecycle.
- **Shared gRPC timeout across all controllers** (`api/register.go`): `registerControllersViaGRPC` used a single 1-second `context.WithTimeout` for all sequential `Register` calls. Later calls could have nearly zero time left. Now each call gets its own 5-second deadline.
- **Dead return values in `registerControllersViaGRPC`** (`api/register.go`): Signature returned `(string, string, error)` but always returned `"", ""`. Simplified to return only `error`.
- **Dead package-level variables**: Removed unused `var client` in `db/database.go` and package-level `register`/`keepAlive` in `initialization/server.go`.
- **Unused `ctx` field in `KeepAlive` struct** (`api/keepalive.go`): Removed the stored context and unused logger, since `GetKeepAlive` never used them.
- **`os.Getwd()` error silently discarded** (`initialization/environment.go`): `readEnv` now returns the error from `os.Getwd()` instead of ignoring it with `_`.
- **Dead `panic` after `logger.Fatalf`** (`db/database.go`): `Fatalf` calls `os.Exit(1)`, so the `panic` on the next line was unreachable dead code. Removed both — `InitDb` now returns errors instead.


## Security Fixes

- **HTTP client with no timeout** (`httputil/http.go`): `http.Get`/`http.Post` used the default client with no timeout. If a downstream service hung, goroutines would leak indefinitely. Replaced with a shared `http.Client` with a 10-second timeout.
- **TLS minimum version not set** (`grpcutil/grpc.go`): Added `MinVersion: tls.VersionTLS13` to the `tls.Config` used for gRPC transport credentials.
- **MongoDB URL logged in plaintext** (`initialization/environment.go`, `db/database.go`): `printEnv` now prints `[REDACTED]` instead of the real connection string, and `InitDb` no longer logs the URL at all.
- **Internal model leaked in API response** (`api/register.go`): `PostRegister` previously returned the raw `models.Device` (including internal MongoDB fields). Now returns a purpose-built `DeviceRegisterRes` struct with only the public fields.
- **Hardened Dockerfile**: Uses the `dhi.io/alpine-base:3.23` minimal base image (no Go toolchain in runtime), runs as non-root user via `USER 65534` directive (nobody), and pre-creates a dedicated `/logs` directory with correct ownership.
- **Input validation tightened** (`api/register.go`): Added `alphanum` constraint to `FeatureReq.Name` to reject names containing special characters.


## Idiomatic Go Improvements

- **Context removed from structs** (`api/register.go`, `api/keepalive.go`): Context is now passed as a function parameter, and handlers use the request context (`c.Request.Context()`) instead of a long-lived background context.
- **Request context propagation** (`api/register.go`): DB queries, gRPC calls, and transactions now use the request context, so client disconnections are properly respected.
- **`Start()` return signature simplified** (`initialization/start.go`): Removed `context.Context` from the return tuple since handlers now derive context from each request. Callers (`main.go`, tests) updated accordingly.
- **`Filter` generic simplified** (`utils/slice_utils.go`): Changed `Filter[T any, M bool]` to `Filter[T any]` — the type parameter `M` constrained to `bool` was unnecessary.
- **`else` after `return` removed** (`db/database.go`): `getDbName` now returns early without an else block.
- **`interface{}` replaced with `any`** (`testutils/db_utils.go`, `api/register.go`): Updated all occurrences including the `WithTransaction` callback signature.
- **Short variable declarations** (`api/register.go`, `httputil/http.go`): `var errFields = ""` and `var payloadBody = ...` changed to use `:=`.
- **`strings.Builder` for concatenation** (`utils/validator.go`): Replaced repeated string concatenation with `strings.Builder` for efficiency.
- **Structured logging** (`api/register.go`, `initialization/environment.go`, `initialization/server.go`): Replaced all `Errorf`/`Info` format-string and concatenation log calls with `Errorw`/`Infow`/`Warnw` structured key-value logging for better log aggregation.
- **`context.TODO()` replaced with real contexts**: `db.InitDb` now uses the `ctx` parameter for `client.Ping` instead of `context.TODO()`. `main.go` uses a 5-second timeout context for `client.Disconnect` instead of `context.TODO()`.
- **Library functions return errors instead of panicking**: `db.InitDb` now returns `(*mongo.Client, error)` and `InitEnv` now returns `error`. Previously both called `panic()` on failure, which is non-idiomatic for library code. `Start()` propagates these errors to callers; `main.go` uses `logger.Fatalw` and tests use `Expect(err).ShouldNot(HaveOccurred())`.
- **`utils/` grab-bag split into focused packages**: `utils/grpc.go` moved to `grpcutil/grpc.go`, `utils/http.go` moved to `httputil/http.go`. The `utils/` package retains only generic helpers (`slice_utils.go`, `validator.go`).
- **`testuutils/` renamed to `testutils/`**: Fixed misspelled package name and updated all imports.
- **Struct literal initialization** (`api/register.go`): Replaced field-by-field assignment of `models.Device` with a single composite literal.
- **`if`/`else` instead of `if`/`if`** (`api/register.go`): Merged two mutually exclusive `if isSecure` / `if !isSecure` checks into a single `if`/`else`.
- **Unnecessary intermediate variable removed** (`grpcutil/grpc.go`): `BuildSecurityDialOption` now returns directly from each branch instead of assigning to `securityDialOption` first.
- **`filepath.Join` for path construction** (`grpcutil/grpc.go`, `initialization/logger_config.go`, `initialization/environment.go`): Replaced string concatenation with `filepath.Join` to handle trailing slashes correctly.
- **Bare `var err error` replaced with `:=`** (`testutils/db_utils.go`): Used short variable declaration for the first error assignment.
- **`Filter` simplified** (`utils/slice_utils.go`): Replaced `slices.Collect` with push-iterator with a plain `append` loop — simpler and equally efficient for non-lazy use.
- **Duplicate `getDbName()` call removed** (`db/database.go`): `GetCollections` now calls `getDbName()` once and reuses the `*mongo.Database` handle.
- **Stutter comments replaced with meaningful doc comments**: Updated all `// TypeName struct`-style comments across `customerrors/`, `db/`, `initialization/`, `models/`, and `api/` to describe purpose and behavior.


## Features

- **Device `name` field initialized on insert** (`models/device.go`, `api/register.go`): Added a `Name` string field to `models.Device` (stored in MongoDB as `"name"`, excluded from API responses via `json:"-"`). On registration, `Name` is automatically set to the device's MAC address. The field is not accepted or exposed via the REST API.


## Test Improvements

- **`EnsureCollections` helper** (`testutils/db_utils.go`): New helper creates `profiles` and `devices` collections if they don't exist, preventing test failures on a fresh MongoDB instance without a replica-set `create` event.
- **Keepalive handler route fix** (`integration_tests/register_test.go`): Changed mock HTTP mux pattern from `/keepalive` to `/keepalive/` to match the trailing-slash URL built from env vars.
- **DB verification for `Name` field** (`integration_tests/register_test.go`): The three success test cases (controller, sensor, hybrid) now query MongoDB directly after registration and assert that `device.Name` equals the registered MAC address.


## Chores

- update all dependencies
