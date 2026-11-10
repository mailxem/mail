// What's new in Xem, newest first. Each entry is something people can use
// today, written from the merged pull request it links to. Add new entries at
// the top when a change is live, not when it merges.
export type Change = {
  date: string; // YYYY-MM-DD
  title: string;
  body: string;
  pr: number;
  href?: string;
};

export const changes: Change[] = [
  {
    date: "2026-11-08",
    title: "A click map for every campaign",
    body: "Open a sent campaign to see the email as it went out, with each link coloured by how many people clicked it.",
    pr: 104,
  },
  {
    date: "2026-11-08",
    title: "Overview counts every email",
    body: "Sent, Opened and Clicked on the Overview now match the Email tab, including emails sent through the API.",
    pr: 103,
  },
  {
    date: "2026-11-08",
    title: "Analytics and a click heatmap for newsletter websites",
    body: "See views, visitors and where people click on your newsletter's website, without adding a script.",
    pr: 101,
  },
  {
    date: "2026-11-08",
    title: "Research: email authentication across the top 10,000 websites",
    body: "We checked SPF, DKIM and DMARC for the Tranco top 10,000 domains and published the results and the method.",
    pr: 102,
    href: "/research/email-authentication",
  },
  {
    date: "2026-11-07",
    title: "Admin API keys, with expiry you choose",
    body: "Create API keys with full workspace access, and pick no expiry, 7 days, 30 days, 90 days or a year.",
    pr: 98,
  },
  {
    date: "2026-11-07",
    title: "One-click DNS for GoDaddy, Cloudflare and Namecheap",
    body: "Add your sending records with one click at your DNS provider, and see on the dashboard when sending is ready.",
    pr: 97,
  },
  {
    date: "2026-11-07",
    title: "Ask Xem can do things, not only answer",
    body: "Ask Xem drafts campaigns, finds contacts and reads your results, and changes nothing until you approve.",
    pr: 96,
  },
  {
    date: "2026-11-07",
    title: "Campaign results on the campaign page",
    body: "Each campaign's page shows how it did, and a campaign that hasn't gone out yet is labelled Scheduled.",
    pr: 95,
  },
  {
    date: "2026-11-07",
    title: "Newsletter websites",
    body: "Publish every issue on a website with 14 templates, email-confirmed signups and your own domain.",
    pr: 92,
  },
  {
    date: "2026-11-06",
    title: "Edit live automations safely",
    body: "Changes to a running automation are saved as a draft and go live when you publish, with results for each step.",
    pr: 68,
  },
  {
    date: "2026-11-06",
    title: "Automations on opens and clicks, and webhooks",
    body: "Start an automation when someone opens or clicks, and get a webhook for every email event.",
    pr: 61,
  },
  {
    date: "2026-11-06",
    title: "List cleanup",
    body: "Readers who ignore your emails for a long time get one question. If they don't answer, Xem stops emailing them.",
    pr: 59,
  },
  {
    date: "2026-11-06",
    title: "Opens and clicks from people, not link scanners",
    body: "Security scanners that open and click every link are left out of your stats.",
    pr: 58,
  },
  {
    date: "2026-11-06",
    title: "Bot protection for signup forms",
    body: "Public signup forms turn away bots before they reach your list.",
    pr: 57,
  },
  {
    date: "2026-11-06",
    title: "A new template library and an AI email designer",
    body: "Start from a hand-built template, or describe the email you want and edit the design Xem makes.",
    pr: 50,
    href: "/templates",
  },
  {
    date: "2026-11-06",
    title: "Self-host Xem with one command",
    body: "Run the whole of Xem on your own server with a single docker compose up.",
    pr: 49,
  },
  {
    date: "2026-11-06",
    title: "A free email deliverability checker",
    body: "Check any domain's MX, SPF, DKIM and DMARC records and get the exact record to fix each problem.",
    pr: 46,
    href: "/tools/email-checker",
  },
  {
    date: "2026-11-05",
    title: "Guides for switching to Xem",
    body: "Step-by-step pages for moving from Mailchimp, beehiiv, Substack, Kit and others, with honest price comparisons.",
    pr: 43,
    href: "/compare",
  },
  {
    date: "2026-11-05",
    title: "Guided onboarding and a new dashboard",
    body: "A checklist takes you from choosing how you send to your first campaign, on a redesigned dashboard.",
    pr: 42,
  },
  {
    date: "2026-10-30",
    title: "First campaigns go out in stages",
    body: "A list's first campaign goes to a small group first, and the rest follows only if bounces stay low.",
    pr: 26,
  },
  {
    date: "2026-10-30",
    title: "Pricing per email sent",
    body: "Xem Cloud charges for the emails you send, not your list size. 3,000 emails a month are free.",
    pr: 25,
    href: "/pricing",
  },
];
