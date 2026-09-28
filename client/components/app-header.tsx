"use client";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "./ui/button";
import { PageHeader } from "./page-header";
import { CreateKey } from "./api-keys/create-key";
const pageTitles: Record<string, { title: string; description?: string }> = {
  "/": { title: "Ask Xem", description: "Plan, write, and manage your email." },
  "/campaigns": {
    title: "Campaigns",
    description: "Plan your next send and keep track of every campaign.",
  },
  "/campaigns/new": {
    title: "New campaign",
    description: "Create a new email campaign",
  },
  "/templates": {
    title: "Templates",
    description: "Manage your email templates",
  },
  "/templates/new": {
    title: "New template",
    description: "Create a new email template",
  },
  "/audience": {
    title: "Audience",
    description: "Manage your contacts and lists",
  },
  "/audience/lists": {
    title: "Contact lists",
    description: "Manage your mailing lists",
  },
  "/audience/tags": {
    title: "Tags",
    description: "Organize your contacts with tags",
  },
  "/audience/dashboard": {
    title: "Audience overview",
    description: "View audience analytics",
  },
  "/analytics": {
    title: "Analytics",
    description: "View your campaign performance",
  },
  "/analytics/campaigns": { title: "Campaign analytics" },
  "/analytics/audience": { title: "Audience analytics" },
  "/analytics/team": { title: "Team analytics" },
  "/analytics/trends": { title: "Trends" },
  "/automations": {
    title: "Automations",
    description: "Set up automated workflows",
  },
  "/forms": { title: "Forms", description: "Create and manage forms" },
  "/inbox": { title: "Inbox", description: "Your email inbox" },
  "/settings": {
    title: "Settings",
    description: "Manage your account settings",
  },
  "/settings/billing": {
    title: "Billing",
    description: "Your plan, usage, and billing details.",
  },
  "/settings/api-keys": {
    title: "API keys",
    description: "Manage access for your applications and integrations.",
  },
  "/settings/smtp": {
    title: "SMTP senders",
    description: "Connect and manage the servers you send email through.",
  },
  "/settings/imap": {
    title: "IMAP mailboxes",
    description: "Connect the mailboxes you receive email in.",
  },
  "/settings/webhooks": { title: "Webhooks", description: "Manage webhooks" },
  "/team": { title: "Team", description: "Manage the people who share your workspace." },
  "/developer/logs/emails": {
    title: "Outbox",
    description: "View email delivery logs",
  },
  "/forms/new": { title: "New form", description: "Create a new form" },
};

function getPageTitle(pathname: string): {
  title: string;
  description?: string;
} {
  // Try exact match first
  if (pageTitles[pathname]) {
    return pageTitles[pathname];
  }

  // Try pattern matching for dynamic routes
  if (pathname.startsWith("/campaigns/") && pathname !== "/campaigns/new") {
    return { title: "Campaign details", description: "View and edit your campaign." };
  }
  if (pathname.startsWith("/templates/") && pathname.includes("/edit")) {
    return {
      title: "Edit template",
      description: "Customize your email template",
    };
  }
  if (pathname.startsWith("/audience/lists/")) {
    return { title: "List details", description: "Manage list contacts." };
  }
  if (pathname.startsWith("/settings/api-keys/")) {
    return { title: "API key details" };
  }
  if (pathname.startsWith("/automations/new/")) {
    return {
      title: "New automation",
      description: "Create automated workflow",
    };
  }

  // Default fallback
  return { title: "Dashboard" };
}


export function AppHeader() {
  const pathname = usePathname();
  const { title, description } = getPageTitle(pathname);
  const actions: Record<string, [string, string]> = {
    "/campaigns": ["Create campaign", "/campaigns/new"],
    "/settings/smtp": ["Add sender", "/settings/smtp?dialog=true"],
    "/settings/imap": ["Add mailbox", "/settings/imap?dialog=true"],
  };
  const action = actions[pathname];
  return (
    <PageHeader heading={title} description={description}>
      {action && <Button asChild><Link href={action[1]}><Plus size={15} strokeWidth={1.65} />{action[0]}</Link></Button>}
      {pathname === "/settings/api-keys" && <CreateKey />}
    </PageHeader>
  );
}
