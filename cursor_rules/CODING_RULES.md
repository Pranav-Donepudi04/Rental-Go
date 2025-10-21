# Coding Rules and Guidelines

<!-- This file serves as persistent rules that should be followed across all development sessions -->

## General Development Principles

### Code Quality Standards
- **Always write clean, readable code** with meaningful variable and function names
- **Follow the DRY principle** (Don't Repeat Yourself)
- **Write self-documenting code** - code should be readable without extensive comments
- **Use consistent formatting** and follow language-specific style guides
- **Implement proper error handling** - never ignore errors silently

### Documentation Requirements
- **Comment complex logic** and business rules
- **Write comprehensive README files** for all projects
- **Document API endpoints** with clear examples
- **Keep documentation up-to-date** with code changes

### Testing Standards
- **Write tests for all business logic** before implementation
- **Aim for high test coverage** (minimum 80%)
- **Use descriptive test names** that explain the scenario
- **Test edge cases and error conditions**
- **Write integration tests** for critical user flows

## Project-Specific Rules

### Go Development Standards
- **Follow Go idioms** and conventions (gofmt, golint)
- **Use interfaces** for abstraction and testability
- **Implement proper package structure** (cmd/, internal/, pkg/)
- **Use context.Context** for request-scoped values
- **Handle errors explicitly** - don't use panic for regular error conditions
- **Write godoc comments** for all exported functions and types

### API Design Principles
- **Use RESTful conventions** for HTTP APIs
- **Return appropriate HTTP status codes**
- **Implement proper input validation**
- **Use structured error responses**
- **Version your APIs** from the start
- **Implement rate limiting** for production APIs

### Database and Storage
- **Use database transactions** for multi-step operations
- **Implement proper connection pooling**
- **Use prepared statements** to prevent SQL injection
- **Design for scalability** from the beginning
- **Implement proper indexing** strategies

## Security Guidelines

### Data Protection
- **Never log sensitive information** (passwords, tokens, personal data)
- **Use environment variables** for configuration and secrets
- **Implement proper authentication** and authorization
- **Validate all input** from external sources
- **Use HTTPS** in production environments

### Code Security
- **Keep dependencies updated** and scan for vulnerabilities
- **Use secure coding practices** to prevent common vulnerabilities
- **Implement proper session management**
- **Use strong encryption** for sensitive data at rest

## Performance and Scalability

### Optimization Guidelines
- **Profile before optimizing** - measure first, then optimize
- **Use caching strategies** appropriately
- **Implement database query optimization**
- **Use connection pooling** for external services
- **Monitor application performance** in production

### Scalability Considerations
- **Design for horizontal scaling** where possible
- **Use stateless design patterns**
- **Implement proper caching layers**
- **Consider microservices architecture** for large applications

## Deployment and DevOps

### Environment Management
- **Use infrastructure as code** (Terraform, CloudFormation)
- **Implement proper CI/CD pipelines**
- **Use containerization** (Docker) for consistency
- **Implement proper logging and monitoring**
- **Use configuration management** tools

### Production Readiness
- **Implement health checks** and monitoring
- **Use proper logging levels** and structured logging
- **Implement graceful shutdown** handling
- **Use proper backup strategies**
- **Plan for disaster recovery**

## Code Review Standards

### Review Checklist
- **Code follows established patterns** and conventions
- **Tests are comprehensive** and meaningful
- **Documentation is updated** and accurate
- **Security considerations** are addressed
- **Performance implications** are considered
- **Error handling** is proper and complete

### Review Process
- **All code changes require review** before merging
- **Use meaningful commit messages** following conventional commits
- **Break large changes** into smaller, reviewable chunks
- **Address all feedback** before requesting re-review

## Version Control Best Practices

### Git Workflow
- **Use feature branches** for all development
- **Write clear commit messages** with conventional commit format
- **Keep commits atomic** - one logical change per commit
- **Use pull requests** for all changes
- **Maintain a clean git history**

### Branch Management
- **Use descriptive branch names** (feature/, bugfix/, hotfix/)
- **Delete merged branches** to keep repository clean
- **Use tags** for releases and important milestones
- **Protect main/master branch** from direct pushes

---

**Last Updated**: [Auto-updated on each session]
**Version**: 1.0
**Maintained By**: Development Team

> **Note**: These rules should be reviewed and updated regularly to reflect evolving best practices and project needs.

