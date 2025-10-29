# 🏗️ File Structure Restructuring Plan

## **Current Structure Issues:**
- Models mixed with other files in `internal/models/`
- Repository interfaces and implementations in same directory
- Templates in root-level `templates/` directory
- Documentation scattered in `cursor_rules/`

## **Proposed New Structure:**

```
backend-form/m/
├── cmd/                          # Application entry points
│   ├── server/
│   │   └── main.go              # Main application server
│   └── test/
│       └── main.go              # Test utilities
├── internal/                     # Private application code
│   ├── config/                   # Configuration management
│   │   ├── config.go            # Config struct and loading
│   │   └── env.go               # Environment variable handling
│   ├── domain/                   # Business entities (models)
│   │   ├── unit.go              # Unit business entity
│   │   ├── tenant.go            # Tenant business entity
│   │   ├── family_member.go     # Family member entity
│   │   └── payment.go           # Payment entity
│   ├── repository/               # Data access layer
│   │   ├── interfaces/          # Repository contracts (interfaces)
│   │   │   ├── unit_repository.go
│   │   │   ├── tenant_repository.go
│   │   │   └── payment_repository.go
│   │   └── postgres/            # PostgreSQL implementations
│   │       ├── unit_repository.go
│   │       ├── tenant_repository.go
│   │       └── payment_repository.go
│   ├── service/                  # Business logic layer
│   │   ├── unit_service.go      # Unit business logic
│   │   ├── tenant_service.go    # Tenant business logic
│   │   └── payment_service.go   # Payment business logic
│   └── handlers/                 # HTTP handlers (presentation layer)
│       └── rental_handler.go    # Rental management HTTP handlers
├── web/                          # Web assets and templates
│   └── templates/
│       ├── dashboard.html        # Main dashboard template
│       └── unit-detail.html     # Unit detail template
├── docs/                         # Documentation
│   └── cursor_rules/
│       ├── AI_ASSISTANT_RULES.md
│       ├── CODE_FLOW_SUMMARY.md
│       ├── CODING_RULES.md
│       ├── plan.md
│       ├── PROJECT_PREFERENCES.md
│       └── RENTAL_MANAGEMENT_PRD.md
├── greetings/                    # Example module (can be removed)
│   ├── go.mod
│   └── greetings.go
├── go.mod
└── go.sum
```

## **Benefits of New Structure:**

### **1. Clear Separation of Concerns:**
- **`domain/`** - Pure business entities, no dependencies
- **`repository/interfaces/`** - Data access contracts
- **`repository/postgres/`** - Database implementations
- **`service/`** - Business logic orchestration
- **`handlers/`** - HTTP request/response handling

### **2. Better Dependency Flow:**
```
handlers → service → repository/interfaces → repository/postgres
    ↓         ↓              ↓                    ↓
  domain ← domain ← domain ← domain
```

### **3. Easier Navigation:**
- **Domain entities** in one place
- **Repository contracts** separated from implementations
- **Web assets** in dedicated directory
- **Documentation** organized

### **4. Scalability:**
- Easy to add new database implementations
- Easy to add new handlers
- Easy to add new business logic
- Clear boundaries between layers

## **Migration Steps:**

### **Step 1: Create New Directories**
```bash
mkdir -p internal/domain
mkdir -p internal/repository/interfaces
mkdir -p internal/repository/postgres
mkdir -p web/templates
mkdir -p docs/cursor_rules
```

### **Step 2: Move Files**
```bash
# Move models to domain
mv internal/models/* internal/domain/

# Move repository interfaces
mv internal/repository/*_repository.go internal/repository/interfaces/

# Move PostgreSQL implementations
mv internal/repository/postgres_*_repository.go internal/repository/postgres/

# Move templates
mv templates/* web/templates/

# Move documentation
mv cursor_rules/* docs/cursor_rules/
```

### **Step 3: Update Import Paths**
- Update all import statements to reflect new paths
- Update `go.mod` if needed
- Update template parsing in `main.go`

### **Step 4: Clean Up**
- Remove empty directories
- Update documentation
- Test the application

## **Import Path Examples:**

### **Before:**
```go
import "backend-form/m/internal/models"
import "backend-form/m/internal/repository"
```

### **After:**
```go
import "backend-form/m/internal/domain"
import "backend-form/m/internal/repository/interfaces"
import "backend-form/m/internal/repository/postgres"
```

## **File Organization by Layer:**

### **Domain Layer (Business Entities):**
- `internal/domain/unit.go`
- `internal/domain/tenant.go`
- `internal/domain/family_member.go`
- `internal/domain/payment.go`

### **Repository Layer (Data Access):**
- **Interfaces:** `internal/repository/interfaces/`
- **Implementations:** `internal/repository/postgres/`

### **Service Layer (Business Logic):**
- `internal/service/unit_service.go`
- `internal/service/tenant_service.go`
- `internal/service/payment_service.go`

### **Handler Layer (HTTP):**
- `internal/handlers/rental_handler.go`

### **Web Layer (Templates):**
- `web/templates/dashboard.html`
- `web/templates/unit-detail.html`

This structure follows Go best practices and makes the application flow much clearer! 🚀
