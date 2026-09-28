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
  DELIVERY = "delivery",
  BOUNCE = "bounce",
  COMPLAINT = "complaint",
  CLICK = "click",
  OPEN = "open",
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
}

export interface APIKeyRequest {
  name: string;
  expiresAt?: Date;
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
  createdAt: Date;
  updatedAt: Date;
}

export interface Campaign {
  id: string;
  name: string;
  description?: string;
  status: "DRAFT" | "SCHEDULED" | "SENDING" | "COMPLETED" | "FAILED";
  templateId: string;
  listId: string;
  schedule?: string;
  scheduledFor?: Date;
  recurringSchedule?: string;
  cronExpression?: string;
  smtpConfigId: string;
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
