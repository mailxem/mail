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
export const faqs = [
  {
    q: "What can I do with Xem?",
    a: "Create and send campaigns, publish recurring newsletters, design reusable email templates, and build visual automations. Contact lists, tags, lead forms, CRM, and analytics live alongside your email tools.",
  },
  {
    q: "Can I use my own email provider?",
    a: "Yes. Connect your SMTP sender for outgoing email and an IMAP mailbox for your inbox. Your sender must be configured and verified in your workspace before you send.",
  },
  {
    q: "Are the templates actually editable?",
    a: "Yes. The library contains over 200 templates made from native editor blocks. Change the copy, images, colors, and structure, then reuse your saved design in campaigns, newsletters, and automations.",
  },
  {
    q: "What does the AI writing tool create?",
    a: "In the template editor, AI creates a structured, editable email design. In the outbox, it helps write a rich-text email. You can review and edit the result before using it.",
  },
  {
    q: "Can I schedule a recurring newsletter?",
    a: "Yes. Connect a newsletter to an audience, sender, and saved template, choose your cadence and timezone, and manage its editions from the newsletter workspace.",
  },
  {
    q: "How should I read the analytics?",
    a: "Xem separates messages accepted by your SMTP server from audience engagement. Opens are directional because privacy tools can inflate them. Use clicks, audience cohorts, and delivery health together to understand what happened.",
  },
  {
    q: "Can I connect Xem to my own product?",
    a: "Use API keys and webhooks to connect your product to Xem. Manage your integrations in Settings & workspace, alongside your sender connections and team.",
  },
  {
    q: "Can Xem handle email delivery for me?",
    a: "Yes. Xem supports managed sending through Amazon SES, with verified domains, workspace approval, delivery history, and revocable SMTP credentials. Complete the domain checks and approval steps in your workspace before sending. You can also connect your own SMTP provider.",
  },
  {
    q: "Will managed sending replace my inbox?",
    a: "No. Managed sending handles outgoing email. Keep your existing mailbox and receiving MX records, and use your mailbox address for replies. The domain setup uses separate records for email authentication and a dedicated bounce subdomain.",
  },
  {
    q: "How can I help improve the new setup?",
    a: "Xem is open source. DNS provider guides, clearer onboarding copy, accessibility improvements, and reproducible SMTP integration examples are useful places to contribute. Explore the repositories on GitHub and discuss your idea before starting a larger change. Use sample data and never include credentials in an issue.",
  },
];
