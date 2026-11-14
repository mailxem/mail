export enum SMTPProviderType {
  GMAIL = "GMAIL",
  CUSTOM = "CUSTOM",
  OUTLOOK = "OUTLOOK",
  AMAZON = "AMAZON",
}

export enum DomainStatus {
  PENDING = "pending",
  VERIFIED = "verified",
  FAILED = "failed",
}

export interface Domain {
  id: string;
  domain: string;
  status: DomainStatus;
  isVerified: boolean;
  isActive: boolean;
}

export enum WebhookEventType {
  SENT = "sent",
  DELIVERED = "delivered",
  OPEN = "open",
  CLICK = "click",
  REPLY = "reply",
  BOUNCE = "bounce",
  COMPLAINT = "complaint",
  UNSUBSCRIBED = "unsubscribed",
  FAILED = "failed",
}

export interface Webhook {
  id: string;
  name: string;
  url: string;
  events: WebhookEventType[];
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
  lastAttempt: Date;
  lastSuccess: Date;
  lastFailure: Date;
  failureCount: number;
  successCount: number;
  lastStatus: number;
  teamId: string;
  team?: Team;
}

export enum CampaignStatus {
  DRAFT = "draft",
  SCHEDULED = "scheduled",
  SENDING = "sending",
  COMPLETED = "completed",
  FAILED = "failed",
}

export interface APIKey {
  id: string;
  name: string;
  key: string;
  teamId: string;
  createdAt: string;
  lastUsedAt: string;
  updatedAt: string;
  expiresAt: string;
  isDeleted: boolean;
  /** Full access to the workspace's product data. */
  admin?: boolean;
}

/** A key's expiry, or null for a key that never expires (Go sends year 1). */
export function apiKeyExpiry(key: Pick<APIKey, "expiresAt">) {
  const at = key.expiresAt ? new Date(key.expiresAt) : null;
  return at && at.getUTCFullYear() > 1 ? at : null;
}

export interface APIKeyResponse {
  data: APIKey;
  error?: ApiError;
}

export interface DomainResponse {
  id: string;
  domain: string;
  verified: boolean;
  active: boolean;
}

export interface SMTPConfig {
  id: string;
  provider: SMTPProviderType;
  host: string;
  port: number;
  username: string;
  password: string;
  from: string;
  isDefault: boolean;
}

export interface EmailTemplate {
  id: string;
  name: string;
  subject: string;
  variables: string[];
  createdAt: Date;
  updatedAt: Date;
  categoryId: string;
  html?: string;
  teamId: string;
  team?: Team;
  design?: Record<string, any>;
  designJson?: string;
}

export interface EmailCategory {
  id: string;
  name: string;
}

export interface EmailTemplateCategory {
  id: string;
  name: string;
  templates: EmailTemplate[];
}

export interface TemplateEditorProps {
  templateId: string;
  starterKey?: string;
}

export interface User {
  id: string;
  firstName: string;
  lastName: string;
  name: string;
  email: string;
  role: "USER" | "ADMIN";
  teamId?: string;
}

export interface Team {
  id: string;
  name: string;
  users: User[];
  logo: string;
}

export interface TeamInvite {
  id: string;
  name: string;
  email: string;
  code: string;
  teamId: string;
  status: "PENDING" | "ACCEPTED" | "REJECTED";
  expiresAt: Date;
  inviterId: string;
}

export interface TeamInviteVerifyResponse {
  valid: boolean;
  email: string;
  name: string;
  role: string;
  team_name: string;
  expires_at: Date;
  error?: string;
}

export interface Contact {
  id: string;
  email: string;
  name?: string;
  metadata?: Record<string, any>;
  listId: string;
  teamId: string;
}

export interface MailingList {
  id: string;
  name: string;
  description?: string;
  teamId: string;
  subscribersCount?: number;
  listCleanup?: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export type CampaignAudience = "not_opened" | "not_clicked";

export interface Campaign {
  id: string;
  name: string;
  description?: string;
  status: "DRAFT" | "SCHEDULED" | "SENDING" | "COMPLETED" | "FAILED" | "PAUSED";
  /** Unset: staged automatically for never-emailed readers on managed sending. */
  sendInStages?: boolean;
  /** Why staged sending paused the campaign. */
  pauseReason?: string;
  /** When staged sending checks results and sends the next stage. */
  nextStageAt?: string;
  subject?: string;
  timezone?: string;
  /** Off sends without the open pixel or click-tracking links. */
  trackEngagement?: boolean;
  /** The newsletter this is an edition of, or whose website it's published on. */
  newsletterId?: string;
  archiveNewsletterId?: string;
  /** Signs the campaign; unset uses the workspace's default author. */
  authorId?: string;
  templateId: string;
  listId: string;
  /** Narrows the list to this segment's members, as a follow-up does. */
  segmentId?: string;
  schedule?: string;
  scheduledFor?: Date;
  recurringSchedule?: string;
  cronExpression?: string;
  smtpConfigId: string;
  replyTo?: string;
  data?: Record<string, string>;
  archiveNewsletterId?: string;
  /** Emails created so far; above 0 once the campaign has sent. */
  processed?: number;
  /** A follow-up: sends only to readers of this campaign who did not open (or click) it. */
  audienceCampaignId?: string;
  audienceFilter?: CampaignAudience;
  teamId: string;
  userId: string;
  analytics?: Record<string, any>;
  template?: EmailTemplate;
  createdAt: Date;
  updatedAt: Date;
}

export interface ApiError {
  message: string;
  code?: string;
  details?: Record<string, any>;
}
