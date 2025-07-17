# Go HTTP Service Best Practices Guide

This guide outlines the essential best practices for building production-ready Go HTTP services.

## 1. Project Structure

### Standard Layout
- `/cmd/` - Application entry points
- `/internal/` - Private application code
- `/pkg/` - Public library code (if any)
- `go.mod` and `go.sum` - Go module files
- `README.md` - Project documentation
- `Makefile` - Build automation

### Internal Structure
- `/internal/server/` - HTTP server logic
- `/internal/database/` - Data persistence layer
- `/internal/models/` - Data models (can be combined with database)

## 2. HTTP Server Setup

### Server Configuration
- Use proper timeouts (ReadTimeout, WriteTimeout, IdleTimeout)
- Configure graceful shutdown with signal handling
- Use environment variables for configuration
- Set appropriate server address and port

### Router Selection
- Use a robust router like `chi` or `gorilla/mux`
- Implement middleware for cross-cutting concerns
- Group related routes logically

## 3. Error Handling

### Structured Error Responses
- Use consistent error response format
- Include error codes, messages, and context
- Validate input data with proper error messages
- Handle different error types appropriately

### HTTP Status Codes
- Use correct HTTP status codes (200, 201, 400, 404, 500, etc.)
- Return 400 for validation errors
- Return 404 for not found resources
- Return 500 for internal server errors

## 4. Request/Response Handling

### JSON Handling
- Use proper JSON encoding/decoding
- Validate request payloads
- Set appropriate Content-Type headers
- Handle malformed JSON gracefully

### Request Validation
- Validate all input data
- Use structured validation with detailed error messages
- Check required fields and data types
- Implement business logic validation

## 5. Database Integration

### Connection Management
- Use connection pooling
- Implement proper connection configuration
- Handle database errors gracefully
- Use health checks for database connectivity

### Data Layer Abstraction
- Define interfaces for database operations
- Implement repository pattern
- Use proper SQL error handling
- Implement database migrations (if applicable)

## 6. Middleware and CORS

### Essential Middleware
- Request logging
- CORS handling
- Authentication/authorization (if required)
- Request ID tracking
- Rate limiting (if applicable)

### CORS Configuration
- Configure allowed origins, methods, and headers
- Set appropriate max age
- Handle preflight requests

## 7. Security

### Input Validation
- Validate all user inputs
- Sanitize data appropriately
- Use parameterized queries for database operations
- Implement rate limiting where necessary

### Headers and Security
- Set security headers
- Use HTTPS in production
- Handle sensitive data properly
- Implement proper authentication

## 8. Testing

### Test Coverage
- Unit tests for business logic
- Integration tests for database operations
- HTTP handler tests
- Health check tests

### Test Structure
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error conditions
- Use testcontainers for integration tests

## 9. Observability

### Logging
- Use structured logging
- Log at appropriate levels
- Include request context in logs
- Avoid logging sensitive information

### Health Checks
- Implement health endpoints
- Check database connectivity
- Include dependency status
- Return appropriate status codes

### Metrics
- Track key performance indicators
- Monitor response times
- Track error rates
- Monitor resource usage

## 10. Configuration Management

### Environment Variables
- Use environment variables for configuration
- Provide sensible defaults
- Document all configuration options
- Use configuration validation

### Secret Management
- Never commit secrets to version control
- Use proper secret management tools
- Rotate secrets regularly
- Separate configuration from secrets

## 11. Build and Deployment

### Build Process
- Use Go modules for dependency management
- Implement reproducible builds
- Use multi-stage Docker builds
- Optimize binary size

### Deployment
- Use container orchestration
- Implement blue-green deployments
- Monitor application health
- Use proper resource limits

## 12. Documentation

### Code Documentation
- Document public APIs
- Use clear variable and function names
- Add comments for complex logic
- Maintain up-to-date README

### API Documentation
- Document all endpoints
- Include request/response examples
- Document error conditions
- Provide client SDKs if applicable

## Compliance Checklist

Use this checklist to verify adherence to the guide:

- [ ] Standard project structure with cmd/ and internal/
- [ ] Proper HTTP server configuration with timeouts
- [ ] Graceful shutdown implementation
- [ ] Structured error handling with consistent format
- [ ] Input validation with detailed error messages
- [ ] Proper HTTP status code usage
- [ ] Database connection pooling and health checks
- [ ] CORS configuration
- [ ] Request logging middleware
- [ ] Comprehensive test coverage
- [ ] Health check endpoints
- [ ] Environment-based configuration
- [ ] Security best practices
- [ ] Proper documentation
- [ ] Build automation with Makefile