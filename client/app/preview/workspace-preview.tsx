"use client";
import { DashboardPreview } from "./dashboard-preview";
import { AssistantPreview } from "@/components/assistant/assistant-preview";
import { SendingPreview } from "./sending-preview";
import { workspaceClassName } from "@/lib/workspace-styles";
import { useState } from "react";
import { AppShell } from "@/components/app-shell";
import { PreviewTransport, type Transport } from "@/lib/marketing/api";
import { FormsPage } from "@/components/marketing/forms";
import { NewslettersPage } from "@/components/marketing/newsletters";
import { TemplatesPage } from "@/components/marketing/templates";
import { CRMPage } from "@/components/marketing/crm";
import { AutomationsPage } from "@/components/marketing/automations";
import { InboxPage, OutboxPage } from "@/components/marketing/inbox";
const stamp = "2026-09-09T08:45:00Z";
const emailHTML =
  '<html><body style="font:15px/1.8 Arial;color:#5d5664;padding:25px"><p>Hi team,</p><p>I’ve gathered our latest product updates and a few ideas for the next edition. The focus is on making the experience feel simpler, more thoughtful, and easier to use.</p><p>Take a look and let me know what you think. I’d love to hear your feedback before we share it with the community.</p><p>Thanks,<br>Ava</p></body></html>';
