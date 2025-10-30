# Rental Property Management System - PRD

## Project Overview
A web-based rental property management system for a 3-floor house with 6 rental units, managing tenants, rent collection, and property administration.

## Property Structure

### Floor Layout
- **Ground Floor (G):**
  - 1BHK - ₹7,000/month
  - 1 Single Room - ₹3,000/month

- **1st Floor:**
  - 1BHK - ₹9,000/month  
  - 1 Single Room - ₹2,500/month

- **2nd Floor:**
  - 2 Single Rooms - ₹1,500/month each

### Unit Identification
- G1: Ground Floor 1BHK (₹7,000)
- G2: Ground Floor Single Room (₹3,000)
- 1A: 1st Floor 1BHK (₹9,000)
- 1B: 1st Floor Single Room (₹2,500)
- 2A: 2nd Floor Single Room 1 (₹1,500)
- 2B: 2nd Floor Single Room 2 (₹1,500)

## User Roles

### Owner (Primary User)
- Login/authentication required
- Full access to all data
- Manage tenants, rent collection, reports

### Tenant (Future Feature)
- Limited access to their own data
- View rent history, payment status

## Data Models

### Tenant Information
**Primary Tenant (Rent Payer):**
- Name
- Phone
- Aadhar Number (required)
- Move-in Date
- Number of people staying
- Unit ID (which unit they rent)

**Family Members:**
- Name
- Age
- Relationship to primary tenant
- Aadhar Number (optional)

### Financial Data
- Monthly Rent Amount
- Security Deposit Amount: 1 month rent for all units
- Payment Due Dates:
  - Single Rooms: 5th of every month
  - 1BHK Units: 10th of every month
- Payment Method: UPI (9848790200@ybl)
- Payment Tracking: Paid status + Payment date

## Core Features

### Phase 1: Owner Dashboard
1. **Property Overview**
   - Total units: 6
   - Occupied/Vacant units
   - Monthly rent collection summary

2. **Tenant Management**
   - Add new tenant
   - View tenant details
   - Edit tenant information
   - Move-out tenant

3. **Rent Collection**
   - Mark rent as paid
   - View payment history
   - Track overdue payments
   - Generate rent receipts

4. **Reports**
   - Monthly collection report
   - Overdue payments
   - Tenant directory

### Phase 2: Advanced Features (Future)
- Tenant login portal
- Automated rent reminders
- Maintenance requests
- Document storage
- Financial analytics

## Technical Architecture

### Database Schema
- `units` table: Unit details, rent amounts
- `tenants` table: Primary tenant information
- `family_members` table: Additional family members
- `payments` table: Rent payment history
- `users` table: Owner authentication

### Service Layer
- `UnitService`: Manage property units
- `TenantService`: Tenant CRUD operations
- `PaymentService`: Rent collection and tracking
- `ReportService`: Generate reports and analytics

## Questions for Refinement

### Immediate Questions:
1. **Security Deposits:** What's the security deposit amount for each unit type?
2. **Payment Due Dates:** What are the specific due dates for each tenant?
3. **Late Fees:** Any late payment penalties?
4. **Lease Terms:** Fixed lease periods or month-to-month?

### Future Considerations:
5. **Maintenance:** Who handles repairs - owner or tenants?
6. **Utilities:** Are utilities included in rent or separate?
7. **Notice Period:** How much notice for move-out?
8. **Rent Increases:** Annual rent increase policy?

## Development Phases

### Phase 1: Core System (Current)
- Database setup with proper schema
- Service layer implementation
- Owner authentication
- Basic CRUD operations

### Phase 2: User Interface
- Owner dashboard
- Tenant management forms
- Payment tracking interface

### Phase 3: Advanced Features
- Reporting system
- Tenant portal
- Mobile responsiveness

### Phase 4: Production Features
- Data backup
- Security hardening
- Performance optimization

---

**Last Updated:** 2024-10-29
**Status:** In Development
**Next Steps:** Database schema design and service layer implementation
