BEGIN;
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-free', 'xem-free', 'Free', '3000 emails a month. No card needed.', 'active', 0/100.0, 'USD', 0/100.0, 'MONTHLY', 0, false, 1, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-free-emails', 'xem-free', 3000, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-starter', 'pdt_0NrErmGUDDA4aSFZQ3iUU', 'Xem Starter', '20000 emails a month included, then $0.50 per 1,000.', 'active', 900/100.0, 'USD', 900/100.0, 'MONTHLY', 0, false, 2, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-starter-emails', 'xem-starter', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-growth', 'pdt_0NqqGJoGLrw3QmYPYXVBH', 'Xem Growth', '100000 emails a month included, then $0.30 per 1,000.', 'active', 2900/100.0, 'USD', 2900/100.0, 'MONTHLY', 0, false, 3, true, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-growth-emails', 'xem-growth', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-scale', 'pdt_0NqqGJqdeU0DZQmZPbKaN', 'Xem Scale', '500000 emails a month included, then $0.20 per 1,000.', 'active', 9900/100.0, 'USD', 9900/100.0, 'MONTHLY', 0, false, 4, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-scale-emails', 'xem-scale', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-starter-yearly', 'pdt_0NrErmPNi1hTO0XXGL6Vz', 'Xem Starter (Yearly)', '240000 emails a year included, then $0.50 per 1,000.', 'active', 9000/100.0, 'USD', 9000/100.0, 'YEARLY', 0, false, 5, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-starter-yearly-emails', 'xem-starter-yearly', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-growth-yearly', 'pdt_0NrEmxjTWIccWzp7PbW9m', 'Xem Growth (Yearly)', '1200000 emails a year included, then $0.30 per 1,000.', 'active', 29000/100.0, 'USD', 29000/100.0, 'YEARLY', 0, false, 6, true, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-growth-yearly-emails', 'xem-growth-yearly', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-scale-yearly', 'pdt_0NrEmxkLv6pvqM2a1XU0e', 'Xem Scale (Yearly)', '6000000 emails a year included, then $0.20 per 1,000.', 'active', 99000/100.0, 'USD', 99000/100.0, 'YEARLY', 0, false, 7, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, billing_cycle = EXCLUDED.billing_cycle, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-scale-yearly-emails', 'xem-scale-yearly', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
UPDATE subscriptions SET plan_id = 'xem-free' WHERE plan_id = 'solo-plan-001';
UPDATE plans SET status = 'inactive', updated_at = NOW() WHERE id NOT IN ('xem-free', 'xem-starter', 'xem-growth', 'xem-scale', 'xem-starter-yearly', 'xem-growth-yearly', 'xem-scale-yearly') AND status = 'active';
COMMIT;
