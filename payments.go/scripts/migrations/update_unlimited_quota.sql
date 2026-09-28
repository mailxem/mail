-- Migration: Update unlimited quota values from magic number to constant
-- Description: Replace all quota values of 99999999 with -1 (UnlimitedQuota constant)
-- Date: 2026-01-15

BEGIN;

-- Update plan_features table to use new unlimited quota constant
UPDATE plan_features 
SET quota = -1 
WHERE quota = 99999999;

-- Verify the update
SELECT 
    COUNT(*) as updated_count,
    'plan_features' as table_name
FROM plan_features 
WHERE quota = -1;

COMMIT;

-- Rollback query (if needed):
-- BEGIN;
-- UPDATE plan_features SET quota = 99999999 WHERE quota = -1;
-- COMMIT;
