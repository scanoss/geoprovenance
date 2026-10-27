# Changelog

This service provides provenance information of projects based on users country of origin.
Two endpoints are available for this.
- Based on declared and human curated location
- Based on calculations of the user data (commits, history, timezone,name)

## [Unreleased]

## [0.7.1] - 2026-10-27

### Fixed
- **The 0.7.0 binaries and Docker image do not start on Debian bookworm.** The DuckDB driver made the build cgo, so the binary links against glibc, and both builds ran on a newer one than the bookworm runtime: the release on `ubuntu-latest` and the image on `golang:1.25`, now Debian 13. They require `GLIBC_2.38` and bookworm has 2.36, so they failed at startup with "version `GLIBC_2.38' not found". The release now builds in a `golang:1.25-bookworm` container and the Dockerfile's build stage uses the same image; the binaries require glibc 2.34 at most

## [0.7.0] - 2026-10-27

### Changed
- **The service now reads the serving knowledge base instead of the mining database.** The mining path is gone; there is no configuration that selects it. Endpoint routes are unchanged
- **Contributor geography is keyed by `purl_id`.** Mining reached it through four tables - `github_contributors` linked a purl name to a contributor login, `vendors` turned that login into an id and a type, and `vendor_locations` held the location data. Serving has `contribution(purl_id → contributor_id)`, `vendor` (login, type, mine), `vendor_location` (the curated country) and `vendor_location_declared` (the declared location and the timezone-inferred country)
- `origin` reads the timezone-inferred country from `vendor_location_declared.timezone_country`, an ISO code (`DE`, `US`) as mining's `timezone_based_country` was. An earlier revision of this branch computed it from the curated country, with full names, because the export did not carry the timezone one yet
- Component resolution no longer goes through `go-component-helper`, which resolves against `all_urls`, `mines` and `versions` - none of which exist in the serving schema. It is replaced by `pkg/resolver`
- The country lookup cache is now guarded by a mutex. It was a bare package-level map that every request goroutine could populate at once
- Binaries are built with `CGO_ENABLED=1`, and the runtime image moved to `debian:bookworm-slim` with `libstdc++6`. The DuckDB driver is a cgo binding, so a pure-Go build can no longer be produced
- A source link that points back at the component itself (11.4 of 13.9 million) is no longer followed. It cost an unindexed full scan of `component`, about 2.5 s of CPU, for a fallback that could only repeat the lookup

### Added
- **DuckDB support**, the intended production backend for the serving knowledge base. Set `DB_DRIVER=duckdb` and `DB_PATH` to the serving file, which is opened read-only
- `pkg/db`, opening the serving database for any of duckdb, sqlite or postgres
- `pkg/resolver`, the serving component resolution
- `pkg/models/test_schema.go`, holding the serving schema the tests build their database from. The fixtures under `tests/` used to carry their own `CREATE TABLE` statements, several in PostgreSQL dialect (`serial4`, `int4`)
- `declared_locations` is populated from `vendor_location_declared`, with mining's filters: a typed vendor from the GitHub mine with a non-empty location
- The `TOO_MANY_CONTRIBUTORS` status, from `too_many_contributors`, on both endpoints. Unlike mining it is reported even when the component has no location data: mining checked for data first, which hid the status from 408 of the 1,150 listed components

### Security
- **Hardened against an unimplemented method killing the server.** The service embedded the papi `GeoProvenanceServer` *interface* rather than the generated `UnimplementedGeoProvenanceServer`. The interface value is nil, so any method the service does not implement would dereference it and panic, taking the whole process down. All seven methods of the current papi interface are implemented, so nothing reaches it today - but the first method papi adds would, silently. scanoss/components shipped exactly that. Pre-existing and unrelated to the serving migration
- Upgraded `google.golang.org/grpc` to v1.83.1, `golang.org/x/net` to v0.57.0 and the OpenTelemetry modules (`otel`, `otel/sdk`, `otlptrace`, `otlptracegrpc`) to v1.45.0 to fix open Dependabot security advisories
- `go-grpc-helper` v0.16.0 → v0.17.0. A failed REST call no longer answers `200` with an empty, unparseable body when the handler set no `x-http-code` trailer: the gateway now falls back to the standard grpc-gateway error response, and gRPC status codes map to their HTTP equivalent, so `Unimplemented` is `501` rather than `500`

### Removed
- `pkg/models/versions.go`, which queried the mining `versions` table and was never instantiated
- The 0.6.0 mining-schema SQLite support: `config/app-config-sqlite.json` and the fixtures under `pkg/models/tests/sqlite` describe the mining export, which the service no longer reads. SQLite still works as a backend, holding the serving schema (`DB_DRIVER=sqlite`)

### Fixed
- Carried over from 0.6.0 to the serving queries: `origin` counts only GitHub vendors (`vendor.mine_id = 5`), and contributor counts are `int` - an `int16` overflowed the percentage past 327 contributors in a country
- Three tests held the pool's only connection while the code under test queried the pool, which deadlocks on an in-memory SQLite database - it is per-connection, so the pool is pinned to one. They did not deadlock before only because the queries went through the very connection they held
- **Scoped npm packages, and other purls the export stores encoded or with qualifiers, resolve again.** `component.purl` keeps scoped npm and jsr names percent-encoded (`pkg:npm/%40angular/core`), qualifiers on generic, conda and cran keys, and a raw `+` in some rpm and deb names. The resolver decoded the purl and rebuilt the key, so 1.96 million components (6%) - plus 1.6 million qualified ones - answered `COMPONENT_NOT_FOUND` however they were spelled. It now tries the caller's spelling, the canonical rendering and the decoded form in one indexed lookup

