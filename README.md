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

# The serving knowledge base. DuckDB is the production backend.
DB_DRIVER=duckdb                    # duckdb, sqlite or postgres
DB_PATH=/var/lib/scanoss/serving.duckdb   # required by duckdb and sqlite; opened read-only

# To serve the same schema from a PostgreSQL instance instead:
# DB_DRIVER=postgres
# DB_DSN=postgres://user:pass@host/db?sslmode=disable
```

### Database
The service reads the SCANOSS serving knowledge base, selected with `DB_DRIVER`:

* `duckdb` (production) - opens the serving `.duckdb` file given in `DB_PATH`, read-only.
* `sqlite` - opens a SQLite file holding the serving schema, given in `DB_PATH`. The unit tests use it.
* `postgres` - connects using `DB_DSN`, against a database holding the serving schema.

The mining schema (`all_urls`, `vendors`, `vendor_locations`, `github_contributors`, ...) is no longer read.


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

To run against a different serving file, override its path:
```shell
DB_PATH=/path/to/serving.duckdb go run cmd/server/main.go -json-config config/app-config-dev.json -debug
```

After changing a Provenance version, please run the following command:
```shell
go mod tidy -compat=1.19
```
https://mholt.github.io/json-to-go/
