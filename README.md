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

```bash
docker run -it -v "$(pwd)":"$(pwd)" -p 50051:50051 ghcr.io/scanoss/scanoss-geoprovenance -json-config $(pwd)/config/app-config-docker-local-dev.json -debug
```

## Development

To run locally on your desktop, please use the following command:

```shell
go run cmd/server/main.go -json-config config/app-config-dev.json -debug
```

After changing a Provenance version, please run the following command:
```shell
go mod tidy -compat=1.19
```
https://mholt.github.io/json-to-go/