### Deployment
- **Breaking:** requires a serving knowledge base. Point `DB_PATH` at the `.duckdb` file and set `DB_DRIVER=duckdb`
- **The export now carries what `origin` and `declared_locations` need** (since the 2026-09-23 export, verified on 2026-10-16): `vendor_location_declared` has a timezone country for 2,305,059 vendors - the same count mining had - and a declared location for 1,825,344 rows, and `too_many_contributors` lists 1,150 components. The blocker earlier revisions of this entry recorded is lifted
- The curated coverage itself is complete: serving's `vendor_location` has 258,386 rows (256,731 before the update) against mining's 258,681 vendors with a curated country, and `curated_countries_ids` never held more than one country, so collapsing it to a single `country_id` loses nothing
- `contribution` covers 16,966,369 rows against `github_contributors`' 16,929,664, and `country` matches `countries` at 284 rows
- Verified against the production serving file (72.6 GiB): both endpoints, single and batch, with unknown and invalid purls
- `make build_arm` now cross-compiles with cgo and needs an `aarch64` toolchain (set `CC`); the amd64 build is unaffected

## [0.6.0] - 2026-09-23
### Added
- SQLite support: queries now return the same results on PostgreSQL and on the SQLite export (all `TEXT` columns, `''` instead of `NULL`)
- `config/app-config-sqlite.json` sample config for a read-only SQLite database
- SQLite-schema test fixtures (`pkg/models/tests/sqlite`); model and use case tests run against both schemas
### Fixed
- SQL injection in the purl name `IN (...)` lists; purl names are now bound as query parameters
- Origin counted contributors without a timezone based country as an empty-named country (SQLite)
- Provenance ignored curated countries of contributors with a `NULL` declared location
- Origin counted non-GitHub vendors sharing a username with a GitHub contributor
- Origin percentage overflow for more than 327 contributors in a country, and ignored query errors

## [0.5.0] - 2026-06-22
### Added
- `/health` liveness endpoint (GET) on the REST gateway
### Changed
- Upgraded `scanoss/go-grpc-helper` to `v0.16.0`

## [0.4.0] - 2026-04-16
### Changed
- Replaced `error_message`/`error_code` with `info_message`/`info_code` on per-component info
### Added
- Local `pkg/domain` package with provenance-specific `TooManyContributors` status code

## [0.3.0] - 2026-04-14
### Changed
- Integrated `go-component-helper` for component handling
- Simplified `provenance_service.go` gRPC handler pipeline with a single generic `runPipeline`
- Updated dependencies to the latest versions
### Added
- Added per-component status on batch responses
### Fixed
- Return a bad request error when requests contain no PURLs or only empty PURLs
### Refactored
- Simplified provenance service logic and improved handling of bad requests
### Removed
- Obsolete `response_handler` and `purl` tests and dead helper code
### Chore
- Resolved staticcheck/linter issues across the codebase (package comments, doc style, initialism renames)
- Bumped `golangci-lint` to v2.10.1 and modernized GitHub Actions to `setup-go@v5` using `go.mod`

## [0.2.0] - 2025-09-23
### Changed
- Refactored geo provenance service core architecture
### Added
- New structured error handling system with ServiceError type
- ComponentDTO for component data transfer
- Enhanced status determination logic with HTTP status code mapping
- Enhanced error handling with grpc metadata trailers
- Added gRPC GetCountryContributorsByComponents and REST endpoint POST /v2/geoprovenance/countries/components
- Added gRPC GetCountryContributorsByComponent and REST endpoint GET /v2/geoprovenance/countries/component
- Added gRPC GetOriginByComponents and REST endpoint POST /v2/geoprovenance/origin/components
- Added gRPC GetOriginByComponent and REST endpoint GET /v2/geoprovenance/origin/component
### Fixed
- Improved input validation for PURL requests
- Better error propagation and handling throughout the service
- Enhanced status response messaging with detailed failure information
### Refactored
- Simplified provenance input conversion logic
- Enhanced query summary model with total PURLs tracking


## [0.1.2] - 2025-05-19
### Changed
- Migrated from `go-sqlite3` to `modernc.org/sqlite`
- Updated dependencies to their latest stable versions
- New Makefile command `make unit_test_coverage` to generate and view test coverage reports
### Fixed
- Resolved linter issues
### Test
- Increased overall test coverage
### Chore
- Removed unused code and dead functions


## [0.1.1] - 2025-04-24
## Updated
- Updated SCANOSS PAPI dependency

## [0.1.0] - 2025-04-24
### Added
- Added origin endpoint. This lets the users get countries distribution for a given purl.
## Updated
- Updated Go version to v1.24

## [0.0.6] - 2025-02-05
### Fixed
- Fixed missing curated locations for some contributors

## [0.0.5] - 2025-01-16
### Added
- Added too many contributors warning
- Added a warning message on purls not found/no info/failed to parse


## [0.0.4] - 2024-09-26
### Added
- Added Dockerfile EXPOSE port

[0.7.1]: https://github.com/scanoss/geoprovenance/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/scanoss/geoprovenance/compare/v0.6.0...v0.7.0
[0.4.0]: https://github.com/scanoss/geoprovenance/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/scanoss/geoprovenance/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/scanoss/geoprovenance/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/scanoss/geoprovenance/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/scanoss/geoprovenance/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/scanoss/geoprovenance/compare/v0.0.6...v0.1.0
[0.0.6]: https://github.com/scanoss/geoprovenance/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/scanoss/geoprovenance/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/scanoss/geoprovenance/releases/tag/v0.0.4
