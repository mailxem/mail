# Race Condition and Edge Case Fixes

## Summary

This document describes the fixes applied to address critical race conditions and edge cases in the usage tracking system.

## Issues Fixed

### 1. Race Condition in RecordUsage (CRITICAL)

**Problem:**
Multiple concurrent requests could record usage that exceeds the limit.

**Scenario:**
```
Time   Request A                    Request B
----   --------------------------   --------------------------
T1     Read usage = 90             
T2                                  Read usage = 90
T3     Check: 90 + 5 = 95 < 100 ✓  
T4                                  Check: 90 + 8 = 98 < 100 ✓
T5     Update to 95                
T6                                  Update to 103 (EXCEEDS!)
```

**Solution:**
- Added `UpdateLimitWithCheck()` method to `UsageRepository` that atomically checks and updates at the database level
- Uses SQL `WHERE usage + ? <= ?` clause to ensure the update only succeeds if it won't exceed the limit
- Returns an error if no rows were affected (meaning limit would be exceeded)

**Code Changes:**
- `internal/repositories/subscription_repository.go`: Added `UpdateLimitWithCheck()` method
- `internal/services/subscription_service.go`: Updated `RecordUsage()` to use the new atomic method

### 2. Magic Number for Unlimited Quota

**Problem:**
Hardcoded `99999999` to represent unlimited quota is:
- Not documented
- Could be confused with actual large limits
- Fragile (what if someone uses 99999998?)

**Solution:**
- Added `UnlimitedQuota = -1` constant in `internal/models/plans.go`
- Updated logic to also check for `<= 0` as unlimited
- More semantic and clear intent

**Code Changes:**
- `internal/models/plans.go`: Added `UnlimitedQuota` constant
- `internal/services/subscription_service.go`: Updated `CheckUsageLimit()` to use the constant

## Database-Level Protection

The atomic update query looks like this:
```sql
UPDATE usage_records 
SET usage = usage + ?
WHERE subscription_id = ? 
  AND period = ?
  AND feature = ?
  AND usage + ? <= ?  -- This prevents race conditions!
```

If the WHERE clause fails (usage would exceed limit), zero rows are affected and we return an error.

## Usage Examples

### Setting Unlimited Quota
```go
feature := models.PlanFeatures{
    Feature: models.PlanFeatureQueries,
    Quota:   models.UnlimitedQuota, // -1 for unlimited
}
```

### Checking for Unlimited Quota
```go
if models.IsUnlimitedQuota(feature.Quota) {
    // Handle unlimited usage
}
// This helper function checks for:
// - models.UnlimitedQuota (-1)
// - Zero or negative values
// - Legacy value 99999999 (for backward compatibility)
```

### Recording Usage (now race-safe)
```go
err := subscriptionService.RecordUsage(ctx, teamID, structs.RecordUsageRequest{
    Feature: models.PlanFeatureQueries,
    Usage:   10,
})
if err != nil {
    // This error will be returned if the limit would be exceeded
    // Even if multiple concurrent requests try to record usage
}
```

## Testing Recommendations

### 1. Concurrency Test
```go
func TestRecordUsage_Concurrent(t *testing.T) {
    // Setup subscription with limit of 100
    
    var wg sync.WaitGroup
    errors := make(chan error, 20)
    
    // Try to record 20 concurrent requests of 10 each (total 200)
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            err := service.RecordUsage(ctx, teamID, structs.RecordUsageRequest{
                Feature: models.PlanFeatureQueries,
                Usage:   10,
            })
            if err != nil {
                errors <- err
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    // Get final usage
    stats, _ := service.GetUsageStats(ctx, subscriptionID)
    
    // Should be exactly 100, not more
    assert.LessOrEqual(t, stats.FeatureUsage[models.PlanFeatureQueries], 100)
    
    // Some requests should have failed
    assert.Greater(t, len(errors), 0)
}
```

### 2. Unlimited Quota Test
```go
func TestUnlimitedQuota(t *testing.T) {
    // Test with UnlimitedQuota constant
    feature := models.PlanFeatures{
        Feature: models.PlanFeatureQueries,
        Quota:   models.UnlimitedQuota,
    }
    
    // Should allow any amount
    for i := 0; i < 1000; i++ {
        err := service.RecordUsage(ctx, teamID, structs.RecordUsageRequest{
            Feature: models.PlanFeatureQueries,
            Usage:   1000,
        })
        assert.NoError(t, err)
    }
}
```

## Other Edge Cases Still Present

These were identified but not fixed (consider addressing):

1. **Typo on line 234**: `TrailFeatures` should be `TrialFeatures`
2. **Zero TotalSeats**: If `subscription.TotalSeats` is 0, quota becomes 0
3. **Inconsistent trial logic**: Different between `CheckUsageLimit` and `RecordUsage`
4. **Time zone inconsistency**: Mixed use of UTC and local time
5. **Integer overflow**: `feat.Quota * subscription.TotalSeats` could overflow

## Migration Notes

### Code Migration
The code changes are backward compatible and require no schema changes.

### Data Migration (Recommended)
If you have existing plans with quota value `99999999`, update them to use the new constant:

```sql
-- Run this SQL to update existing unlimited quotas
UPDATE plan_features 
SET quota = -1,
    updated_at = NOW()
WHERE quota = 99999999;
```

**Script Location:** `scripts/update_quota_to_unlimited.sql`

**Why migrate?**
- Semantic clarity: -1 is now the official unlimited value
- The code will work with both values, but -1 is preferred
- Prevents confusion between "unlimited" (99999999) and "very large limit"

**Safe to skip?**
Yes! The code maintains backward compatibility by checking for:
- `feat.Quota == models.UnlimitedQuota` (checks for -1, the new standard)
- `feat.Quota <= 0` (catches -1, 0, and other negative values)
- `feat.Quota == 99999999` (legacy unlimited value, for backward compatibility)

So existing 99999999 values will continue to work as unlimited, but migrating to -1 is recommended for clarity.

**When to remove 99999999 check?**
After all production data has been migrated, you can remove the `|| feat.Quota == 99999999` check from the code. Until then, both values are supported.

## Performance Impact

**Minimal:**
- The atomic update uses the same number of database queries
- WHERE clause adds negligible overhead
- No additional transactions or locks required

## Rollback Plan

If issues arise:
1. Revert `internal/services/subscription_service.go` to use `UpdateLimit()` instead of `UpdateLimitWithCheck()`
2. Race condition will return, but functionality remains
3. No data corruption risk

## Version Compatibility

- Requires GORM for the `RowsAffected` check
- PostgreSQL, MySQL, SQLite all supported
- No external dependency changes
