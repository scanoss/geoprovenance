# Changelog

This service provides provenance information of projects based on users country of origin.
Two endpoints are available for this.
- Based on declared and human curated location
- Based on calculations of the user data (commits, history, timezone,name)


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

[0.1.2]: https://github.com/scanoss/geoprovenance/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/scanoss/geoprovenance/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/scanoss/geoprovenance/compare/v0.0.6...v0.1.0
[0.0.6]: https://github.com/scanoss/geoprovenance/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/scanoss/geoprovenance/compare/v0.0.4...v0.0.5
[0.0.4]: https://github.com/scanoss/geoprovenance/releases/tag/v0.0.4
