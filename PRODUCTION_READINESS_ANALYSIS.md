# Production Readiness Analysis & Improvement Plan

**Generated:** 2024  
**System:** Rental Management Application  
**Language:** Go 1.22.7  
**Database:** PostgreSQL (Neon)

---

## Executive Summary

This document provides a comprehensive analysis of the current codebase, identifying redundancies, gaps, conflicting logic, and security concerns. It outlines a roadmap to transform the system into a production-ready, robust, and maintainable application.

### Current State Assessment

**Strengths:**
- ✅ Clean layered architecture (handlers → services → repositories)
- ✅ Separation of concerns with dedicated services
- ✅ Database transaction support for critical operations
- ✅ Partial payment support with transaction verification workflow
- ✅ Notification system with Telegram integration

**Critical Issues:**
- 🔴 **Security vulnerabilities** (password hashing, session management)
- 🔴 **No structured logging** (using fmt.Println)
- 🔴 **No error recovery** (panic will crash server)
- 🔴 **No graceful shutdown** (data loss risk)
- 🔴 **Redundant authentication** (multiple validations per request)
- 🔴 **No monitoring/observability**

---

## Table of Contents

1. [Code Flow Analysis](#code-flow-analysis)
2. [Redundancies Identified](#redundancies-identified)
3. [Gaps in Logic](#gaps-in-logic)
4. [Conflicting Logic](#conflicting-logic)
5. [Security Concerns](#security-concerns)
6. [Performance Issues](#performance-issues)
7. [Production Readiness Checklist](#production-readiness-checklist)
8. [Implementation Roadmap](#implementation-roadmap)

---

## Code Flow Analysis

### Request Flow (Current)

```
HTTP Request
    ↓
Router.SetupRoutes() [router.go:46]
    ↓
Middleware: requireOwner/requireTenant [router.go:101-130]
    ├─ Load session cookie
    ├─ Validate session (AuthService.ValidateSession)
    ├─ Get user from DB (UserRepository.GetByID)
    └─ Check role (owner/tenant)
    ↓
Handler (e.g., RentalHandler.Dashboard)
    ├─ [REDUNDANT] Some handlers re-validate session
    ├─ [REDUNDANT] Some handlers re-fetch user from DB
    └─ Call service layer
        ↓
Service Layer (e.g., DashboardService)
    └─ Call repository layer
        ↓
Repository Layer (PostgreSQL)
    └─ Execute SQL queries
```

### Issues in Current Flow

1. **Double Validation:** Middleware validates session, but handlers also validate (e.g., `rental_handler.go:209-228`)
2. **Double User Lookup:** User fetched in middleware, then again in handler
3. **Inconsistent Context Usage:** Some handlers use context, others don't
4. **No Request Tracing:** No correlation IDs for debugging

---

## Redundancies Identified

### 1. Duplicate Session Validation

**Location:** `router.go` (middleware) + `rental_handler.go` + `tenant_handler.go`

**Problem:**
- Middleware validates session in `requireOwner`/`requireTenant`
- Handlers re-validate session again

**Example:**
```go
// router.go:104 - Middleware validates
user, err := r.loadSessionAndValidateRole(req, "owner")

// rental_handler.go:209-228 - Handler validates AGAIN
cookie, err := r.Cookie(h.cookieName)
sess, err := h.authService.ValidateSession(cookie.Value)
user, err := h.userRepo.GetByID(sess.UserID)
```

**Impact:**
- 2x database queries per request
- 2x session validation overhead
- Unnecessary latency

**Fix:**
- Remove duplicate validation from handlers
- Use `GetUserFromContext()` from middleware

---

### 2. Unused Middleware File

**Location:** `internal/http/middleware/auth.go`

**Problem:**
- Complete middleware implementation exists but is **never used**
- Router implements its own middleware in `router.go`

**Impact:**
- Dead code (maintenance burden)
- Confusion about which middleware to use

**Fix:**
- Either use `middleware/auth.go` OR remove it
- Consolidate to single middleware implementation

---

### 3. Commented-Out Deprecated Code

**Location:** `internal/service/payment_service.go` (lines 164-514)

**Problem:**
- Large blocks of commented-out code for deprecated methods
- Makes file harder to read
- Indicates incomplete refactoring

**Impact:**
- Code bloat
- Maintenance confusion

**Fix:**
- Remove all commented-out deprecated code
- Use git history for reference if needed

---

### 4. Inconsistent Context Key Access

**Location:** `router.go` vs `middleware/auth.go`

**Problem:**
- `router.go` uses string key: `context.WithValue(ctx, userContextKeyString, user)`
- `middleware/auth.go` uses string key: `context.WithValue(r.Context(), "user", user)`
- `router.go` has `GetUserFromContext()` that uses `UserKey` (type), but sets with string

**Impact:**
- Type safety issues
- Potential runtime errors

**Fix:**
- Standardize on typed context key
- Use single `GetUserFromContext()` function

---

### 5. Duplicate Payment Status Logic

**Location:** `domain/payment.go` + `handlers/rental_handler.go`

**Problem:**
- `Payment.GetStatus()` in domain model
- `getPaymentStatus()` helper in handler (line 392)

**Impact:**
- Logic duplication
- Inconsistency risk

**Fix:**
- Remove handler helper, use domain method

---

## Gaps in Logic

### 1. No Structured Logging

**Current State:**
- Uses `fmt.Println()` for logging
- Uses `log.Fatal()` for errors (crashes server)
- No log levels (DEBUG, INFO, WARN, ERROR)
- No structured fields (request ID, user ID, etc.)

**Impact:**
- Cannot debug production issues
- No audit trail
- Cannot monitor system health

**Fix:**
- Implement structured logging (e.g., `zerolog`, `zap`, or `slog`)
- Add request correlation IDs
- Log all important operations with context

---

### 2. No Error Recovery

**Current State:**
- No panic recovery middleware
- Panic will crash entire server
- No graceful error handling

**Impact:**
- Single request panic can bring down entire system
- No error reporting to monitoring systems

**Fix:**
- Add panic recovery middleware
- Log panics with stack traces
- Return 500 error instead of crashing

---

### 3. No Graceful Shutdown

**Current State:**
- Server uses `log.Fatal(http.ListenAndServe(...))`
- No shutdown signal handling
- No connection draining
- No cleanup on exit

**Impact:**
- Active requests lost on shutdown
- Database connections not closed properly
- Notification scheduler not stopped gracefully

**Fix:**
- Implement graceful shutdown with signal handling
- Drain active connections
- Close database connections
- Stop background workers

---

### 4. No Input Validation Middleware

**Current State:**
- Validation done in handlers/services
- Inconsistent validation
- No centralized validation rules

**Impact:**
- Security vulnerabilities (SQL injection risk if not careful)
- Inconsistent error messages
- Duplicate validation code

**Fix:**
- Add input validation middleware
- Use validation library (e.g., `go-playground/validator`)
- Centralize validation rules

---

### 5. No Rate Limiting

**Current State:**
- No rate limiting on any endpoints
- Login endpoint vulnerable to brute force
- API endpoints vulnerable to abuse

**Impact:**
- Brute force attacks possible
- DoS vulnerability
- Resource exhaustion

**Fix:**
- Add rate limiting middleware
- Use token bucket or sliding window algorithm
- Different limits for different endpoints

---

### 6. No CORS Handling

**Current State:**
- No CORS headers set
- Cannot be used from web frontend
- Security risk if misconfigured

**Impact:**
- Frontend integration issues
- Potential security vulnerabilities

**Fix:**
- Add CORS middleware
- Configure allowed origins
- Set appropriate headers

---

### 7. Hardcoded Values

**Location:** Multiple files

**Examples:**
- `payment_service.go:56`: `UPIID: "9848790200@ybl"` (hardcoded)
- `payment_service.go:55`: `PaymentMethod: "UPI"` (hardcoded)
- `auth_handler.go:111`: `cookieName: "sid"` (hardcoded in multiple places)

**Impact:**
- Cannot configure per environment
- Difficult to change without code changes

**Fix:**
- Move to configuration
- Use environment variables or config file

---

### 8. No Database Health Checks

**Current State:**
- Only ping on startup
- No periodic health checks
- No connection pool monitoring

**Impact:**
- Cannot detect database issues early
- Connection pool exhaustion not detected

**Fix:**
- Add health check endpoint
- Periodic database ping
- Monitor connection pool metrics

---

### 9. No Metrics/Monitoring

**Current State:**
- No metrics collection
- No performance monitoring
- No business metrics (payments processed, etc.)

**Impact:**
- Cannot monitor system health
- Cannot detect performance degradation
- No business intelligence

**Fix:**
- Add metrics collection (Prometheus)
- Instrument key operations
- Track business metrics

---

### 10. No API Versioning

**Current State:**
- All endpoints under `/api/*`
- No versioning strategy
- Breaking changes will affect all clients

**Impact:**
- Cannot evolve API without breaking clients
- Difficult to support multiple client versions

**Fix:**
- Add API versioning (e.g., `/api/v1/*`)
- Plan migration strategy

---

### 11. Inconsistent Error Responses

**Current State:**
- Some errors return JSON, others return plain text
- Inconsistent error format
- No error codes

**Impact:**
- Frontend cannot handle errors consistently
- Difficult to debug

**Fix:**
- Standardize error response format
- Use error codes
- Include request ID in errors

---

### 12. No Request Timeout Handling

**Current State:**
- No request timeout configuration
- Long-running queries can hang
- No cancellation support

**Impact:**
- Resource exhaustion
- Poor user experience

**Fix:**
- Add request timeout middleware
- Use context with timeout
- Cancel long-running operations

---

### 13. Session Cleanup Not Scheduled

**Current State:**
- Expired sessions only cleaned on login (`auth_service.go:61`)
- No background cleanup job
- Sessions accumulate in database

**Impact:**
- Database bloat
- Performance degradation over time

**Fix:**
- Add scheduled cleanup job
- Run periodically (e.g., daily)

---

### 14. No Password Strength Validation

**Current State:**
- Only checks length >= 6 (`tenant_handler.go:154`)
- No complexity requirements
- Weak passwords allowed

**Impact:**
- Security vulnerability
- Easy to brute force

**Fix:**
- Add password strength validation
- Require complexity (uppercase, lowercase, numbers, symbols)
- Minimum length 8-12 characters

---

### 15. No Audit Logging

**Current State:**
- No audit trail for sensitive operations
- Cannot track who did what and when

**Impact:**
- Compliance issues
- Cannot investigate security incidents

**Fix:**
- Add audit logging for:
  - Payment verifications
  - Tenant creation/deletion
  - Password changes
  - Login attempts

---

## Conflicting Logic

### 1. Payment Status Calculation

**Location:** `domain/payment.go` vs `handlers/rental_handler.go`

**Problem:**
- `Payment.GetStatus()` uses `IsPaid` field
- `Payment.GetUserFacingStatus()` uses `IsFullyPaid` and transactions
- Handler has `getPaymentStatus()` that uses different logic

**Conflict:**
- `IsPaid` is legacy field, `IsFullyPaid` is current
- Inconsistent status calculation

**Fix:**
- Remove `IsPaid` field (or mark as deprecated)
- Use only `IsFullyPaid` and `GetUserFacingStatus()`
- Remove handler helper function

---

### 2. Middleware Implementation Duplication

**Location:** `router.go` vs `middleware/auth.go`

**Problem:**
- Two different middleware implementations
- Router uses its own, middleware file unused

**Conflict:**
- Which one is correct?
- Maintenance confusion

**Fix:**
- Consolidate to single implementation
- Use middleware package, not router methods

---

### 3. Context Key Type Mismatch

**Location:** `router.go:163` vs `router.go:168`

**Problem:**
```go
// Sets with string
context.WithValue(ctx, userContextKeyString, user)

// Gets with type
user, ok := ctx.Value(UserKey).(*domain.User)
```

**Conflict:**
- Setting and getting use different keys
- Will always return `nil`

**Fix:**
- Use consistent key (prefer typed key)

---

### 4. Payment Creation Logic Duplication

**Location:** `payment_service.go` vs `tenant_service.go`

**Problem:**
- `PaymentService.CreatePaymentForTenant()` creates payment
- `TenantService.createFirstPayment()` also creates payment
- Similar but not identical logic

**Conflict:**
- Logic duplication
- Potential inconsistency

**Fix:**
- TenantService should call PaymentService method
- Single source of truth

---

## Security Concerns

### 1. Weak Password Hashing

**Location:** `auth_service.go:24-26`

**Problem:**
```go
func (s *AuthService) HashPassword(plain string) string {
    sum := sha256.Sum256([]byte(plain))
    return hex.EncodeToString(sum[:])
}
```

**Issue:**
- SHA256 is not suitable for password hashing
- No salt
- Vulnerable to rainbow table attacks
- Fast to compute (vulnerable to brute force)

**Fix:**
- Use `bcrypt` or `argon2id`
- Automatic salt generation
- Configurable cost factor

---

### 2. No Password Complexity Requirements

**Location:** `tenant_handler.go:154`

**Problem:**
- Only checks length >= 6
- No complexity requirements

**Fix:**
- Require: uppercase, lowercase, number, symbol
- Minimum length 8-12

---

### 3. No Rate Limiting on Login

**Location:** `auth_handler.go:34`

**Problem:**
- Login endpoint has no rate limiting
- Vulnerable to brute force attacks

**Fix:**
- Add rate limiting (e.g., 5 attempts per 15 minutes per IP)
- Lock account after multiple failures

---

### 4. Session Token Generation

**Location:** `auth_service.go:79-84`

**Problem:**
```go
func (s *AuthService) generateToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}
```

**Issue:**
- Uses `crypto/rand` (good)
- But hex encoding reduces entropy (64 hex chars = 256 bits, but only 4 bits per char)
- Should use base64url for better encoding

**Fix:**
- Use base64url encoding
- Or use JWT tokens with proper signing

---

### 5. Cookie Security

**Location:** `auth_handler.go:53`

**Problem:**
```go
http.SetCookie(w, &http.Cookie{
    Name: h.cookieName, 
    Value: sess.Token, 
    Path: "/", 
    HttpOnly: true, 
    Expires: sess.ExpiresAt
})
```

**Issues:**
- No `Secure` flag (should be true in production with HTTPS)
- No `SameSite` flag (CSRF vulnerability)
- No `Domain` restriction

**Fix:**
- Set `Secure: true` (HTTPS only)
- Set `SameSite: http.SameSiteStrictMode`
- Set `Domain` if needed

---

### 6. No CSRF Protection

**Current State:**
- No CSRF tokens
- Forms vulnerable to CSRF attacks

**Fix:**
- Add CSRF middleware
- Generate and validate CSRF tokens

---

### 7. SQL Injection Risk (Low, but present)

**Current State:**
- Uses parameterized queries (good)
- But no input sanitization middleware

**Fix:**
- Add input validation
- Sanitize all user inputs
- Use prepared statements (already done, but verify)

---

### 8. No HTTPS Enforcement

**Current State:**
- No HTTPS redirect
- No HSTS headers

**Fix:**
- Add HTTPS redirect middleware
- Set HSTS headers
- Use TLS in production

---

### 9. Sensitive Data in Logs

**Current State:**
- Configuration printed to console (main.go:33-48)
- Database URL partially printed (good, but still risky)

**Fix:**
- Never log passwords, tokens, or sensitive data
- Use log levels appropriately
- Mask sensitive data in logs

---

### 10. No Request ID for Tracing

**Current State:**
- No correlation IDs
- Cannot trace requests across services

**Fix:**
- Generate request ID in middleware
- Include in all logs
- Return in response headers

---

## Performance Issues

### 1. N+1 Query Problem

**Location:** Multiple services

**Example:** `tenant_service.go:119-136`
```go
for _, tenant := range tenants {
    if tenant.UnitID > 0 {
        unit, err := s.unitRepo.GetUnitByID(tenant.UnitID) // N queries!
    }
}
```

**Impact:**
- If 100 tenants, 100+ database queries
- Should be 1-2 queries with JOIN

**Fix:**
- Use JOIN queries in repository
- Batch load related data

---

### 2. No Database Query Caching

**Current State:**
- Every request hits database
- No caching layer

**Impact:**
- High database load
- Slower response times

**Fix:**
- Add caching for:
  - User sessions (Redis)
  - Unit data (if rarely changes)
  - Dashboard summaries (TTL cache)

---

### 3. No Connection Pooling Monitoring

**Current State:**
- Connection pool configured
- But no monitoring of pool usage

**Impact:**
- Cannot detect pool exhaustion
- No visibility into connection health

**Fix:**
- Add metrics for:
  - Active connections
  - Idle connections
  - Wait time for connections

---

### 4. Large Template Loading

**Location:** `main.go:18-23`

**Problem:**
- Templates loaded at startup
- No hot reload in development

**Impact:**
- Must restart server for template changes
- Slower development

**Fix:**
- Use `template.ParseGlob()` for dynamic loading
- Hot reload in development mode

---

### 5. No Response Compression

**Current State:**
- No gzip compression
- Large JSON responses not compressed

**Impact:**
- Higher bandwidth usage
- Slower response times

**Fix:**
- Add compression middleware
- Compress JSON responses

---

## Production Readiness Checklist

### Critical (Must Fix Before Production)

- [ ] **Security:**
  - [ ] Replace SHA256 with bcrypt/argon2 for passwords
  - [ ] Add rate limiting on login
  - [ ] Add CSRF protection
  - [ ] Set Secure and SameSite cookie flags
  - [ ] Add password complexity requirements
  - [ ] Implement HTTPS enforcement

- [ ] **Reliability:**
  - [ ] Add panic recovery middleware
  - [ ] Implement graceful shutdown
  - [ ] Add database health checks
  - [ ] Add request timeout handling

- [ ] **Observability:**
  - [ ] Implement structured logging
  - [ ] Add request correlation IDs
  - [ ] Add metrics collection
  - [ ] Add error tracking (Sentry, etc.)

### High Priority (Fix Soon)

- [ ] **Code Quality:**
  - [ ] Remove duplicate authentication logic
  - [ ] Remove unused middleware file
  - [ ] Remove commented-out deprecated code
  - [ ] Fix context key inconsistencies

- [ ] **Performance:**
  - [ ] Fix N+1 query problems
  - [ ] Add response compression
  - [ ] Add caching layer

- [ ] **Configuration:**
  - [ ] Move hardcoded values to config
  - [ ] Add environment-specific configs

### Medium Priority (Nice to Have)

- [ ] **Features:**
  - [ ] Add API versioning
  - [ ] Add audit logging
  - [ ] Add scheduled session cleanup
  - [ ] Add CORS middleware

- [ ] **Developer Experience:**
  - [ ] Add hot reload for templates
  - [ ] Add development mode flags
  - [ ] Improve error messages

---

## Implementation Roadmap

### Phase 1: Critical Security Fixes (Week 1)

1. **Replace Password Hashing**
   - Install `golang.org/x/crypto/bcrypt`
   - Update `AuthService.HashPassword()`
   - Migrate existing password hashes (on next login)

2. **Add Rate Limiting**
   - Install rate limiting library (e.g., `golang.org/x/time/rate`)
   - Add middleware for login endpoint
   - Configure limits per endpoint

3. **Fix Cookie Security**
   - Add `Secure` and `SameSite` flags
   - Make configurable per environment

4. **Add CSRF Protection**
   - Install CSRF library
   - Add middleware
   - Update forms

### Phase 2: Reliability & Observability (Week 2)

1. **Structured Logging**
   - Install `zerolog` or `zap`
   - Replace all `fmt.Println` and `log.*` calls
   - Add request correlation IDs
   - Configure log levels

2. **Panic Recovery**
   - Add recovery middleware
   - Log panics with stack traces
   - Return 500 errors

3. **Graceful Shutdown**
   - Add signal handling
   - Implement connection draining
   - Stop background workers
   - Close database connections

4. **Health Checks**
   - Add `/health` endpoint
   - Check database connectivity
   - Check external services

### Phase 3: Code Quality (Week 3)

1. **Remove Redundancies**
   - Remove duplicate session validation
   - Remove duplicate user lookups
   - Use context consistently

2. **Consolidate Middleware**
   - Use `middleware/auth.go` or remove it
   - Single middleware implementation

3. **Clean Up Code**
   - Remove commented-out deprecated code
   - Fix context key inconsistencies
   - Remove duplicate helper functions

### Phase 4: Performance & Scalability (Week 4)

1. **Fix N+1 Queries**
   - Add JOIN queries in repositories
   - Batch load related data
   - Add repository methods for bulk loading

2. **Add Caching**
   - Install Redis client
   - Cache user sessions
   - Cache dashboard summaries (TTL)

3. **Add Compression**
   - Add gzip middleware
   - Compress JSON responses

### Phase 5: Production Hardening (Week 5)

1. **Configuration Management**
   - Move hardcoded values to config
   - Add environment-specific configs
   - Validate config on startup

2. **Monitoring & Metrics**
   - Add Prometheus metrics
   - Instrument key operations
   - Add business metrics

3. **Documentation**
   - API documentation
   - Deployment guide
   - Runbook for operations

---

## Recommended Libraries

### Logging
- **zerolog** - Fast, structured logging
- **zap** - Uber's structured logger (alternative)

### Security
- **golang.org/x/crypto/bcrypt** - Password hashing
- **golang.org/x/time/rate** - Rate limiting
- **gorilla/csrf** - CSRF protection

### HTTP
- **gorilla/mux** - Better router (optional, current router works)
- **compress/gzip** - Response compression

### Monitoring
- **prometheus/client_golang** - Metrics
- **sentry-go** - Error tracking

### Caching
- **redis/go-redis** - Redis client

### Validation
- **go-playground/validator** - Input validation

---

## Code Examples

### Example: Structured Logging

```go
import "github.com/rs/zerolog/log"

// Instead of:
fmt.Println("Starting server...")

// Use:
log.Info().
    Str("port", cfg.Port).
    Str("env", cfg.Environment).
    Msg("Starting server")
```

### Example: Panic Recovery Middleware

```go
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Error().
                    Interface("error", err).
                    Str("path", r.URL.Path).
                    Stack().
                    Msg("Panic recovered")
                
                http.Error(w, "Internal server error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Example: Graceful Shutdown

```go
func main() {
    // ... setup code ...
    
    server := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: router,
    }
    
    // Graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan
        
        log.Info().Msg("Shutting down server...")
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := server.Shutdown(ctx); err != nil {
            log.Error().Err(err).Msg("Server shutdown error")
        }
        
        // Stop background workers
        notificationScheduler.Stop()
        
        // Close database
        db.Close()
        
        log.Info().Msg("Server stopped")
    }()
    
    log.Info().Str("port", cfg.Port).Msg("Server starting")
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal().Err(err).Msg("Server failed")
    }
}
```

### Example: Rate Limiting

```go
import "golang.org/x/time/rate"

var limiter = rate.NewLimiter(rate.Every(time.Minute/10), 5) // 5 requests per minute

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

---

## Conclusion

This analysis identifies **15+ critical issues**, **10+ security concerns**, and **5+ performance problems** that must be addressed before production deployment. The implementation roadmap provides a structured approach to fixing these issues over 5 weeks.

**Priority Order:**
1. **Security fixes** (Week 1) - Critical for protecting user data
2. **Reliability & Observability** (Week 2) - Essential for production operations
3. **Code Quality** (Week 3) - Reduces technical debt
4. **Performance** (Week 4) - Improves user experience
5. **Production Hardening** (Week 5) - Final polish

**Estimated Effort:** 5 weeks for full implementation

**Risk if Not Fixed:**
- Security breaches
- System crashes
- Data loss
- Poor user experience
- Compliance issues

---

**Next Steps:**
1. Review this document with team
2. Prioritize fixes based on business needs
3. Create GitHub issues for each fix
4. Begin implementation following roadmap
5. Set up staging environment for testing

---

*Document Version: 1.0*  
*Last Updated: 2024*

