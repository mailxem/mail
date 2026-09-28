/** Deterministic, synthetic marketing data. Never connected to an account or API. */
export const DEMO_END = "2026-09-10";
export const campaigns = [
  {
    id: "sunday",
    name: "The Sunday Edit",
    type: "newsletters",
    subject: "Good things for a slower Sunday",
    color: "#8370bf",
  },
  {
    id: "show-work",
    name: "Show Your Work",
    type: "newsletters",
    subject: "The long and short of it",
    color: "#ad94df",
  },
  {
    id: "most-wanted",
    name: "August’s Most Wanted",
    type: "campaigns",
    subject: "Your next favorite thing",
    color: "#d88c64",
  },
  {
    id: "product-notes",
    name: "Freshly Shipped",
    type: "campaigns",
    subject: "A little update. A big difference.",
    color: "#e5b684",
  },
  {
    id: "welcome",
    name: "A Warm Welcome",
    type: "automations",
    subject: "So glad you’re here",
    color: "#367b65",
  },
  {
    id: "next-step",
    name: "Your Next Chapter",
    type: "automations",
    subject: "Let’s keep the conversation going",
    color: "#74a491",
  },
] as const;
export type SendType = (typeof campaigns)[number]["type"];
export const palette = {
  newsletters: "#8d76c5",
  campaigns: "#e2a076",
  automations: "#397e69",
  opens: "#bca5e4",
  clicks: "#407f69",
  mobile: "#9078c4",
  desktop: "#43806a",
  other: "#e1a17a",
};
export interface DemoRecord {
  date: string;
  label: string;
  campaign: string;
  type: SendType;
  accepted: number;
  opens: number;
  clicks: number;
  rejected: number;
  deferred: number;
  subscribers: number;
  unsubscribed: number;
  mobile: number;
  desktop: number;
  other: number;
}
const fraction = (n: number) => {
  const x = Math.sin(n * 127.1 + 311.7) * 43758.5453123;
  return x - Math.floor(x);
};
export const records: DemoRecord[] = Array.from({ length: 180 }, (_, i) => {
  const date = new Date(Date.UTC(2026, 8, 10) - (179 - i) * 86400000);
  const iso = date.toISOString().slice(0, 10);
  const day = date.getUTCDay();
  return campaigns.map((campaign, j) => {
    const noise = fraction(i * 7 + j + 13);
    const base =
      campaign.type === "automations"
        ? 390
        : campaign.type === "newsletters"
          ? 980
          : 650;
    const burst =
      campaign.type === "newsletters"
        ? day === 0 || day === 3
          ? 3.1
          : 0.45
        : campaign.type === "campaigns"
          ? day === 2 || day === 5
            ? 3.8
            : 0.3
          : 1.2;
    const accepted = Math.round(
      base * burst * (0.8 + noise * 0.5) * (1 + i / 240),
    );
    const opens = Math.round(accepted * (0.36 + fraction(i + j * 53) * 0.19));
    const clicks = Math.round(
      accepted * (0.042 + fraction(i * 2 + j * 23) * 0.086),
    );
    const mobile = Math.round(clicks * (0.5 + noise * 0.15));
    const desktop = Math.round((clicks - mobile) * 0.84);
    return {
      date: iso,
      label: date.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        timeZone: "UTC",
      }),
      campaign: campaign.id,
      type: campaign.type,
      accepted,
      opens,
      clicks,
      rejected: Math.round(accepted * (0.004 + noise * 0.006)),
      deferred: Math.round(accepted * 0.002),
      subscribers: Math.round(accepted * (0.023 + noise * 0.01)),
      unsubscribed: Math.round(accepted * 0.0009),
      mobile,
      desktop,
      other: clicks - mobile - desktop,
    };
  });
}).flat();
export function getDemo(days: number, campaignId: string) {
  const cutoff = records[(180 - days) * campaigns.length].date;
  const selected = records.filter(
    (r) =>
      r.date >= cutoff && (campaignId === "all" || r.campaign === campaignId),
  );
  const daily = Array.from(new Set(selected.map((r) => r.date))).map((date) => {
    const rows = selected.filter((r) => r.date === date);
    const sum = (key: keyof DemoRecord) =>
      rows.reduce((n, r) => n + (Number(r[key]) || 0), 0);
    const byType = (type: SendType, key: "accepted" | "clicks") =>
      rows.filter((r) => r.type === type).reduce((n, r) => n + r[key], 0);
    return {
      date,
      label: rows[0].label,
      accepted: sum("accepted"),
      opens: sum("opens"),
      clicks: sum("clicks"),
      subscribers: sum("subscribers"),
      unsubscribed: sum("unsubscribed"),
      newsletters: byType("newsletters", "accepted"),
      campaigns: byType("campaigns", "accepted"),
      automations: byType("automations", "accepted"),
      newslettersClicks: byType("newsletters", "clicks"),
      campaignsClicks: byType("campaigns", "clicks"),
      automationsClicks: byType("automations", "clicks"),
    };
  });
  const sum = (key: keyof DemoRecord) =>
    selected.reduce((n, r) => n + (Number(r[key]) || 0), 0);
  const totals = {
    accepted: sum("accepted"),
    opens: sum("opens"),
    clicks: sum("clicks"),
    subscribers: sum("subscribers"),
    unsubscribed: sum("unsubscribed"),
    rejected: sum("rejected"),
    deferred: sum("deferred"),
  };
  const table = campaigns
    .filter((c) => campaignId === "all" || campaignId === c.id)
    .map((c) => {
      const rows = selected.filter((r) => r.campaign === c.id);
      const accepted = rows.reduce((n, r) => n + r.accepted, 0);
      const clicks = rows.reduce((n, r) => n + r.clicks, 0);
      return {
        ...c,
        accepted,
        clicks,
        rate: (clicks / accepted) * 100,
        opens: rows.reduce((n, r) => n + r.opens, 0),
      };
    });
  const devices = [
    {
      name: "Mobile",
      key: "mobile",
      value: sum("mobile"),
      fill: palette.mobile,
    },
    {
      name: "Desktop",
      key: "desktop",
      value: sum("desktop"),
      fill: palette.desktop,
    },
    { name: "Other", key: "other", value: sum("other"), fill: palette.other },
  ];
  const before = records.filter(
    (r) =>
      r.date < cutoff && (campaignId === "all" || r.campaign === campaignId),
  );
  let audience =
    18240 + before.reduce((n, r) => n + r.subscribers - r.unsubscribed, 0);
  const growth = daily.map((row) => {
    audience += row.subscribers - row.unsubscribed;
    return {
      label: row.label,
      total: audience,
      new: row.subscribers,
      lost: row.unsubscribed,
    };
  });
  const heatmap = Array.from({ length: 7 }, (_, day) =>
    Array.from({ length: 6 }, (_, hour) => {
      const sum = selected
        .filter(
          (r) => (new Date(r.date + "T12:00:00Z").getUTCDay() + 6) % 7 === day,
        )
        .reduce((n, r) => n + r.clicks, 0);
      const portions = [0.04, 0.12, 0.29, 0.26, 0.2].map((weight) =>
        Math.round(sum * weight),
      );
      return hour === 5
        ? sum - portions.reduce((n, count) => n + count, 0)
        : portions[hour];
    }),
  );
  const previousCutoff = records[(180 - days * 2) * campaigns.length].date;
  const previousRows = before.filter((r) => r.date >= previousCutoff);
  const previous = {
    accepted: previousRows.reduce((n, r) => n + r.accepted, 0),
    clicks: previousRows.reduce((n, r) => n + r.clicks, 0),
    subscribers: previousRows.reduce((n, r) => n + r.subscribers, 0),
  };
  return { daily, totals, table, devices, growth, heatmap, cutoff, previous };
}
export const number = (value: number) =>
  new Intl.NumberFormat("en-US").format(Math.round(value));
export const compact = (value: number) =>
  new Intl.NumberFormat("en-US", {
    notation: "compact",
    maximumFractionDigits: 1,
  }).format(value);