const formFields = [
  {
    Label: "Email address",
    FieldType: "EMAIL",
    Required: true,
    mapToContactField: "email",
  },
  {
    Label: "First name",
    FieldType: "TEXT",
    Required: false,
    mapToContactField: "first_name",
  },
];
const forms = [
  ["User Feedback", "User experience feedback form", 879, 3420, "PUBLISHED"],
  [
    "Product Survey",
    "Monthly product satisfaction survey",
    245,
    1992,
    "ARCHIVED",
  ],
  [
    "Event Registration",
    "Upcoming webinars registration",
    532,
    1767,
    "ARCHIVED",
  ],
  [
    "Lead Generation",
    "Contact form for potential clients",
    1015,
    4452,
    "PUBLISHED",
  ],
].map((x, i) => ({
  id: `form-${i}`,
  Name: x[0],
  description: x[1],
  SubmissionCount: x[2],
  ViewCount: x[3],
  Status: x[4],
  Slug: `preview-${i}`,
  fields: formFields,
  AddToListID: "audience",
  successMessage: "Thanks for joining us!",
  SubmitButtonText: "Subscribe",
  updatedAt: stamp,
}));
const options = {
  lists: [
    { id: "audience", name: "The Xem community", subscribersCount: 2486 },
  ],
  templates: [
    {
      id: "template",
      name: "The weekly edit",
      subject: "A little inspiration for your week",
      htmlBody: emailHTML,
    },
  ],
  senders: [
    {
      id: "sender",
      name: "Xem",
      fromEmail: "hello@example.com",
      supportsTLS: true,
    },
  ],
  starters: [
    {
      key: "editorial",
      name: "The weekly edit",
      description: "A considered collection of stories, ideas and links.",
      subject: "Your weekly dose of inspiration",
      color: "#7a61c7",
      htmlBody: emailHTML,
    },
    {
      key: "product",
      name: "Product notes",
      description: "Keep your community close to what you are building.",
      subject: "A little update. A big difference.",
      color: "#408d87",
      htmlBody: emailHTML,
    },
    {
      key: "digest",
      name: "Community digest",
      description: "Member stories, upcoming events and everything in between.",
      subject: "Good things are happening here",
      color: "#b9825b",
      htmlBody: emailHTML,
    },
    {
      key: "launch",
      name: "Something new",
      description: "Make your next announcement one worth opening.",
      subject: "Meet your new favorite thing",
      color: "#4f5363",
      htmlBody: emailHTML,
    },
  ],
};
const contacts = [
  ["Ava", "Thompson", "Studio North", "QUALIFIED"],
  ["Ethan", "Carter", "BrightCo", "CUSTOMER"],
  ["Sophia", "Lee", "Luma Design", "LEAD"],
  ["Oliver", "James", "Nissan", "QUALIFIED"],
  ["Helena", "Ross", "Acme", "LEAD"],
  ["Liam", "Wilson", "Linear", "CUSTOMER"],
].map((a, i) => ({
  id: `contact-${i}`,
  firstName: a[0],
  lastName: a[1],
  company: a[2],
  lifecycleStage: a[3],
  email: `${a[0].toLowerCase()}@example.com`,
  status: "ACTIVE",
  listId: "audience",
  phone: "",
  createdAt: stamp,
}));
const newsletters = [
  {
    id: "newsletter",
    name: "The Sunday Edit",
    description: "Ideas for a slower Sunday.",
    subject: "Good things, delivered weekly",
    templateId: "template",
    listId: "audience",
    smtpConfigId: "sender",
    status: "SCHEDULED",
    cadence: "WEEKLY",
    timezone: "Asia/Kolkata",
    nextSendAt: "2026-09-13T04:30:00Z",
    editions: 12,
    postalAddress: "Xem, 42 Market Street, Bengaluru, India",
  },
  {
    id: "draft",
    name: "Behind the build",
    subject: "A little update. A big difference.",
    templateId: "template",
    listId: "audience",
    smtpConfigId: "sender",
    status: "DRAFT",
    cadence: "MONTHLY",
    timezone: "Asia/Kolkata",
    nextSendAt: null,
    editions: 0,
  },
];
const mail = contacts.map((c, i) => ({
  messageId: `message-${i}`,
  subject: [
    "Re: Sitemap refinements",
    "A few ideas for Sunday’s edition",
    "Your next chapter starts here",
    "Let’s talk about the launch",
    "Community notes · September",
    "Something worth sharing",
  ][i],
  from: `${c.firstName} ${c.lastName} <${c.email}>`,
  to: "Alex <alex@example.com>",
  cc: "",
  body: emailHTML,
  date: stamp,
  flags: i === 0 ? ["\\Flagged"] : [],
  attachments: [],
}));
const transport: Transport = async <T,>(
  path: string,
  method = "GET",
  body?: unknown,
) => {
  if (method !== "GET")
    throw new Error(
      "Visual preview only. Changes and email sending are available in your authenticated workspace.",
    );
  if (path.startsWith("marketing/contacts?")) {
    const params = new URLSearchParams(path.split("?")[1]);
    const search = (params.get("search") || "").toLowerCase();
    const stage = params.get("stage");
    const limit = Number(params.get("limit")) || 10;
    const filtered = contacts.filter(
      (c) =>
        (!stage || (c.lifecycleStage || "LEAD") === stage) &&
        [c.firstName, c.lastName, c.email, c.company]
          .join(" ")
          .toLowerCase()
          .includes(search),
    );
    const page = Math.min(
      Number(params.get("page")) || 1,
      Math.max(1, Math.ceil(filtered.length / limit)),
    );
    return {
      data: filtered.slice((page - 1) * limit, page * limit).map((c) => ({
        ...c,
        listName: options.lists.find((l) => l.id === c.listId)?.name || "",
      })),
      total: filtered.length,
      page,
      limit,
      summary: {
        total: contacts.length,
        qualified: contacts.filter((c) => c.lifecycleStage === "QUALIFIED")
          .length,
        customers: contacts.filter((c) => c.lifecycleStage === "CUSTOMER")
          .length,
        subscribed: contacts.filter((c) => c.status === "ACTIVE").length,
      },
    } as T;
  }
  const result =
    path === "marketing/options"
      ? options
      : path === "marketing/forms"
        ? forms
        : path === "marketing/newsletters"
          ? newsletters
          : path.startsWith("contacts?")
            ? { data: contacts }
            : path === "automations"
              ? []
              : path === "imap/folders"
                ? {
                    folders: [
                      { Name: "INBOX", Total: 6 },
                      { Name: "Sent", Total: 24 },
                      { Name: "Drafts", Total: 2 },
                      { Name: "Archive", Total: 18 },
                    ],
                  }
                : path.startsWith("emails?")
                  ? {
                      data: mail.map((m, i) => ({
                        ...m,
                        id: `preview-${i}`,
                        createdAt: m.date,
                        status: "SENT",
                        body: btoa(unescape(encodeURIComponent(m.body))),
                      })),
                      total: mail.length,
                      page: 1,
                    }
                  : path.startsWith("imap/emails")
                    ? { emails: mail, total_emails: 6, offset: 0, limit: 20 }
                    : [];
  return result as T;
};
export function WorkspacePreview() {
  const [page, setPage] = useState("/dashboard");
  return (
    <PreviewTransport.Provider value={transport}>
      <div className={workspaceClassName("preview-banner !h-[26px] whitespace-nowrap overflow-hidden !text-[9px]")}>
        LOCAL PREVIEW · Sample data · Sending disabled
      </div>
      <div
        className="[&_.product-frame]:!h-[calc(100dvh-26px)]"
        onClickCapture={(event) => {
          const anchor = (event.target as HTMLElement).closest("a");
          const href = anchor?.getAttribute("href");
          if (href === "/onboarding" || href === "/settings/sending") {
            event.preventDefault();
            event.stopPropagation();
            setPage(href);
          }
        }}
      >
        <AppShell previewPage={page} onPreviewNavigate={setPage}>
          {page === "/dashboard" || page.startsWith("/analytics") ? (
            <DashboardPreview onNavigate={setPage} />
          ) : page === "/" ? (
            <AssistantPreview />
          ) : page === "/onboarding" ? (
            <SendingPreview />
          ) : page === "/settings/sending" ? (
            <SendingPreview dashboard />
          ) : page === "/developer/logs/emails" ? (
            <OutboxPage />
          ) : page === "/newsletters" ? (
            <NewslettersPage />
          ) : page === "/templates" ? (
            <TemplatesPage />
          ) : page === "/crm" ? (
            <CRMPage />
          ) : page === "/automations" ? (
            <AutomationsPage />
          ) : page === "/inbox" ? (
            <InboxPage />
          ) : page === "/forms" ? (
            <FormsPage />
          ) : (
            <div className="m-6 rounded-xl border bg-card p-8">
              <h2 className="text-xl font-semibold">Open your workspace</h2>
              <p className="mt-2 text-muted-foreground">
                This page uses your account data and is available in the
                authenticated workspace.
              </p>
              <a
                className="mt-5 inline-flex text-primary underline"
                href={page}
              >
                Open {page}
              </a>
            </div>
          )}
        </AppShell>
      </div>
    </PreviewTransport.Provider>
  );
}
