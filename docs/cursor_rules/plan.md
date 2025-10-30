<!-- 860da2ac-1fb5-4212-89d8-856e4d037ef5 0411f9b1-21e6-49da-8f7d-f829117603f6 -->
# Go Backend Development Learning Path

## Phase 1: Foundation Setup

**Goal**: Get Go installed and create the initial "Hello World" server

- Install Go on Windows and verify installation
- Initialize a Go module in the project directory
- Create a simple HTTP server in `main.go` that responds "Hello World" on port 8080
- **Learning Focus**: Go modules, `net/http` package basics, and running Go programs

## Phase 2: Basic Project Structure

**Goal**: Introduce proper project layout and handle your first POST endpoint

- Restructure to use `cmd/server/main.go` as the entry point
- Create `internal/handlers/` package with a health check handler
- Add a POST endpoint `/api/submit` that accepts JSON with name, email, address
- Parse and log the incoming request body
- **Learning Focus**: Why `cmd/` and `internal/` exist, package organization, JSON handling with `encoding/json`, HTTP methods

## Phase 3: Data Models and Validation

**Goal**: Add structure and validation to incoming data

- Create `internal/models/user.go` with a User struct and validation tags
- Add input validation logic (required fields, email format, etc.)
- Return proper HTTP status codes and error responses (400 for bad input, 200 for success)
- **Learning Focus**: Struct tags, validation patterns, idiomatic error handling, HTTP status codes

## Phase 4: Configuration and Environment Setup

**Goal**: Externalize configuration and prepare for AWS integration

- Create `internal/config/config.go` to load environment variables
- Add `.env` file support (using a library like `godotenv` or manual parsing)
- Store AWS region, DynamoDB table name, and server port in config
- **Learning Focus**: Configuration management, environment variables, the 12-factor app principle

## Phase 5: AWS Credentials and DynamoDB Setup

**Goal**: Set up AWS credentials and create DynamoDB table

- Guide through creating AWS account/IAM user (if needed)
- Set up AWS credentials file or environment variables
- Create DynamoDB table (either via AWS Console or AWS CLI guidance)
- Define table schema: partition key and attributes
- **Learning Focus**: AWS IAM basics, DynamoDB concepts (partition keys, attributes), AWS SDK setup

## Phase 6: Database Layer with DynamoDB

**Goal**: Implement data persistence with proper abstraction

- Add AWS SDK dependencies (`github.com/aws/aws-sdk-go-v2`)
- Create `internal/repository/user_repository.go` with an interface
- Implement DynamoDB client initialization
- Write `SaveUser()` method to persist data to DynamoDB
- **Learning Focus**: Interfaces for abstraction, dependency injection, AWS SDK v2, DynamoDB PutItem operations

## Phase 7: Service Layer and Business Logic

**Goal**: Separate concerns with a service layer

- Create `internal/service/user_service.go`
- Move validation and business logic from handler to service
- Wire service → repository → database
- Update handlers to use the service layer
- **Learning Focus**: Layered architecture, separation of concerns, dependency injection patterns

## Phase 8: Structured Logging

**Goal**: Add professional logging throughout the application

- Introduce structured logging (using `slog` from Go 1.21+ or `zerolog`/`zap`)
- Add context-aware logging in handlers, services, and repository
- Log requests, errors, and important operations
- **Learning Focus**: Structured vs unstructured logging, log levels, production logging best practices

## Phase 9: Error Handling and Recovery

**Goal**: Implement robust error handling

- Create custom error types in `internal/errors/`
- Add panic recovery middleware
- Implement proper error propagation through layers
- Return user-friendly error messages while logging details
- **Learning Focus**: Error wrapping with `fmt.Errorf` and `%w`, middleware pattern, panic recovery

## Phase 10: Testing and Production Readiness

**Goal**: Add tests and finalize the project

- Write unit tests for service layer (with mocked repository)
- Write integration tests for handlers
- Add a `Makefile` or build scripts
- Create a README with setup and run instructions
- **Learning Focus**: Table-driven tests, mocking with interfaces, `httptest` package, Go testing conventions

## Success Criteria

- You can POST user data via curl and see it stored in DynamoDB
- The codebase is modular, testable, and follows Go idioms
- You understand why each layer exists and how to extend it
- The application handles errors gracefully with proper logging

### To-dos

- [ ] Install Go and initialize module with hello world server
- [ ] Restructure to cmd/internal layout and add POST endpoint
- [ ] Create User model with validation logic
- [ ] Add configuration management with environment variables
- [ ] Guide AWS credentials setup and DynamoDB table creation
- [ ] Implement DynamoDB repository with interface
- [ ] Create service layer with business logic
- [ ] Add structured logging throughout the application
- [ ] Implement robust error handling and recovery
- [ ] Write tests, Makefile, and documentation