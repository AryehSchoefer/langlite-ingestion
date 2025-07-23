# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Rate limiting test shortcuts
test-rate-limit:
	@echo "Running rate limit tests..."
	@./scripts/test_helper.sh test

reset-rate-limit:
	@echo "Resetting rate limits..."
	@./scripts/test_helper.sh reset

rate-limit-status:
	@echo "Checking rate limit status..."
	@./scripts/test_helper.sh status

# Async processing test shortcuts
test-async:
	@echo "Testing async processing pipeline..."
	@./scripts/test_async_processing.sh

queue-status:
	@echo "Checking queue status..."
	@curl -s -X GET http://localhost:8080/queue-status -H "Authorization: Bearer test-key-123" | python3 -m json.tool

worker-status:
	@echo "Checking worker status..."
	@curl -s -X GET http://localhost:8080/worker-status -H "Authorization: Bearer test-key-123" | python3 -m json.tool

processing-status:
	@echo "Checking processing status..."
	@curl -s -X GET http://localhost:8080/processing-status -H "Authorization: Bearer test-key-123" | python3 -m json.tool

# Metrics and monitoring shortcuts
test-metrics:
	@echo "Testing metrics and monitoring..."
	@./scripts/test_metrics.sh

metrics:
	@echo "Opening metrics endpoint..."
	@curl -s http://localhost:8080/metrics | head -20

prometheus:
	@echo "Prometheus available at: http://localhost:9090"
	@echo "Targets: http://localhost:9090/targets"

grafana:
	@echo "Grafana available at: http://localhost:3000"
	@echo "Login: admin/admin"

.PHONY: all build run test clean watch docker-run docker-down itest test-rate-limit reset-rate-limit rate-limit-status test-async queue-status worker-status processing-status test-metrics metrics prometheus grafana
