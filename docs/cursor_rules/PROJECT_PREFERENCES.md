# Project Preferences and Standards

<!-- This file contains project-specific preferences that should be maintained across all development sessions -->

## Project Context

### Current Project: Go Backend Development Learning Path
- **Primary Language**: Go (Golang)
- **Database**: AWS DynamoDB
- **Cloud Platform**: AWS
- **Architecture**: RESTful API with layered architecture
- **Learning Focus**: Building production-ready backend services

## Development Environment Preferences

### IDE and Tools
- **Primary IDE**: VS Code / PyCharm (as per user preference)
- **Go Version**: Latest stable version
- **Package Manager**: Go Modules
- **Code Formatting**: gofmt + goimports
- **Linting**: golangci-lint
- **Testing Framework**: Built-in Go testing package

### Project Structure Standards
```
project-root/
├── cmd/                    # Application entry points
│   └── server/
│       └── main.go
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── handlers/         # HTTP handlers
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic layer
│   └── errors/           # Custom error types
├── pkg/                  # Public library code (if any)
├── test/                 # Test files and test data
├── scripts/              # Build and deployment scripts
├── docs/                 # Documentation
├── .env.example         # Environment variables template
├── .gitignore           # Git ignore rules
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── Makefile             # Build automation
└── README.md            # Project documentation
```

## Code Style Preferences

### Go-Specific Standards
- **Line Length**: 100 characters maximum
- **Indentation**: Tabs (Go standard)
- **Naming Conventions**:
  - Exported functions: PascalCase
  - Unexported functions: camelCase
  - Constants: PascalCase or ALL_CAPS
  - Variables: camelCase
- **Error Handling**: Always handle errors explicitly
- **Context Usage**: Use context.Context for request-scoped operations

### Documentation Standards
- **Package Comments**: Every package should have a doc comment
- **Function Comments**: All exported functions must have doc comments
- **Example Usage**: Provide examples for complex functions
- **README Structure**: Include setup, usage, and API documentation

## Testing Preferences

### Test Organization
- **Test Files**: Co-locate with source files (e.g., `user_test.go` with `user.go`)
- **Test Functions**: Use descriptive names like `TestUserValidation_EmptyEmail_ReturnsError`
- **Test Data**: Use table-driven tests for multiple scenarios
- **Mocking**: Use interfaces for dependency injection and mocking

### Coverage Requirements
- **Minimum Coverage**: 80% for business logic
- **Critical Paths**: 100% coverage for authentication, data validation, and core business logic
- **Integration Tests**: Cover all API endpoints and database operations

## Configuration Management

### Environment Variables
- **Development**: Use `.env` files with `godotenv` library
- **Production**: Use environment variables or AWS Parameter Store
- **Required Variables**:
  ```
  AWS_REGION=us-east-1
  DYNAMODB_TABLE_NAME=users
  SERVER_PORT=8080
  LOG_LEVEL=info
  ```

### Configuration Structure
- **Centralized Config**: Single config package for all configuration
- **Validation**: Validate all configuration on startup
- **Defaults**: Provide sensible defaults for all configuration options
- **Type Safety**: Use typed configuration structs

## Logging Standards

### Logging Library
- **Primary**: Go's built-in `slog` (Go 1.21+) or `zerolog`
- **Format**: Structured JSON logging for production
- **Levels**: DEBUG, INFO, WARN, ERROR, FATAL

### Logging Best Practices
- **Context**: Include request ID and user context where available
- **Performance**: Avoid logging in hot paths
- **Sensitivity**: Never log passwords, tokens, or personal data
- **Correlation**: Use correlation IDs for request tracing

## Error Handling Standards

### Error Types
- **Custom Errors**: Create specific error types for different failure modes
- **Error Wrapping**: Use `fmt.Errorf` with `%w` for error chaining
- **HTTP Errors**: Map business errors to appropriate HTTP status codes
- **User Messages**: Provide user-friendly error messages

### Error Response Format
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format",
    "details": {
      "field": "email",
      "value": "invalid-email"
    }
  }
}
```

## API Design Preferences

### RESTful Standards
- **URL Structure**: Use nouns for resources, verbs for actions
- **HTTP Methods**: GET, POST, PUT, DELETE, PATCH appropriately
- **Status Codes**: Use standard HTTP status codes
- **Response Format**: Consistent JSON structure

### Request/Response Standards
```json
// Success Response
{
  "data": { ... },
  "meta": {
    "timestamp": "2024-01-01T00:00:00Z",
    "request_id": "uuid"
  }
}

// Error Response
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

## Database Preferences

### DynamoDB Standards
- **Table Design**: Single table design for related data
- **Access Patterns**: Design tables around query patterns
- **Indexing**: Use GSIs for alternative access patterns
- **Naming**: Use consistent naming conventions

### Data Modeling
- **Partition Keys**: Use meaningful, evenly distributed keys
- **Sort Keys**: Use for range queries and sorting
- **Attributes**: Use consistent attribute naming
- **Types**: Use appropriate DynamoDB attribute types

## Security Standards

### Authentication & Authorization
- **JWT Tokens**: Use for stateless authentication
- **Role-Based Access**: Implement RBAC for authorization
- **Input Validation**: Validate all inputs
- **Rate Limiting**: Implement API rate limiting

### Data Protection
- **Encryption**: Encrypt sensitive data at rest
- **HTTPS**: Use TLS for all communications
- **Secrets**: Store secrets in AWS Secrets Manager
- **Audit Logging**: Log all security-relevant events

## Performance Standards

### Response Time Targets
- **API Responses**: < 200ms for simple operations
- **Database Queries**: < 100ms for single item operations
- **Complex Operations**: < 1s for multi-step operations

### Optimization Guidelines
- **Caching**: Implement caching for frequently accessed data
- **Connection Pooling**: Use connection pooling for external services
- **Async Processing**: Use goroutines for I/O operations
- **Monitoring**: Implement performance monitoring

## Deployment Preferences

### Container Strategy
- **Docker**: Use multi-stage builds for optimized images
- **Base Image**: Use official Go alpine images
- **Size**: Keep images under 100MB
- **Security**: Scan images for vulnerabilities

### AWS Deployment
- **Compute**: Use AWS ECS or Lambda for serverless
- **Database**: DynamoDB with proper backup strategies
- **Monitoring**: CloudWatch for logging and metrics
- **Security**: Use IAM roles and VPC for network isolation

## Documentation Standards

### API Documentation
- **OpenAPI/Swagger**: Generate API documentation
- **Examples**: Provide request/response examples
- **Error Codes**: Document all possible error responses
- **Authentication**: Document authentication requirements

### Code Documentation
- **README**: Comprehensive setup and usage instructions
- **Architecture**: Document system architecture and design decisions
- **Deployment**: Document deployment procedures
- **Troubleshooting**: Common issues and solutions

---

**Project Started**: 2024
**Last Updated**: [Auto-updated on each session]
**Maintainer**: Development Team

> **Note**: These preferences should be reviewed and updated as the project evolves and requirements change.
