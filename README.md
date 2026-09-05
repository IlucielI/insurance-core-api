# Insurance Core API

Insurance Policy Application backend for the technical test.

## Scope

This API will focus on the core assessment requirements:

- Browse insurance products
- Simulate premium from age, sum assured, and payment term
- Submit policy applications
- Review applications through the underwriting lifecycle
- Provide one practical AI-powered feature

## Architecture

The project uses a simple layered structure inside `internal`:

- `controllers`: HTTP request/response handlers
- `services`: business logic and use cases
- `repositories`: data access contracts and implementations
- `routes`: HTTP route registration and middleware setup
- `adapters`: external-service adapters such as database, object storage, or LLM clients
- `config`: environment-based application configuration

## Prerequisites

For local Go development:

- Go 1.23 or newer

For Docker development:

- Docker
- Docker Compose

## Environment Setup

Create a local environment file before running the app:

```bash
cp .env.example .env
```

Default values:

```env
APP_NAME=insurance-core-api
APP_ENV=development
APP_VERSION=0.1.0
GIT_HASH=dev
HTTP_PORT=8080
DATABASE_URL=postgres://insurance:insurance@localhost:5432/insurance_core?sslmode=disable
LLM_BASE_URL=http://localhost:20128/v1
LLM_API_KEY=
LLM_COMPLETION_MODEL=
LLM_EMBEDDING_MODEL=
```

`APP_VERSION` and `GIT_HASH` can be injected during Docker build. Remove them from `.env` if you want the binary build metadata to be used instead of runtime overrides.

## Database

For Docker development, PostgreSQL with pgvector is included in `deployment/docker-compose.yaml`.

When running the API locally with Go, start PostgreSQL from Compose first:

```bash
docker compose -f deployment/docker-compose.yaml up -d postgres
```

The default local connection string is:

```env
DATABASE_URL=postgres://insurance:insurance@localhost:5432/insurance_core?sslmode=disable
LLM_BASE_URL=http://localhost:20128/v1
LLM_API_KEY=
LLM_COMPLETION_MODEL=
LLM_EMBEDDING_MODEL=
```

## LLM Adapter

The LLM adapter is OpenAI-compatible and intended for OmniRoute. It supports:

- Chat completions through `/v1/chat/completions`
- Embeddings through `/v1/embeddings`

Configure these values when an AI feature starts using the adapter:

```env
LLM_BASE_URL=http://localhost:20128/v1
LLM_API_KEY=
LLM_COMPLETION_MODEL=your-completion-model
LLM_EMBEDDING_MODEL=your-embedding-model
```

## Run Locally with Go

Install dependencies and start the API:

```bash
go mod tidy
go run ./cmd/api
```

The API starts on `http://localhost:8080` by default.

## Run with Docker

Build the dependency base image once:

```bash
./deployment/build-base.sh
```

Build the application image:

```bash
./deployment/build-api.sh
```

Run the container:

```bash
docker compose -f deployment/docker-compose.yaml up
```

If only source code changes, rerun `./deployment/build-api.sh`. If `go.mod` or `go.sum` changes, rerun both build scripts.

## Health Check

```bash
curl http://localhost:8080/health
```

Example response:

```json
{
  "version": "0.1.0",
  "uptime": "10.5s",
  "git_hash": "dev"
}
```

## Product Endpoints

List all products:

```bash
curl http://localhost:8080/api/v1/products
```

List featured products for the landing page:

```bash
curl "http://localhost:8080/api/v1/products?featured=true&limit=3"
```

Get product detail by slug:

```bash
curl http://localhost:8080/api/v1/products/secure-life-plus
```

## Premium Quote Endpoint

Create an indicative premium quote for a product:

```bash
curl -X POST http://localhost:8080/api/v1/products/secure-life-plus/quotes \
  -H "Content-Type: application/json" \
  -d '{
    "age": 32,
    "gender": "male",
    "sum_assured": 500000000,
    "payment_term": 10,
    "payment_frequency": "monthly",
    "smoker": "no",
    "occupation_class": "standard",
    "health_risk": "low"
  }'
```

The quote is rule-based and returns premium breakdown factors so the calculation remains explainable.
