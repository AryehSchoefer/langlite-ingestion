# langlite-ingestion

The production-ready Go ingestion API that serves as the backend for the LangLite TypeScript SDK. Handles observability data collection including LLM tracing, span tracking, event logging, and evaluation scoring with robust validation, batch processing, and high-throughput performance for TypeScript applications.

## Environment Variables

The following environment variables are required for the service to run:

### Database Configuration

- `LANGLITE_DB_HOST` - PostgreSQL host (e.g., `localhost`)
- `LANGLITE_DB_PORT` - PostgreSQL port (e.g., `5432`)
- `LANGLITE_DB_DATABASE` - Database name (e.g., `langlite`)
- `LANGLITE_DB_USERNAME` - Database username
- `LANGLITE_DB_PASSWORD` - Database password
- `LANGLITE_DB_SCHEMA` - Database schema (e.g., `langlite`)

### Server Configuration

- `PORT` - Server port (default: 8080)

### Optional Configuration

- `LANGLITE_CORS_ORIGINS` - Comma-separated list of allowed CORS origins (defaults to localhost and app.langlite.com)

## Getting Started

### Using Docker Compose (Recommended)

The easiest way to run the application is using Docker Compose, which automatically sets up both the application and PostgreSQL database with the required schema:

```bash
docker-compose up
```

This will:
- Start a PostgreSQL database with the schema automatically initialized
- Build and run the Go application
- Set up networking between the services

### Manual Setup

If you prefer to run the application manually, ensure your PostgreSQL database is set up with the required schema:

```bash
psql -h $LANGLITE_DB_HOST -p $LANGLITE_DB_PORT -U $LANGLITE_DB_USERNAME -d $LANGLITE_DB_DATABASE -f schema.sql
```

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB Container

```bash
make docker-down
```

DB Integrations Test:

```bash
make itest
```

Live reload the application:

```bash
make watch
```

Run the test suite:

```bash
make test
```

Clean up binary from the last build:

```bash
make clean
```

## API Endpoints

### Health Check

- `GET /health` - Returns database health status

### Observability Endpoints

- `POST /v1/trace` - Create a new trace
- `POST /v1/generation` - Create a new generation
- `POST /v1/span` - Create a new span
- `POST /v1/span/{id}` - Update an existing span
- `POST /v1/trace/batch` - Create multiple traces in batch
- `POST /v1/event` - Create a new event
- `POST /v1/score` - Create a new score

All endpoints return JSON responses with appropriate HTTP status codes and detailed error messages for validation failures.
