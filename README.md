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

### Quick Start with Docker Compose (Recommended)

The easiest way to run the complete application stack:

```bash
make docker-run
```

This automatically:
- Builds and starts both the Go application and PostgreSQL database
- Initializes the database schema from `schema.sql`
- Sets up networking between services
- Falls back to `docker-compose` if `docker compose` isn't available

To stop the services:
```bash
make docker-down
```

### Development Setup

For local development with live reload, you'll need to use your own PostgreSQL instance since the Makefile doesn't provide a database-only option:

```bash
# Set up your own PostgreSQL with the schema
psql -h $LANGLITE_DB_HOST -p $LANGLITE_DB_PORT -U $LANGLITE_DB_USERNAME -d $LANGLITE_DB_DATABASE -f schema.sql

# Run Go app with live reload
make watch         # Installs 'air' if needed
```

### Alternative Setups

**Run the Go application locally** (requires external PostgreSQL):
```bash
make run
```

**Use your own PostgreSQL instance:**
```bash
# First, set up your database with the required schema
psql -h $LANGLITE_DB_HOST -p $LANGLITE_DB_PORT -U $LANGLITE_DB_USERNAME -d $LANGLITE_DB_DATABASE -f schema.sql

# Then run the application
make run
```

## For Production

When you want to build for production deployment:

```bash
docker build --target prod -t myapp:latest .
```

This creates an optimized production image without development tools like air. The multi-stage build automatically handles dependencies - Docker will build the `build` stage first to compile the Go binary, then create the minimal `prod` stage with just the compiled application.

## Project Structure

```
├── cmd/                 # Application entrypoints
├── internal/            # Private application code
├── migrations/          # Database migrations (Goose format)
├── scripts/             # Development and testing scripts
│   ├── test_helper.sh   # Rate limiting test utilities
│   └── *.sh            # Other test scripts
├── sql/                 # SQL setup files
│   ├── schema.sql       # Main database schema
│   └── setup_test_data.sql # Test data for development
├── docker-compose.yml   # Development environment
└── Makefile            # Build and development commands
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

## Rate Limiting Development

Test rate limiting functionality:

```bash
make test-rate-limit      # Run basic rate limit tests
make reset-rate-limit     # Reset rate limits for testing
make rate-limit-status    # Check current rate limit usage
```

Or use the helper script directly:

```bash
./scripts/test_helper.sh help    # Show all available commands
./scripts/test_helper.sh status  # Check rate limit status
./scripts/test_helper.sh reset   # Reset rate limits
./scripts/test_helper.sh test    # Run basic tests
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
