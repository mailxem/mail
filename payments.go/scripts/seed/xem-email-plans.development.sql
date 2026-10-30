BEGIN;
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-free', 'xem-free', 'Free', '3000 emails a month. No card needed.', 'active', 0/100.0, 'USD', 0/100.0, 'MONTHLY', 0, false, 1, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-free-emails', 'xem-free', 3000, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-starter', 'pdt_0NqqGJnJQcN9N86P02vbW', 'Xem Starter', '20000 emails a month included, then $0.40 per 1,000.', 'active', 900/100.0, 'USD', 900/100.0, 'MONTHLY', 0, false, 2, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-starter-emails', 'xem-starter', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-growth', 'pdt_0NqqGJoGLrw3QmYPYXVBH', 'Xem Growth', '100000 emails a month included, then $0.30 per 1,000.', 'active', 2900/100.0, 'USD', 2900/100.0, 'MONTHLY', 0, false, 3, true, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-growth-emails', 'xem-growth', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
INSERT INTO plans (id, dodo_plan_id, name, description, status, price, currency, price_per_seat, billing_cycle, trial_period_days, trial_enabled, display_order, is_popular, is_enterprise, metadata, created_at, updated_at)
VALUES ('xem-scale', 'pdt_0NqqGJqdeU0DZQmZPbKaN', 'Xem Scale', '500000 emails a month included, then $0.20 per 1,000.', 'active', 9900/100.0, 'USD', 9900/100.0, 'MONTHLY', 0, false, 4, false, false, '{}', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET dodo_plan_id = EXCLUDED.dodo_plan_id, name = EXCLUDED.name, description = EXCLUDED.description, status = 'active', price = EXCLUDED.price, price_per_seat = EXCLUDED.price_per_seat, display_order = EXCLUDED.display_order, is_popular = EXCLUDED.is_popular, updated_at = NOW();
INSERT INTO plan_features (id, plan_id, quota, feature, created_at, updated_at, trail)
VALUES ('xem-scale-emails', 'xem-scale', -1, 'emails_per_month', NOW(), NOW(), false)
ON CONFLICT (id) DO UPDATE SET quota = EXCLUDED.quota, updated_at = NOW();
UPDATE plans SET status = 'inactive', updated_at = NOW() WHERE id NOT IN ('xem-free', 'xem-starter', 'xem-growth', 'xem-scale') AND status = 'active';
COMMIT;
