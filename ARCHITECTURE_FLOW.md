# 🏗️ Rental Management System - Architecture Flow

## **📊 System Architecture Diagram**

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP REQUEST                            │
│                    (Browser → Server)                          │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HANDLERS LAYER                              │
│              (HTTP Request/Response Handling)                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              rental_handler.go                          │   │
│  │  • Dashboard() - Main dashboard view                    │   │
│  │  • UnitDetails() - Unit detail view                     │   │
│  │  • GetUnits() - API endpoint                            │   │
│  │  • CreateTenant() - API endpoint                        │   │
│  │  • VacateTenant() - API endpoint                        │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    SERVICE LAYER                               │
│                (Business Logic Orchestration)                  │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐   │
│  │  unit_service   │ │ tenant_service  │ │ payment_service │   │
│  │  • GetUnits()   │ │ • GetTenants()  │ │ • GetPayments() │   │
│  │  • CreateUnit() │ │ • CreateTenant()│ │ • CreatePayment()│   │
│  │  • UpdateUnit() │ │ • MoveOutTenant()│ │ • MarkPaid()    │   │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                  REPOSITORY LAYER                              │
│                    (Data Access)                               │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                INTERFACES                               │   │
│  │  • UnitRepository interface                            │   │
│  │  • TenantRepository interface                          │   │
│  │  • PaymentRepository interface                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                POSTGRES IMPLEMENTATIONS                 │   │
│  │  • postgres_unit_repository.go                         │   │
│  │  • postgres_tenant_repository.go                       │   │
│  │  • postgres_payment_repository.go                      │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DATABASE LAYER                              │
│                    (PostgreSQL)                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  • units table                                         │   │
│  │  • tenants table                                       │   │
│  │  • family_members table                                │   │
│  │  • payments table                                      │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## **🔄 Request Flow Example:**

### **1. User Clicks Unit → Unit Details**
```
Browser → /unit/1 → rental_handler.UnitDetails() → 
unitService.GetUnitByID() → unitRepo.GetUnitByID() → 
Database → Response → Template → HTML
```

### **2. User Adds Tenant → Create Tenant**
```
Browser → /api/tenants (POST) → rental_handler.CreateTenant() → 
tenantService.CreateTenant() → tenantRepo.CreateTenant() → 
Database → Response → Page Reload
```

### **3. User Vacates Tenant → Move Out**
```
Browser → /api/tenants/vacate (POST) → rental_handler.VacateTenant() → 
tenantService.MoveOutTenant() → paymentRepo.DeletePaymentsByTenantID() → 
tenantRepo.DeleteTenant() → unitRepo.UpdateUnitOccupancy() → 
Database → Response → Dashboard Redirect
```

## **📁 File Organization by Layer:**

### **🌐 Presentation Layer (Handlers)**
```
internal/handlers/
└── rental_handler.go          # All HTTP request handling
```

### **🧠 Business Layer (Services)**
```
internal/service/
├── unit_service.go            # Unit business logic
├── tenant_service.go          # Tenant business logic
└── payment_service.go         # Payment business logic
```

### **💾 Data Layer (Repositories)**
```
internal/repository/
├── interfaces/                # Repository contracts
│   ├── unit_repository.go
│   ├── tenant_repository.go
│   └── payment_repository.go
└── postgres/                  # PostgreSQL implementations
    ├── unit_repository.go
    ├── tenant_repository.go
    └── payment_repository.go
```

### **🏢 Domain Layer (Models)**
```
internal/domain/
├── unit.go                    # Unit business entity
├── tenant.go                  # Tenant business entity
├── family_member.go           # Family member entity
└── payment.go                 # Payment entity
```

### **🎨 Web Layer (Templates)**
```
web/templates/
├── dashboard.html             # Main dashboard
└── unit-detail.html          # Unit detail page
```

## **🔗 Dependency Flow:**

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Handlers  │───▶│   Services  │───▶│ Repositories│───▶│  Database   │
│             │    │             │    │             │    │             │
│ • HTTP      │    │ • Business  │    │ • Data      │    │ • Storage   │
│ • Templates │    │ • Logic     │    │ • Access    │    │ • Queries   │
│ • Validation│    │ • Rules     │    │ • SQL       │    │ • Tables    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Domain    │    │   Domain    │    │   Domain    │
│  Entities   │    │  Entities   │    │  Entities   │
│             │    │             │    │             │
│ • Unit      │    │ • Unit      │    │ • Unit      │
│ • Tenant    │    │ • Tenant    │    │ • Tenant    │
│ • Payment   │    │ • Payment   │    │ • Payment   │
└─────────────┘    └─────────────┘    └─────────────┘
```

## **✅ Benefits of This Structure:**

1. **Clear Separation** - Each layer has a single responsibility
2. **Easy Testing** - Can mock interfaces for unit testing
3. **Scalable** - Easy to add new features or change implementations
4. **Maintainable** - Changes in one layer don't affect others
5. **Readable** - Flow is obvious from file organization
6. **Go Best Practices** - Follows standard Go project layout

This structure makes the rental management system much easier to understand and maintain! 🏠✨
