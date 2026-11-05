"use client";
import Link from "next/link";
import { useResourcePage } from "@/hooks/use-resource-page";
import { useMarketingQuery } from "@/lib/marketing/api";
import type { LeadForm } from "@/lib/marketing/types";
import styles from "./chat.module.css";

function useGlance() {
  const contacts = useResourcePage("contacts", 1, 1);
  const campaigns = useResourcePage("campaigns", 1, 1);
  const forms = useMarketingQuery<LeadForm[]>("marketing/forms");
  return {
    ready: !contacts.isLoading && !campaigns.isLoading && !forms.isLoading,
    contacts: contacts.data?.total ?? 0,
    campaigns: campaigns.data?.total ?? 0,
    forms: forms.data?.length ?? 0,
  };
}

const plural = (n: number, one: string) => `${n.toLocaleString()} ${one}${n === 1 ? "" : "s"}`;

/** "You have 1,240 contacts, 3 campaigns and 2 forms" with each count linking to its page. */
export function GlanceLine() {
  const g = useGlance();
  if (!g.ready) return <p className={styles.subtitle}>&nbsp;</p>;
  if (!g.contacts && !g.campaigns && !g.forms)
    return (
      <p className={styles.subtitle}>
        Your workspace is fresh. Start by{" "}
        <Link href="/audience/lists" className={styles.chipBlue}>adding contacts</Link>,{" "}
        <Link href="/templates" className={styles.chipAmber}>picking a template</Link> and{" "}
        <Link href="/campaigns/new" className={styles.chipPurple}>sending a campaign</Link>.
      </p>
    );
  return (
    <p className={styles.subtitle}>
      You have{" "}
      <Link href="/crm" className={styles.chipBlue}>{plural(g.contacts, "contact")}</Link>,{" "}
      <Link href="/campaigns" className={styles.chipAmber}>{plural(g.campaigns, "campaign")}</Link> and{" "}
      <Link href="/forms" className={styles.chipPurple}>{plural(g.forms, "form")}</Link>. Ask about any of them.
    </p>
  );
}
