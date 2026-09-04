# Changelog

This service provides provenance information of projects based on users country of origin.
Two endpoints are available for this.
- Based on declared and human curated location
- Based on calculations of the user data (commits, history, timezone,name)

## [Unreleased]

### Changed
- **The service now reads the serving knowledge base instead of the mining database.** The mining path is gone; there is no configuration that selects it. Endpoint routes are unchanged
- **Contributor geography is keyed by `purl_id`.** Mining reached it through four tables - `github_contributors` linked a purl name to a contributor login, `vendors` turned that login into an id and a type, and `vendor_locations` held the location data. Serving reduces it to `contribution(purl_id → contributor_id)` → `vendor_location(vendor_id → country_id)` → `country`
- **Both endpoints now derive from the curated country.** `curated_locations` did already; `origin` used `vendor_locations.timezone_based_country`, a country inferred from commit timezones, which serving does not carry
- **`locations[].name` changes from an ISO code to a country name.** Mining's `timezone_based_country` held `DE`, `US`, `PL`; serving joins `country.name`, so the same component reports `Germany`, `United States`, `Poland`
- Component resolution no longer goes through `go-component-helper`, which resolves against `all_urls`, `mines` and `versions` - none of which exist in the serving schema. It is replaced by `pkg/resolver`
- The country lookup cache is now guarded by a mutex. It was a bare package-level map that every request goroutine could populate at once
- Binaries are built with `CGO_ENABLED=1`, and the runtime image moved to `debian:bookworm-slim` with `libstdc++6`. The DuckDB driver is a cgo binding, so a pure-Go build can no longer be produced

### Added
- **DuckDB support**, the intended production backend for the serving knowledge base. Set `DB_DRIVER=duckdb` and `DB_PATH` to the serving file, which is opened read-only
- `pkg/db`, opening the serving database for any of duckdb, sqlite or postgres
- `pkg/resolver`, the serving component resolution
- `pkg/models/test_schema.go`, holding the serving schema the tests build their database from. The fixtures under `tests/` used to carry their own `CREATE TABLE` statements, several in PostgreSQL dialect (`serial4`, `int4`)

### Security
- **Hardened against an unimplemented method killing the server.** The service embedded the papi `GeoProvenanceServer` *interface* rather than the generated `UnimplementedGeoProvenanceServer`. The interface value is nil, so any method the service does not implement would dereference it and panic, taking the whole process down. All seven methods of the current papi interface are implemented, so nothing reaches it today - but the first method papi adds would, silently. scanoss/components shipped exactly that. Pre-existing and unrelated to the serving migration

### Removed
- `declared_locations` is always empty. It needed `vendor_locations.declared_location` and `vendors.type`, neither of which serving carries; 1,803,028 vendors have a declared location in mining
- The `TOO_MANY_CONTRIBUTORS` status is gone. It came from the mining `too_many_contributors` table (1,150 purls), which has no serving equivalent
- `pkg/models/versions.go`, which queried the mining `versions` table and was never instantiated

### Fixed
- Three tests held the pool's only connection while the code under test queried the pool, which deadlocks on an in-memory SQLite database - it is per-connection, so the pool is pinned to one. They did not deadlock before only because the queries went through the very connection they held

### Deployment
- **Breaking:** requires a serving knowledge base. Point `DB_PATH` at the `.duckdb` file and set `DB_DRIVER=duckdb`
- **Blocker: `origin` covers far fewer components.** Mining recorded a timezone-inferred country for 2,305,059 vendors and a curated one for 256,410. Where both exist they agree on 119,906 of 120,988 (99.1%), so the answers are consistent - the curated set is simply much smaller. On 30 randomly sampled components, 38 of 89 located contributors (42%) had a timezone country and 7 (7%) a curated one, and **14 of the 21 components that resolve to an origin today would resolve to none**. The composition changes too: `pkg:github/camunda/zeebe` reports `DE 29, RU 8, IN 7, US 6, CN 5` from mining and `Germany 42.55%, Ukraine 8.51%, United States 8.51%` from serving. **Do not merge until the export carries `timezone_based_country`, or until the narrower coverage is accepted**; this is an ETL gap, not a code one
- The curated coverage itself is complete: serving's `vendor_location` has 256,731 rows against mining's 256,410 vendors with a curated country, and `curated_countries_ids` never held more than one country, so collapsing it to a single `country_id` loses nothing
- `contribution` covers 16,907,601 rows against `github_contributors`' 16,929,664 (99.9%), and `country` matches `countries` at 284 rows
- Verified against the production serving file (55.7 GiB): both endpoints, single and batch, with unknown and invalid purls
- `make build_arm` now cross-compiles with cgo and needs an `aarch64` toolchain (set `CC`); the amd64 build is unaffected

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

[0.4.0]: https://github.com/scanoss/geoprovenance/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/scanoss/geoprovenance/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/scanoss/geoprovenance/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/scanoss/geoprovenance/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/scanoss/geoprovenance/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/scanoss/geoprovenance/compare/v0.0.6...v0.1.0
[0.0.6]: https://github.com/scanoss/geoprovenance/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/scanoss/geoprovenance/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/scanoss/geoprovenance/releases/tag/v0.0.4
