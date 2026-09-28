-- Quick SQL to update all 99999999 quota values to -1 (UnlimitedQuota)
-- Run this after deploying the code changes

-- Preview what will be updated (run this first to see what will change)
SELECT 
    id,
    plan_id,
    feature,
    quota,
    trail
FROM plan_features 
WHERE quota = 99999999;

-- Update the values
UPDATE plan_features 
SET quota = -1,
    updated_at = NOW()
WHERE quota = 99999999;

-- Verify the update
SELECT 
    id,
    plan_id,
    feature,
    quota,
    trail
FROM plan_features 
WHERE quota = -1;
