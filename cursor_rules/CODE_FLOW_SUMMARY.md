# Rental Management System - Code Flow Summary

## 🏗️ Architecture Overview

### **3-Layer Architecture:**
1. **Handlers Layer** (`internal/handlers/`) - HTTP request/response handling
2. **Service Layer** (`internal/service/`) - Business logic and orchestration  
3. **Repository Layer** (`internal/repository/`) - Data access and persistence

---

## 📊 Data Models

### **Core Entities:**
- **Unit**: Property units (G1, G2, 1A, 1B, 2A, 2B) with rent, occupancy status
- **Tenant**: Primary rent payer with personal details, family members
- **Payment**: Rent payments with due dates, payment status, UPI tracking
- **FamilyMember**: Additional occupants linked to primary tenant

### **Database Schema:**
- **units** table: Unit details, rent amounts, occupancy status
- **tenants** table: Primary tenant information, unit assignment
- **family_members** table: Additional occupants (foreign key to tenants)
- **payments** table: Payment records (foreign key to tenants)

---

## 🔄 Request Flow

### **1. HTTP Request → Handler**
```
Client Request → RentalHandler → Service Layer → Repository Layer → Database
```

### **2. Main Entry Points:**
- **`/` or `/dashboard`** → `Dashboard()` - Main property overview
- **`/api/units`** → `GetUnits()` - Unit data API
- **`/api/tenants`** → `GetTenants()` / `CreateTenant()` - Tenant management
- **`/api/payments`** → `GetPayments()` - Payment data API
- **`/api/tenants/vacate`** → `VacateTenant()` - Tenant move-out

---

## 🎯 Current Dashboard Features

### **Property Units Section:**
- Lists all 6 units (G1, G2, 1A, 1B, 2A, 2B)
- Shows occupancy status (Available/Occupied)
- Displays tenant info for occupied units
- **Vacate button** for occupied units

### **Statistics Cards:**
- Total Units, Occupied Units
- Monthly Revenue, Total Potential Revenue

### **Recent Payments Section:**
- Lists all payments with status (Paid/Pending/Overdue)
- Payment amount and due date information

### **Quick Actions:**
- Add New Tenant (modal form)
- Mark Payment as Paid
- Refresh Data

---

## 🔧 Service Layer Responsibilities

### **UnitService:**
- Unit CRUD operations
- Occupancy management
- Rental summary calculations

### **TenantService:**
- Tenant CRUD operations
- Family member management
- Move-in/move-out processes
- **Foreign key constraint handling** (deletes payments before tenant)

### **PaymentService:**
- Payment CRUD operations
- Payment status management
- Payment summaries and reports

---

## 🗄️ Repository Layer

### **PostgreSQL Implementation:**
- **UnitRepository**: Unit data persistence
- **TenantRepository**: Tenant and family member data
- **PaymentRepository**: Payment records and history
- **Foreign key relationships** properly handled

---

## 🎨 Frontend (Templates)

### **Dashboard Template (`dashboard.html`):**
- **Responsive design** with CSS Grid
- **Real-time data** from Go templates
- **JavaScript functions** for AJAX operations
- **Modal forms** for tenant creation
- **Dynamic unit status** display

### **Template Data Flow:**
```
Go Handler → Template Data → HTML Rendering → JavaScript Interactions
```

---

## 🔄 Current Data Flow Example

### **Adding a Tenant:**
1. **Frontend**: User fills modal form
2. **JavaScript**: Sends POST to `/api/tenants`
3. **Handler**: `CreateTenant()` validates data
4. **Service**: `TenantService.CreateTenant()` business logic
5. **Repository**: `CreateTenant()` database insert
6. **Response**: Success/error back to frontend
7. **UI Update**: Page refresh with new data

### **Vacating a Tenant:**
1. **Frontend**: User clicks "Vacate" button
2. **JavaScript**: Confirms action, sends POST to `/api/tenants/vacate`
3. **Handler**: `VacateTenant()` processes request
4. **Service**: `MoveOutTenant()` deletes payments first, then tenant
5. **Repository**: Database operations (payments → tenant → unit update)
6. **Response**: Success confirmation
7. **UI Update**: Page refresh, unit shows as available

---

## 🚀 Next Steps for Enhancement

### **Requested Changes:**
1. **Remove tenant name from vacate button** (just "Vacate")
2. **Remove Recent Payments section** from dashboard
3. **Add unit detail page** - Click unit → detailed tenant info + payment history
4. **Create new route** `/unit/{id}` for unit details
5. **New template** `unit-detail.html` for detailed view

### **Implementation Plan:**
1. Update dashboard template (remove payments, simplify vacate button)
2. Create unit detail handler and template
3. Add click handlers to units in dashboard
4. Implement unit detail page with tenant info and payment history
5. Add navigation between dashboard and unit details

---

## 📁 File Structure
```
cmd/server/main.go              # Application entry point
internal/
├── handlers/
│   ├── rental_handler.go       # Main HTTP handlers
│   └── form_handler.go         # Legacy form handler
├── service/
│   ├── unit_service.go         # Unit business logic
│   ├── tenant_service.go       # Tenant business logic
│   └── payment_service.go      # Payment business logic
├── repository/
│   ├── postgres_unit_repository.go
│   ├── postgres_tenant_repository.go
│   └── postgres_payment_repository.go
└── models/
    ├── unit.go
    ├── tenant.go
    ├── family_member.go
    └── payment.go
templates/
├── dashboard.html              # Main dashboard
├── form.html                   # Legacy form
└── success.html                # Form success page
```
