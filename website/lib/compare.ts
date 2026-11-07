// Competitor facts for /compare and /migrate. Every price comes from the
// vendor's own pricing page or data unless the priceNote says otherwise.
// Re-check by hand before changing `checked`.
export const checked = "November 5, 2026";
/** List sizes in the comparison tables; prices[] follows this order. */
export const sizes = [10000, 50000, 100000] as const;

export type Competitor = {
  id: string;
  name: string;
  /** Short pricing model for meta descriptions, e.g. "charges per contact". */
  modelShort: string;
  model: string;
  plan: string;
  /** Monthly USD on monthly billing, or a short note where no verified number exists. */
  prices: [number | string, number | string, number | string];
  priceNote?: string;
  headline: string;
  summary: string;
  cheaperAnswer: string;
  chooseThem: string[];
  exportSteps: string;
  exportShort: string;
  exportHelp?: string;
  /** Xem's importer detects this vendor's export format. */
  recognised: boolean;
  extraRebuild?: string;
  /** Its MCP server, API and AI-agent support. */
  agents?: string;
  sources: string[];
};

export const competitors: Competitor[] = [
  {
    id: "mailchimp",
    name: "Mailchimp",
    modelShort: "charges per contact, with a monthly send cap",
    model:
      "Per contact in your audience, including unsubscribed and non-subscribed contacts, with a monthly send limit of 10× contacts on Essentials, 12× on Standard and 15× on Premium. Essentials stops at 50,000 contacts and Standard at 100,000.",
    plan: "Standard",
    prices: [135, "Quote on Mailchimp’s site", "Quote on Mailchimp’s site"],
    priceNote:
      "Mailchimp’s pricing page shows prices in local currency and loads US dollar figures with JavaScript, so we could not read them directly. The 10,000-contact figure is from third-party pricing summaries dated October 2026, as in our Mailchimp alternatives article. Send caps and contact limits are from Mailchimp’s own pages.",
    headline: "Pay for sends, not contacts.",
    summary:
      "Mailchimp bills by the size of your audience, so quiet and unsubscribed contacts still cost money. Xem bills by the emails you actually send, with campaigns, automations and forms on every plan.",
    cheaperAnswer:
      "For a weekly newsletter to 10,000 contacts, Mailchimp Standard is about $135 a month against $19 on Xem. Mailchimp counts every contact, including unsubscribed ones; Xem counts only emails sent.",
    chooseThem: [
      "You rely on Mailchimp’s ecommerce integrations or its built-in ad and landing-page tools.",
      "Your team already knows Mailchimp well and the bill is not a concern.",
      "You need a very large template marketplace or an agency partner network.",
    ],
    exportSteps:
      "In Mailchimp, open Audience, choose the audience you want, then Export audience → Export as CSV. Mailchimp gives you a ZIP with separate files for subscribed, unsubscribed and cleaned contacts; import the subscribed and unsubscribed files.",
    exportShort:
      "Yes. Export your audience as a CSV from Audience → Export audience.",
    exportHelp: "https://mailchimp.com/help/view-export-contacts/",
    recognised: true,
    sources: [
      "https://mailchimp.com/pricing/marketing/",
      "https://mailchimp.com/help/about-mailchimp-pricing-plans/",
      "https://mailchimp.com/help/view-export-contacts/",
    ],
  },
  {
    id: "substack",
    name: "Substack",
    modelShort: "is free to publish and takes 10% of paid subscriptions",
    model:
      "Free to publish, with no subscriber limit. If you charge for subscriptions, Substack keeps 10% of that revenue, and card processing fees are taken on top.",
    plan: "its free publishing plan",
    prices: [
      "$0 + 10% of paid revenue",
      "$0 + 10% of paid revenue",
      "$0 + 10% of paid revenue",
    ],
    priceNote:
      "Substack does not charge by list size. Its going-paid page says writers keep 90% of their revenue minus card fees. Xem does not take a share of your revenue.",
    headline: "Own your list, your domain and your revenue.",
    summary:
      "Substack is free until you earn, then takes a tenth of your subscription revenue and sends from its own domain and app. Xem sends from your domain, keeps your readers in your own workspace and never takes a cut.",
    cheaperAnswer:
      "For a free newsletter, Substack costs nothing and Xem costs $19 a month at 10,000 subscribers sending weekly. Once you charge for subscriptions, Substack’s 10% share usually outgrows that: on $2,000 a month of paid revenue it is $200.",
    chooseThem: [
      "You want Substack’s network, recommendations and app to help readers find you.",
      "You publish for free and do not want to manage a sending domain.",
      "You want comments, chat and podcasts in the same place as your newsletter.",
    ],
    exportSteps:
      "In Substack, open your publication’s Subscribers page and choose Export to download a CSV of your subscribers. Paid subscribers’ billing stays in your connected Stripe account.",
    exportShort:
      "Yes. Export your subscribers as a CSV from the Subscribers page.",
    exportHelp: "https://support.substack.com/hc/en-us/articles/6314498343700",
    recognised: true,
    extraRebuild:
      "Paid subscriptions: Substack bills through your Stripe account; plan how you will keep charging paid readers before you switch.",
    sources: [
      "https://substack.com/going-paid",
      "https://support.substack.com/hc/en-us/articles/6314498343700",
    ],
  },
  {
    id: "beehiiv",
    name: "beehiiv",
    modelShort: "charges per subscriber",
    model:
      "Per subscriber, with unlimited sends. The free Launch plan covers 2,500 subscribers with beehiiv branding; Lite stops at 100,000 subscribers and Pro, which removes branding, at 250,000.",
    plan: "Lite",
    prices: [139, 299, 469],
    headline: "The same newsletter, a fraction of the bill.",
    summary:
      "beehiiv is built for newsletter growth and charges by subscriber. Xem charges by email sent, so a weekly newsletter to 50,000 people is $59 a month instead of $299.",
    cheaperAnswer:
      "At 50,000 subscribers sending weekly, beehiiv Lite is $299 a month on monthly billing and Xem is $59. At 10,000 it is $139 against $19.",
    chooseThem: [
      "You want beehiiv’s ad network and referral programme to monetise and grow.",
      "You publish a web-first newsletter and want its hosted website builder.",
      "You send many times a day, where unlimited sends per subscriber work in your favour.",
    ],
    exportSteps:
      "In beehiiv, go to Audience → Subscribers and export your subscribers to a CSV file.",
    exportShort:
      "Yes. Export your subscribers as a CSV from Audience → Subscribers.",
    exportHelp: "https://www.beehiiv.com/support/article/12428610562327",
    recognised: true,
    sources: [
      "https://www.beehiiv.com/pricing",
      "https://www.beehiiv.com/support/article/12428610562327",
    ],
  },
  {
    id: "kit",
    name: "Kit",
    modelShort: "charges per subscriber",
    model:
      "Per subscriber, with no send limits. The free Newsletter plan covers up to 10,000 subscribers with Kit branding and no automations; Creator adds automations and sequences.",
    plan: "Creator",
    prices: [139, "Quote on Kit’s site", "Quote on Kit’s site"],
    priceNote:
      "Kit’s pricing page publishes only its 1,000-subscriber price ($33 a month billed yearly). The 10,000-subscriber figure is from third-party summaries dated October 2026, as in our Mailchimp alternatives article.",
    headline: "Creator tools, without the per-subscriber bill.",
    summary:
      "Kit is built for creators and charges by subscriber once you need automations. Xem includes automations, sequences and forms on every plan, including Free, and charges by email sent.",
    cheaperAnswer:
      "At 10,000 subscribers sending weekly, Kit Creator is about $139 a month and Xem is $19, with automations included. Kit’s free plan is cheaper if you need neither automations nor more than 10,000 subscribers.",
    chooseThem: [
      "You sell digital products and want Kit’s commerce and creator network.",
      "You have fewer than 10,000 subscribers and do not need automations, so Kit’s free plan is enough.",
      "You rely on Kit’s creator integrations and recommendations.",
    ],
    exportSteps:
      "In Kit, open Subscribers, select all, then use the actions menu to export. Kit emails you a download link that expires after five days.",
    exportShort:
      "Yes. Export subscribers from the Subscribers page; Kit emails you the file.",
    exportHelp:
      "https://help.kit.com/en/articles/2502489-how-to-export-subscribers-in-kit",
    recognised: true,
    sources: [
      "https://kit.com/pricing",
      "https://help.kit.com/en/articles/2502489-how-to-export-subscribers-in-kit",
    ],
  },
  {
    id: "mailerlite",
    name: "MailerLite",
    modelShort: "charges per subscriber, with a send cap",
    model:
      "Per subscriber. The free plan covers 250 subscribers and 2,500 emails a month; Comfort, the cheapest paid plan, allows 10× your tier in emails each month.",
    plan: "Comfort",
    prices: [89, 319, 439],
    headline: "Simple pricing that doesn’t grow with your list.",
    summary:
      "MailerLite is a well-priced, easy tool that charges by subscriber. Xem charges by email sent, so at 50,000 subscribers a weekly newsletter is $59 a month instead of $319.",
    cheaperAnswer:
      "At 10,000 subscribers sending weekly, MailerLite Comfort is $89 a month on monthly billing and Xem is $19. At 50,000 it is $319 against $59.",
    chooseThem: [
      "You want MailerLite’s website and landing-page builder in the same subscription.",
      "You sell through its built-in ecommerce integrations.",
      "You have fewer than 250 subscribers, where its free plan fits.",
    ],
    exportSteps:
      "In MailerLite, open Subscribers, select all, then choose Actions → Export to CSV.",
    exportShort: "Yes. Export subscribers as a CSV from Subscribers → Actions.",
    exportHelp: "https://www.mailerlite.com/help/how-to-export-subscribers",
    recognised: false,
    sources: [
      "https://www.mailerlite.com/pricing",
      "https://www.mailerlite.com/help/how-to-export-subscribers",
    ],
  },
  {
    id: "brevo",
    name: "Brevo",
    modelShort: "charges by emails sent, plus a plan fee",
    model:
      "By email volume rather than contacts: a plan fee (Starter from $9) plus a monthly email-credit tier, with up to 100,000 contacts included on the free plan, which sends 300 emails a day. Removing Brevo’s logo on Starter costs extra.",
    plan: "Starter",
    prices: [
      "From $9 + email credits",
      "From $9 + email credits",
      "From $9 + email credits",
    ],
    priceNote:
      "Brevo prices by emails sent, so a price per list size does not apply. Figures are from Brevo’s published pricing data of September 29, 2026; check its calculator for your volume.",
    headline: "Both charge for sends. Only one is open source.",
    summary:
      "Brevo, like Xem, prices by emails sent, and bundles SMS, WhatsApp and a sales CRM. Xem focuses on newsletters and campaigns, includes automations and forms on every plan, and can be self-hosted.",
    cheaperAnswer:
      "It depends on volume. Both charge by email: Xem is $19 a month for 40,000 emails, with branding-free sending and every feature included. Compare Brevo’s calculator for your exact volume and the add-ons you need.",
    chooseThem: [
      "You want SMS, WhatsApp and email in one tool.",
      "You need Brevo’s sales CRM and meetings features.",
      "You send under 300 emails a day and its free plan is enough.",
    ],
    exportSteps:
      "In Brevo, open CRM → Contacts, select all, choose More actions → Export, then download the file from your notifications.",
    exportShort: "Yes. Export contacts as a CSV from CRM → Contacts.",
    exportHelp: "https://help.brevo.com/hc/en-us/articles/5310065850642",
    recognised: false,
    sources: [
      "https://www.brevo.com/pricing/",
      "https://help.brevo.com/hc/en-us/articles/5310065850642",
    ],
  },
  {
    id: "ghost",
    name: "Ghost",
    modelShort: "Ghost(Pro) charges per member",
    model:
      "Ghost(Pro) hosting charges per member and takes no share of paid subscriptions. Starter is limited to 1,000 members without paid subscriptions, so most newsletters need Publisher.",
    plan: "Ghost(Pro) Publisher",
    prices: [105, 249, 329],
    headline: "Two open-source tools. One charges for members.",
    summary:
      "Ghost is an open-source publishing platform with newsletters built in; Ghost(Pro) bills by member. Xem is open-source email marketing, billed by email sent, with automations, forms and a CRM.",
    cheaperAnswer:
      "At 10,000 members sending weekly, Ghost(Pro) Publisher is $105 a month on monthly billing and Xem is $19. Both can be self-hosted for free.",
    chooseThem: [
      "You want a full publishing website and membership site in one product.",
      "Your newsletter is mainly your blog posts sent by email.",
      "You already self-host Ghost and are happy with its newsletter features.",
    ],
    exportSteps:
      "In Ghost, go to Settings → Advanced → Import/Export, export your members as a CSV. On Ghost(Pro) the export link is emailed to you.",
    exportShort:
      "Yes. Export members as a CSV from Settings → Advanced → Import/Export.",
    exportHelp: "https://ghost.org/help/exports/",
    recognised: true,
    sources: ["https://ghost.org/pricing/", "https://ghost.org/help/exports/"],
  },
  {
    id: "listmonk",
    name: "listmonk",
    modelShort: "is free, self-hosted software (AGPL-3.0)",
    model:
      "Free and open source under AGPL-3.0, run on your own server with PostgreSQL. You pay your server and your email provider; through Amazon SES that is about $0.10 per 1,000 emails. There is no hosted plan from the project.",
    plan: "self-hosted on Amazon SES (server not included)",
    prices: ["≈ $4 + server", "≈ $20 + server", "≈ $40 + server"],
    priceNote:
      "listmonk itself is free. These are Amazon SES sending costs at $0.10 per 1,000 emails, before your server. Xem self-hosted costs the same to send; the Xem column is Xem Cloud, where we run it for you.",
    headline: "Both open source. Different amounts of app.",
    summary:
      "listmonk is a fast, lean newsletter and mailing-list manager in a single Go binary, and it's excellent at that. Xem does more: visual automations, signup forms with a CRM, a shared inbox and importers, and it also comes as a hosted service if you'd rather not run a server.",
    cheaperAnswer:
      "Self-hosted, they cost about the same: both are free and you pay your email provider. If you don't want to run a server, Xem Cloud is $19 a month for a weekly newsletter to 10,000 people; listmonk has no hosted plan.",
    chooseThem: [
      "You only need lists and campaigns and want the smallest, simplest thing to run.",
      "You like a single binary and want to keep the moving parts to PostgreSQL.",
      "You prefer AGPL-3.0, or you already run listmonk happily.",
    ],
    exportSteps:
      "listmonk keeps subscribers in PostgreSQL. From psql, run \\copy (select email, name, status from subscribers) to 'subscribers.csv' csv header. Keep the status column: Xem skips anyone marked blocklisted or disabled.",
    exportShort:
      "Yes. Export email, name and status from listmonk's subscribers table as a CSV.",
    recognised: false,
    sources: [
      "https://github.com/knadh/listmonk",
      "https://listmonk.app/docs/",
    ],
  },
  {
    id: "plunk",
    name: "Plunk",
    modelShort: "charges $0.001 per email sent, with no monthly fee",
    model:
      "Pay as you go: $0.001 per email with no monthly fee and no commitment. The free plan covers 1,000 emails a month with Plunk branding on marketing emails. Contacts are unlimited on every plan, emails with an attachment count double, and inbound emails count too.",
    plan: "Pay as you go",
    prices: [40, 200, 400],
    priceNote:
      "Plunk’s pricing page and billing docs give $0.001 per email, so 40,000, 200,000 and 400,000 emails come to $40, $200 and $400 a month. There is no base fee. Plunk is also open source (AGPL-3.0), and self-hosting it is free apart from your own server and email provider.",
    headline: "Both open source. Only one has a flat monthly price.",
    summary:
      "Plunk is an open-source email platform for developers: transactional email, campaigns and workflows, billed at $0.001 per email. Xem is also open source, but it is built for newsletters and marketing, with plan prices that fall per email as you grow, plus forms, a CRM and a shared inbox.",
    cheaperAnswer:
      "For a weekly newsletter to 10,000 subscribers (40,000 emails), Plunk is $40 a month and Xem is $19. At 50,000 subscribers it is $200 against $59, and at 100,000 it is $400 against $99. Plunk has no minimum, so for a few thousand emails a month it can be cheaper.",
    chooseThem: [
      "You send mostly transactional email from your app and want a small API-first tool with an SMTP relay.",
      "You send very little, or you want to pay only per email with no plan fee.",
      "You want the leanest open-source option to self-host, with inbound email handling built in.",
    ],
    exportSteps:
      "Plunk’s docs describe importing contacts from CSV, but we found no dashboard export. Page through your contacts with the List contacts API (GET /contacts, up to 100 per page, using the cursor from each response) and save the email, subscribed status and custom fields as a CSV.",
    exportShort:
      "Partly. We found no CSV export in Plunk’s docs, but the List contacts API returns every contact.",
    exportHelp: "https://docs.useplunk.com/api-reference/contacts/listContacts",
    recognised: false,
    agents:
      "Plunk publishes an official MCP server, @plunk/mcp, with tools for contacts, sending, events, campaigns, segments and workflows. It can run read-only, sends need your confirmation, and billing stays in the dashboard.",
    sources: [
      "https://www.useplunk.com/pricing",
      "https://docs.useplunk.com/concepts/billing",
      "https://docs.useplunk.com/guides/mcp-server",
      "https://docs.useplunk.com/api-reference/contacts/listContacts",
      "https://github.com/useplunk/plunk",
    ],
  },
  {
    id: "sequenzy",
    name: "Sequenzy",
    modelShort: "charges by emails sent, with unlimited contacts",
    model:
      "By emails sent per month, in volume tiers, with unlimited contacts on every plan. The free plan covers 2,500 emails a month and adds a small Sequenzy footer. Paid plans start at 15,000 emails for $19 and step up through 30k, 60k, 120k, 300k, 600k and larger tiers. SMS is a separate $16 a month add-on.",
    plan: "Pro",
    prices: [49, 199, 399],
    priceNote:
      "Sequenzy’s pricing page lists tiers by emails a month: $49 for 60,000, $199 for 300,000 and $399 for 600,000. A weekly newsletter to 10,000, 50,000 and 100,000 subscribers is 40,000, 200,000 and 400,000 emails, which fall in those tiers. The $29 tier covers 30,000 emails a month. Prices are on monthly billing; yearly billing is cheaper.",
    headline: "Both bill by emails sent. Xem costs less.",
    summary:
      "Sequenzy is a newer email platform for SaaS and ecommerce that prices by emails sent, not contacts, and was built with AI agents in mind. Xem uses the same pricing idea, adds open-source code you can self-host, and costs less at newsletter volumes.",
    cheaperAnswer:
      "For a weekly newsletter, Sequenzy Pro is $49 a month at 10,000 subscribers against $19 on Xem. At 50,000 it is $199 against $59, and at 100,000 it is $399 against $99. Both count emails sent and include unlimited contacts.",
    chooseThem: [
      "You run a SaaS product and want its Stripe-based revenue attribution and lifecycle flows.",
      "You want a hosted AI assistant connection that signs in with OAuth instead of an API key, or a ChatGPT plugin.",
      "You send under 2,500 emails a month and its free plan, with a footer, is enough.",
    ],
    exportSteps:
      "Sequenzy’s docs describe the List Subscribers API endpoint and daily subscriber snapshots to your own S3 or Google Cloud Storage bucket. We could not find a documented one-click CSV export, so check the Subscribers page in your dashboard, or pull your contacts through the API into a CSV.",
    exportShort:
      "Probably. Pull subscribers through the API or a data export; check your dashboard for a CSV option.",
    exportHelp: "https://docs.sequenzy.com/integrations/data-exports.md",
    recognised: false,
    agents:
      "Sequenzy has an official MCP server: a hosted remote server that signs in with OAuth, a local @sequenzy/mcp package that uses an API key, and a ChatGPT plugin. Its REST API covers subscribers, lists, segments, sequences, campaigns, templates, transactional email and sending domains.",
    sources: [
      "https://www.sequenzy.com/pricing",
      "https://docs.sequenzy.com/concepts/mcp.md",
      "https://docs.sequenzy.com/api-reference/introduction.md",
      "https://docs.sequenzy.com/integrations/data-exports.md",
    ],
  },
  {
    id: "bento",
    name: "Bento",
    modelShort: "charges per active user, with unlimited sends",
    model:
      "Per active user, with unlimited marketing sends on paid plans, subject to fair use. One plan from $29 a month covers up to 5,000 active users; above that it adds $0.01 per user up to 50,000, then $0.009 up to 250,000. An active user is a subscribed person or anyone with an event in the last 30 days.",
    plan: "Marketing Platform",
    prices: [79, 479, 929],
    priceNote:
      "Bento publishes one Marketing Platform plan: $29 a month for up to 5,000 active users, plus $0.01 per user from 5,001 to 50,000 and $0.009 from 50,001 to 250,000. Its own sample bills ($229 at 25,000 users, $929 at 100,000) match that formula, so the 10,000 and 50,000 figures are calculated from the same rates, counting every subscriber as an active user. The Chat add-on is $30 a month extra and not included.",
    headline: "Both send without a cap. Only one prices by the email.",
    summary:
      "Bento is a polished email platform for ecommerce and SaaS with unlimited sends and a strong API. It charges by active user. Xem charges by email sent, so a weekly newsletter to 50,000 people is $59 a month instead of $479.",
    cheaperAnswer:
      "At 10,000 subscribers sending weekly, Bento is $79 a month on monthly billing and Xem is $19. At 50,000 it is $479 against $59. Bento’s unlimited sends only pay off if you send far more than weekly.",
    chooseThem: [
      "You send many emails a day to a small list, where unlimited sends and per-user pricing work in your favour.",
      "You want Bento’s Chat add-on for a shared inbox with SMS, WhatsApp and push in one place.",
      "You want a hosted product with official SDKs for ten languages.",
    ],
    exportSteps:
      "Bento’s docs cover importing but we found no page on exporting. Check the People section of your Bento account for an export option, or ask Bento support. Include unsubscribed people and keep tags and custom fields as extra columns.",
    exportShort:
      "Probably. Bento documents imports but not a CSV export, so check People in your account or ask support.",
    recognised: false,
    agents:
      "Bento runs an official hosted MCP server, authenticated with API headers, covering subscribers, tags, events, broadcasts, sequences, workflows, templates and stats. Broadcasts are saved as unapproved drafts by default.",
    sources: [
      "https://bentonow.com/pricing",
      "https://bentonow.com/docs/integrations/mcp",
      "https://bentonow.com/docs",
    ],
  },
  {
    id: "sendy",
    name: "Sendy",
    modelShort:
      "is a one-time licence, self-hosted app that sends through Amazon SES",
    model:
      "A one-time licence ($69) for software you run on your own server with PHP and MySQL, sending through your own Amazon SES account. There is no monthly fee and no per-subscriber charge. Updates are free until the next major version, which is a paid upgrade. Each licence covers one domain.",
    plan: "self-hosted on Amazon SES (licence and server not included)",
    prices: [
      "≈ $4 + licence + server",
      "≈ $20 + licence + server",
      "≈ $40 + licence + server",
    ],
    priceNote:
      "Sendy is a one-time $69 licence, with free updates until the next major version and a paid upgrade after that. These figures are Amazon SES sending costs at $0.10 per 1,000 emails, before the licence and your server. Xem self-hosted is free and costs the same to send; the Xem column is Xem Cloud, where we run it for you.",
    headline: "Sendy is $69 once. Xem is $9 once.",
    summary:
      "Sendy is a proven, no-frills way to send newsletters through your own Amazon SES account, with a one-time licence instead of a monthly bill. Xem is also self-hostable and open source, with newsletters on a $9 one-time license, and adds visual automations, signup forms with a CRM and a shared inbox, and comes as a hosted service if you’d rather not run a server.",
    cheaperAnswer:
      "Self-hosted, Xem is cheaper: Sendy is $69 once, and Xem's license, which adds newsletters and the shared inbox, is $9 once. Both cost the same to send through SES. If you don’t want to run a server, Xem Cloud is $19 a month for a weekly newsletter to 10,000 people; Sendy has no hosted plan.",
    chooseThem: [
      "You want a simple PHP and MySQL newsletter sender that you pay for once and rarely touch.",
      "You manage several clients or brands and need Sendy’s white-labelled client accounts with sending limits.",
      "You already run Sendy happily and don’t need automations beyond autoresponders.",
    ],
    exportSteps:
      "In Sendy, open View all lists and choose your list. Pick the status tab you want (active, unsubscribed, bounced) and click Export to download a CSV. Export the active and unsubscribed subscribers separately so Xem keeps unsubscribed people unsubscribed.",
    exportShort:
      "Yes. Export each list as a CSV from View all lists, one status tab at a time.",
    recognised: false,
    agents:
      "Sendy has a documented HTTP API for subscribing, unsubscribing, listing lists and creating or sending a campaign. It has no official MCP server, so agents reach it through third-party connectors.",
    sources: [
      "https://sendy.co",
      "https://sendy.co/api",
      "https://sendy.co/get-updated",
    ],
  },
  {
    id: "mailcheap",
    name: "MailCheap",
    modelShort: "charges by emails sent, in volume tiers",
    model:
      "By emails sent per month, not by subscribers. Plans step from 20,000 emails for $19 to 75,000 for $39, 200,000 for $79, 500,000 for $149, 2,000,000 for $349 and 5,000,000 for $799. Every plan has the same features. There is no free plan, and when you pass the cap new sends wait until you upgrade or the month resets.",
    plan: "the smallest tier that fits your monthly emails",
    prices: [39, 79, 149],
    priceNote:
      "From mailcheap.ai’s pricing page, on monthly billing, for 40,000, 200,000 and 400,000 emails a month. 40,000 needs the 75,000-email tier ($39), 200,000 fits the $79 tier, and 400,000 needs the 500,000-email tier ($149). At 25,000 emails a month MailCheap is $39 and Xem is $11.50. Yearly billing is two months free.",
    headline: "Same billing idea. Different price.",
    summary:
      "MailCheap is an open-source newsletter app that, like Xem, charges for emails sent rather than subscribers, and it has an MCP server and API. Xem costs less at each of these volumes and also includes automations, a CRM and a shared inbox.",
    cheaperAnswer:
      "For a weekly newsletter to 10,000 subscribers, MailCheap is $39 a month and Xem is $19. At 50,000 it is $79 against $59, and at 100,000 it is $149 against $99. Both are open source and can be self-hosted for the cost of your email provider.",
    chooseThem: [
      "You want an MCP server that covers every API endpoint, including billing and team, with OAuth sign-in for ChatGPT and Claude.",
      "You want a public website for your issues as part of the product, with a simple newsletter-only feature set.",
      "You like its safeguards: staged first sends, and campaigns that pause on high bounce or spam rates.",
    ],
    exportSteps:
      "MailCheap exports contacts as a CSV download link, valid for an hour, through its API (GET /api/v1/contacts/export, or the export_contacts MCP tool) with a full-access API key. Keep each contact’s status in the file.",
    exportShort:
      "Yes. Export contacts as a CSV through MailCheap’s API or MCP.",
    exportHelp: "https://mailcheap.ai/docs/api",
    recognised: false,
    agents:
      "MailCheap ships an official MCP server with 90 tools, one per REST endpoint, covering campaigns, contacts, segments, domains, newsletters, team, API keys and billing. ChatGPT, Claude and Grok connect with OAuth; Claude Code, Codex and Cursor use an API key.",
    sources: [
      "https://mailcheap.ai/",
      "https://mailcheap.ai/docs",
      "https://mailcheap.ai/docs/mcp",
      "https://github.com/marclou/mailcheap",
    ],
  },
];

/** Xem's side of the AI-agent comparison. */
export const xemAgents =
  "Xem runs a hosted MCP server at mcp.xem.email with 43 tools for campaigns, newsletters, contacts, CSV imports, forms, CRM notes, automations, sending domains and analytics, or you can run @xem.email/mcp locally. It works with Claude, Cursor, Codex and any MCP client. Campaigns start as drafts, imports preview first, and sending is a separate tool.";

export const getCompetitor = (id: string) =>
  competitors.find((c) => c.id === id);
