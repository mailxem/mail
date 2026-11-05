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
      "Mailchimp bills by the size of your audience, so quiet and unsubscribed contacts still cost money. Xem bills by the emails you actually send, and every plan includes the whole product.",
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
      "Brevo, like Xem, prices by emails sent, and bundles SMS, WhatsApp and a sales CRM. Xem focuses on newsletters and campaigns, includes every feature on every plan, and can be self-hosted.",
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
];

export const getCompetitor = (id: string) =>
  competitors.find((c) => c.id === id);
