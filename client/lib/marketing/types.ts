export type EmailTemplate = {
  id: string;
  name: string;
  subject: string;
  htmlBody: string;
  starterKey?: string;
};
export type Starter = {
  key: string;
  name: string;
  description: string;
  subject: string;
  htmlBody: string;
  color: string;
};
export type Options = {
  lists: { id: string; name: string; subscribersCount: number }[];
  templates: EmailTemplate[];
  senders: {
    id: string;
    name: string;
    fromEmail: string;
    supportsTLS: boolean;
  }[];
  starters: Starter[];
  /** Set when this server can show Cloudflare Turnstile on forms. */
  turnstileSiteKey?: string;
};
export type Newsletter = {
  id: string;
  name: string;
  description: string;
  subject: string;
  templateId: string;
  listId: string;
  smtpConfigId: string;
  status: string;
  cadence: string;
  timezone: string;
  nextSendAt: string | null;
  lastSentAt?: string;
  editions: number;
  postalAddress: string;
  /** Website address (xem.email/{publicSlug}); null when the site is off. */
  publicSlug?: string | null;
  /** The website's look: paper, broadsheet or midnight. */
  siteTheme?: string;
  /** An https:// image shown on the website. */
  siteLogoUrl?: string;
  /** Off sends editions without the open pixel or click-tracking links. */
  trackEngagement?: boolean;
};
export type FormField = {
  id?: string;
  Label: string;
  FieldType: string;
  Required: boolean;
  mapToContactField: string;
};
export type LeadForm = {
  theme?: Partial<import("./form-theme").FormTheme>;
  definition?: import("./form-definition").FormDefinition | null;
  version?: number;
  id: string;
  Name: string;
  description: string;
  Status: string;
  Slug: string;
  AddToListID: string;
  fields: FormField[];
  successMessage: string;
  SubmitButtonText: string;
  SubmissionCount: number;
  ViewCount: number;
  updatedAt: string;
};
export type FormAnalytics = {
  views: number;
  starts: number;
  completions: number;
  completionRate: number;
  partials: number;
  steps: {
    pageId: string;
    title: string;
    views: number;
    completions: number;
  }[];
  sources: { source: string; completions: number }[];
};
export type Contact = {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  company: string;
  phone: string;
  listId: string;
  lifecycleStage: string;
  status: string;
  createdAt: string;
};
export type CRMContactsPage = {
  data: (Contact & { listName: string })[];
  total: number;
  page: number;
  limit: number;
  summary: {
    total: number;
    qualified: number;
    customers: number;
    subscribed: number;
  };
};
export type WorkflowNode = {
  id: string;
  type: string;
  data: Record<string, any>;
};
export type WorkflowEdge = {
  id?: string;
  sourceId: string;
  targetId: string;
  label: string;
};
export type SendWindow = {
  enabled: boolean;
  start: string;
  end: string;
  days: number[];
};
export type Workflow = {
  id: string;
  name: string;
  description: string;
  triggerEvent: string;
  isActive: boolean;
  liveVersion?: number;
  hasDraft?: boolean;
  reentryMode?: string;
  reentryDays?: number;
  exitOnGoal?: boolean | null;
  sendWindow?: SendWindow | null;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  updatedAt: string;
};
export type WorkflowVersion = {
  number: number;
  triggerEvent: string;
  publishedAt: string;
  publishedBy: string;
};
export type StepStats = {
  nodeId: string;
  entered: number;
  waiting: number;
  movedOn: number;
  exited: number;
  conversions: number;
  sent: number;
  delivered: number | null;
  opens: number;
  clicks: number;
  unsubscribes: number;
  openRate: number;
  clickRate: number;
  unsubscribeRate: number;
};
