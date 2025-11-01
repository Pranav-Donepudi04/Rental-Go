# Existing Tenant Migration Guide

This guide explains how to handle tenants who have been staying for a while and had their payments managed manually, when transitioning them into the system.

## Problem

When you add an existing tenant to the system:
- The system automatically creates a payment based on their move-in date
- But they've already paid several months manually
- The system shows those months as pending/overdue (which is incorrect)

## Solution

The system now provides two approaches:

### Option 1: Skip First Payment Creation (Recommended)

When creating an existing tenant, set `is_existing_tenant: true` to skip automatic payment creation, then manually sync their payment history.

**Steps:**

1. **Create tenant with `is_existing_tenant: true`**

```json
POST /api/tenants
{
  "name": "John Doe",
  "phone": "9876543210",
  "aadhar_number": "123456789012",
  "move_in_date": "2024-01-15",  // Original move-in date
  "number_of_people": 2,
  "unit_id": 1,
  "is_existing_tenant": true  // ⭐ This skips first payment creation
}
```

2. **Sync payment history**

```json
POST /api/payments/sync-history
{
  "tenant_id": 123,
  "payments": [
    {
      "month": 1,           // January (1-12)
      "year": 2024,
      "payment_date": "2024-01-20T00:00:00Z",
      "notes": "Paid manually via cash"
    },
    {
      "month": 2,           // February
      "year": 2024,
      "payment_date": "2024-02-15T00:00:00Z",
      "notes": "Paid via UPI"
    },
    {
      "month": 3,           // March
      "year": 2024,
      "payment_date": "2024-03-18T00:00:00Z",
      "notes": ""
    }
  ]
}
```

3. **Adjust first unpaid payment date (if needed)**

If the auto-created first payment's due date doesn't match your payment schedule:

```json
POST /api/payments/adjust-due-date
{
  "tenant_id": 123,
  "due_date": "2024-05-10"  // Next payment due date
}
```

### Option 2: Mark Existing Payment as Paid

If you already created the tenant with auto-payment, you can mark the existing payment as paid:

```json
POST /api/payments/mark-paid
{
  "payment_id": 456,
  "payment_date": "2024-03-15",
  "notes": "Manual payment entry"
}
```

Then sync the remaining historical payments using Option 1's sync-history endpoint.

## Example: Complete Migration Workflow

**Scenario:** Tenant moved in on Jan 15, 2024. Payments for Jan, Feb, Mar were done manually. Now it's April 10, 2024, and you want to add them to the system.

### Step 1: Create Tenant (Skip Payment Creation)

```bash
curl -X POST http://localhost:8080/api/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "phone": "9876543210",
    "aadhar_number": "123456789012",
    "move_in_date": "2024-01-15",
    "number_of_people": 2,
    "unit_id": 1,
    "is_existing_tenant": true
  }'
```

Response:
```json
{
  "success": true,
  "message": "Tenant created successfully",
  "tenant": {
    "id": 123,
    "name": "John Doe",
    ...
  },
  "temp_password": "A3B7K2"
}
```

### Step 2: Sync Historical Payments

```bash
curl -X POST http://localhost:8080/api/payments/sync-history \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 123,
    "payments": [
      {
        "month": 1,
        "year": 2024,
        "payment_date": "2024-01-20T00:00:00Z",
        "notes": "January rent - cash"
      },
      {
        "month": 2,
        "year": 2024,
        "payment_date": "2024-02-15T00:00:00Z",
        "notes": "February rent - UPI"
      },
      {
        "month": 3,
        "year": 2024,
        "payment_date": "2024-03-18T00:00:00Z",
        "notes": "March rent - UPI"
      }
    ]
  }'
```

Response:
```json
{
  "success": true,
  "message": "Synced 3 payment(s)",
  "payments": [
    {
      "id": 789,
      "tenant_id": 123,
      "amount": 15000,
      "amount_paid": 15000,
      "remaining_balance": 0,
      "due_date": "2024-01-10T00:00:00Z",
      "payment_date": "2024-01-20T00:00:00Z",
      "is_fully_paid": true,
      ...
    },
    ...
  ]
}
```

### Step 3: Create Current Payment (if needed)

If there's an unpaid month (e.g., April), the system will auto-create it when needed, or you can create it manually. The payment for April will show as pending if not yet paid.

### Step 4: Verify

Check tenant dashboard or unit detail page to see:
- ✅ January: Fully Paid
- ✅ February: Fully Paid
- ✅ March: Fully Paid
- ⏳ April: Pending (if not yet paid)

## API Reference

### POST /api/payments/sync-history

Syncs historical payment records for an existing tenant.

**Request:**
```json
{
  "tenant_id": 123,
  "payments": [
    {
      "month": 1,           // Required: 1-12
      "year": 2024,         // Required: year
      "payment_date": "2024-01-20T00:00:00Z",  // Required: ISO 8601 format
      "notes": "Optional notes"  // Optional: string
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Synced 3 payment(s)",
  "payments": [...]
}
```

**Behavior:**
- If payment already exists for that month, it will be marked as paid (if not already)
- If payment doesn't exist, it creates a new payment record marked as fully paid
- Uses the unit's monthly rent amount automatically
- Uses the unit's payment due day for the due date

### POST /api/payments/adjust-due-date

Adjusts the due date of the first unpaid payment for a tenant.

**Request:**
```json
{
  "tenant_id": 123,
  "due_date": "2024-05-10"  // Format: YYYY-MM-DD
}
```

**Response:**
```json
{
  "success": true,
  "message": "Payment due date adjusted successfully"
}
```

### POST /api/tenants (with existing tenant flag)

Create a tenant with option to skip first payment creation.

**Request:**
```json
{
  "name": "John Doe",
  "phone": "9876543210",
  "aadhar_number": "123456789012",
  "move_in_date": "2024-01-15",
  "number_of_people": 2,
  "unit_id": 1,
  "is_existing_tenant": true  // Optional: defaults to false
}
```

## Best Practices

1. **Use `is_existing_tenant: true`** when adding tenants who have already paid
2. **Sync all historical payments** before marking any new payments
3. **Verify payment history** after syncing to ensure accuracy
4. **Add notes** to historical payments for record-keeping
5. **Use actual payment dates** rather than due dates when possible

## Common Issues

### Issue: Payment already exists error

**Cause:** You're trying to sync a payment for a month that already has a payment record.

**Solution:** The sync endpoint will automatically mark the existing payment as paid if it's not already paid. If it's already paid, it will just return the existing payment.

### Issue: Wrong due date on first payment

**Cause:** Auto-created payment's due date doesn't match actual payment schedule.

**Solution:** Use `/api/payments/adjust-due-date` to adjust the due date of the first unpaid payment.

### Issue: Missing months in payment history

**Cause:** You didn't sync all paid months.

**Solution:** Call `/api/payments/sync-history` again with the missing months. The endpoint is idempotent and safe to call multiple times.

## Next Steps

After migrating an existing tenant:
1. Share login credentials (phone + temp password)
2. Tenant can submit new payments via transaction ID
3. System will track future payments automatically

