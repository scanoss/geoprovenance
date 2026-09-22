# SCANOSS Platform 2.0 Provenance
Welcome to the Provenance server for SCANOSS Platform 2.0

**Warning** Work In Progress **Warning**

## Repository Structure
This repository is made up of the following components:
* ?

## Configuration

Environmental variables are fed in this order:

dot-env --> env.json -->  Actual Environment Variable

These are the supported configuration arguments:

```
APP_NAME="SCANOSS Provenance Server"
APP_PORT=50051
APP_MODE=dev
APP_DEBUG=false

DB_DRIVER=postgres
DB_HOST=localhost
DB_USER=scanoss
DB_PASSWD=
DB_SCHEMA=scanoss
DB_SSL_MODE=disable
DB_DSN=
```

### Database
The service supports both PostgreSQL and SQLite databases, selected with `DB_DRIVER`:

* `postgres` - connects using `DB_HOST`, `DB_USER`, `DB_PASSWD`, `DB_SCHEMA` and `DB_SSL_MODE` (or a full `DB_DSN`).
* `sqlite` - opens the database file given in `DB_DSN`. Use a `file:` URI with `mode=ro` to open it read-only, e.g.:
  `DB_DSN=file:/path/to/provenance.sqlite?mode=ro` (without the `file:` prefix, query parameters are ignored).

The SQLite export stores every column as `TEXT` and uses empty strings instead of `NULL`
(PostgreSQL mixes both, e.g. `''` for missing declared locations and `NULL` for missing timezone countries).
Queries are written to behave the same on both engines, and the unit tests run against both schemas
(see [pkg/models/tests/sqlite](pkg/models/tests/sqlite)).


## Docker Environment

The provenance server can be deployed as a Docker container.

Adjust configurations by updating an .env file in the root of this repository.


### How to build

You can build your own image of the SCANOSS Provenance Server with the ```docker build``` command as follows.

```bash
make ghcr_build
```


### How to run

Run the SCANOSS Provenance Server Docker image by specifying the environmental file to be used with the ```--env-file``` argument. 

You may also need to expose the ```APP_PORT``` on a given ```interface:port``` with the ```-p``` argument.

## Development

To run locally on your desktop, please use the following command:

```shell
go run cmd/server/main.go -json-config config/app-config-dev.json -debug
```

To run against a local SQLite database instead:
```shell
DB_DSN="file:/path/to/provenance.sqlite?mode=ro" go run cmd/server/main.go -json-config config/app-config-sqlite.json -debug
```

After changing a Provenance version, please run the following command:
```shell
go mod tidy -compat=1.19
```
https://mholt.github.io/json-to-go/
