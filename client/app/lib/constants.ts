export const IS_DEV = process.env.APP_ENV === "development";
export const UNLIMITED_QUOTA = -1;
export const FREE_PLAN_ID = "starter-plan-001";
export const PRO_MONTHLY_PLAN_ID = "pro-monthly-001";
export const PRO_YEARLY_PLAN_ID = "pro-yearly-001";
export const ENTERPRISE_PLAN_ID = "enterprise-plan-001";
// features
export const DATA_SOURCES_FEATURE = "data_sources";
export const QUERIES_FEATURE = "queries";
export const CHATS_FEATURE = "chat";
export const MEMBERS_FEATURE = "members";

// subscription statuses (aligned with dodopayments-go)
export const SUBSCRIPTION_STATUS_PENDING = "pending";
export const SUBSCRIPTION_STATUS_ACTIVE = "active";
export const SUBSCRIPTION_STATUS_ON_HOLD = "on_hold";
export const SUBSCRIPTION_STATUS_PAUSED = "paused";
export const SUBSCRIPTION_STATUS_CANCELLED = "cancelled";
export const SUBSCRIPTION_STATUS_FAILED = "failed";
export const SUBSCRIPTION_STATUS_EXPIRED = "expired";

// routes
export const CONNECTION_ADD_ROUTE = "/api/connections";
export const API_TEAM_INVITE = "/api/team/invite";
export const MIDDLEWARE_TS = "middleware.ts";
export const API_AUTH = "api/auth";
export const NEXT_STATIC = "_next/static";
export const NEXT_IMAGE = "_next/image";
export const FAVICON_ICO = "favicon.ico";
export const LOGIN_ROUTE = "/auth/login";
export const ONBOARDING_ROUTE = "/onboarding";
export const ACCEPT_INVITATION = "accept-invitation";
export const AUTH = "auth*";
export const SHARED = "shared";
export const API_SHARED_PAGES = "api/shared-pages";
export const API_CHAT = "api/chat";
export const PUBLIC = "public";
export const ASSETS = "assets";
export const API_AI_SEMANTIC_MODEL = "api/ai/semantic-model";
export const API_INTEGRATIONS_SLACK_WEBHOOK = "api/integrations/slack/webhook";
export const POST_METHOD = "POST";
export const TEAM_ID_KEY = "teamId";
export const FEATURE_KEY = "feature";
export const USAGE_KEY = "usage";
export const IP_ADDRESS_ALL = "0.0.0.0/0";
export const HTTP_OK = 200;
export const API_USAGE_CHECK = "usage/check";
export const API_USAGE_RECORD = "usage/record";