# Deployment

This folder contains a two-step Docker setup optimized for faster local rebuilds.

## Base Image

`Dockerfile.base` installs Go dependencies only. Rebuild it when `go.mod` or `go.sum` changes.

```bash
./deployment/build-base.sh
```

## Application Image

`Dockerfile` uses the prebuilt base image, copies the current source code, injects build metadata, and builds the API binary.

```bash
./deployment/build-api.sh
```

You can override the version:

```bash
APP_VERSION=0.1.0 ./deployment/build-api.sh
```

`GIT_HASH` is resolved from `git rev-parse --short HEAD`. If the repo has no commit yet, it falls back to `dev`.

## Docker Compose

Compose runs the API image, PostgreSQL, and Maildev for local SMTP testing. It does not build `api-base`, so it will not pull `insurance-core-api-base` from Docker Hub.

```bash
docker compose -f deployment/docker-compose.yaml up
```

## Typical Flow

```bash
./deployment/build-base.sh
./deployment/build-api.sh
docker compose -f deployment/docker-compose.yaml up
```

If only source code changes, rerun `./deployment/build-api.sh`. If dependencies change, rerun both build scripts.

## PostgreSQL with pgvector

Docker Compose includes a PostgreSQL 16 service with pgvector for local development.

```bash
docker compose -f deployment/docker-compose.yaml up -d postgres
```

Default credentials:

```env
POSTGRES_DB=insurance_core
POSTGRES_USER=insurance
POSTGRES_PASSWORD=insurance
```

The API container connects to PostgreSQL using the internal Compose hostname `postgres`.

## Maildev SMTP

Compose includes Maildev so you can inspect outbound emails locally.

- SMTP server: `localhost:1025`
- Web UI: `http://localhost:1080`

The API container uses the internal Compose hostname `maildev` for SMTP delivery.
