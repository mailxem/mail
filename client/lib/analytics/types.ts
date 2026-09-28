export type Rate = {
  value: number | null;
  numerator: number;
  denominator: number;
};
export type Summary = {
  activeSubscribers: number;
  newContactRecords: number;
  recipientsReached: number;
  recipientsClicked: number;
  accepted: number;
  opened: number;
  clickedMessages: number;
  bounced: number;
  complaints: number;
  unsubscribes: number;
  failed: number;
  deliveryUnknown: number;
  trackingKnown: number;
  trackingEnabled: number;
  audienceClickRate: Rate;
  clickRate: Rate;
  openRate: Rate;
  bounceRate: Rate;
  complaintRate: Rate;
};
export type Report = {
  metricVersion: "audience-v2";
  asOf: string;
  from: string;
  to: string;
  timezone: string;
  summary: Summary;
  previous: Summary;
  activity: { date: string; reached: number; clicked: number }[];
  cohorts: { key: string; label: string; count: number }[];
  growth: {
    since: string | null;
    available: boolean;
    start: number | null;
    end: number | null;
    net: number | null;
    series: { date: string; active: number }[];
  };
  charts?: {
    volume: {
      date: string;
      campaigns: number;
      newsletters: number;
      campaignClicks: number;
      newsletterClicks: number;
    }[];
    devices: { device: string; count: number }[];
    clickHours: { day: number; hour: number; count: number }[];
    clickEvents: number;
  };
  warnings: string[];
};
export type BreakdownRow = {
  id: string;
  name: string;
  kind: string;
  active: number;
  suppressed: number;
  reached: number;
  clicked: number;
  accepted: number;
  clickedMessages: number;
  bounced: number;
  complaints: number;
  unsubscribes: number;
  new: number;
  clickRate: Rate;
  audienceClickRate: Rate;
};
export type AnalyticsPage<T> = {
  items: T[];
  totalCount: number;
  page: number;
  pageSize: number;
  asOf?: string;
};
export function formatRate(rate?: Rate) {
  return rate?.value == null ? "—" : `${(rate.value * 100).toFixed(1)}%`;
}
export function rateDifference(current: Rate, previous: Rate) {
  return current.value == null || previous.value == null
    ? "No comparable rate"
    : `${current.value >= previous.value ? "+" : ""}${((current.value - previous.value) * 100).toFixed(1)} pp`;
}

export type EmailTotals = {
  total: number;
  sent: number;
  queued: number;
  failed: number;
  drafts: number;
  unknown: number;
  opened: number;
  clicked: number;
  bounced: number;
  campaigns: number;
  other: number;
};
export type EmailReport = {
  metricVersion: "email-v1";
  from: string;
  to: string;
  asOf: string;
  timezone: string;
  summary: EmailTotals;
  openRate: Rate;
  clickRate: Rate;
  bounceRate: Rate;
  series: (EmailTotals & { date: string })[];
};
