export const productViews = [
  {
    id: "design",
    label: "Design",
    title: "Make it unmistakably you.",
    description:
      "Start with an editable template. Shape every block, image, and word in the visual editor — with a little help from AI when you need it.",
    image: "editor-figma",
    alt: "The actual Xem editor displaying Show Your Work from the first 15 templates, with green editorial type, illustrations, and native design blocks.",
    path: "/templates",
    bullets: [
      "Editable layouts and reusable templates",
      "AI assistance for complete email designs",
      "Desktop and mobile previews",
    ],
  },
  {
    id: "send",
    label: "Send",
    title: "Give good ideas a rhythm.",
    description:
      "Connect your newsletter to a template and an audience. Set the cadence and timezone, then keep every edition in the same place.",
    image: "newsletters",
    alt: "Xem newsletter workspace showing The Sunday Edit and Behind the Build with their schedules.",
    path: "/newsletters",
    bullets: [
      "Reusable newsletter templates",
      "Scheduled editions and timezones",
      "Campaigns for your one-off announcements",
    ],
  },
  {
    id: "automate",
    label: "Automate",
    title: "The right moment. Already planned.",
    description:
      "Build a welcome, a follow-up, or a longer conversation. Connect triggers, emails, delays, and conditions on a visual canvas.",
    image: "automations",
    alt: "Xem visual automation builder showing a contact trigger, email step, wait step, and exit.",
    path: "/automations",
    bullets: [
      "Visual workflow builder",
      "Contact and email event triggers",
      "Reusable email templates in your journeys",
    ],
  },
  {
    id: "understand",
    label: "Understand",
    title: "See the people behind the numbers.",
    description:
      "Explore audience activity, compare campaigns, and understand delivery health. Follow the signals back to the contacts who make them.",
    image: "analytics",
    alt: "Xem campaign analytics with date filters, accepted message totals, click rates, and audience activity chart using sample data.",
    path: "/analytics",
    bullets: [
      "Date, list, and tag filters",
      "Campaign and newsletter comparisons",
      "Audience cohorts and CRM drilldowns",
    ],
  },
] as const;
export { default as templates } from "./templates.json";
// The questions that decide whether someone signs up; the rest live in the docs.
export const faqs = [
  {
    q: "Can I use my own email provider?",
    a: "Yes. Connect your SMTP sender for outgoing email and an IMAP mailbox for your inbox. Your sender must be configured and verified in your workspace before you send.",
  },
  {
    q: "Can I bring my list from Mailchimp or beehiiv?",
    a: "Yes. Export a CSV from Mailchimp, beehiiv, Substack, Kit or most other tools and import it into a list. Unsubscribed contacts stay unsubscribed, and there's a switching guide for each of them.",
  },
  {
    q: "How does Xem protect deliverability?",
    a: "Managed sending goes through Amazon SES from your own verified domain, with DKIM, SPF and DMARC checks before you send. Hard bounces and spam complaints are suppressed automatically, and marketing emails carry a one-click unsubscribe header. If hard bounces reach 5% or complaints 0.3% of a week’s sends, Xem pauses sending before it hurts your domain. Readers who ignore five sends over 60 days get one question, and are set aside as inactive if they don’t answer within 14 days.",
  },
  {
    q: "Can Claude or another AI agent run Xem?",
    a: "Yes. The Xem MCP server gives Claude, Cursor, Codex and any MCP client 43 tools for campaigns, newsletters, contacts, CSV imports, forms, automations, sending domains and analytics. Connect to the hosted server at mcp.xem.email with a workspace API key, or run @xem.email/mcp locally. Campaigns start as drafts and sending is a separate tool, so a person can approve every send.",
  },
  {
    q: "What do I get on the free plan?",
    a: "3,000 emails a month, up to 100 a day, with no card needed. Campaigns, automations, forms, the CRM, analytics, templates and the API are all included. Free emails end with a small “Sent with Xem” link; paid plans start at $9 and remove it.",
  },
];
