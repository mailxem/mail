"use client";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Bell, Clock, CreditCard, Globe, Palette, PenLine, User } from "lucide-react";
import { ProfileSettings } from "./profile-settings";
import { BrandingSettings, TrackingConsentSettings } from "./branding-settings";
import { SendingRulesSettings } from "./sending-rules-settings";
import { SummaryEmailSettings } from "./summary-email-settings";
import { CustomDomains } from "./custom-domains";
import { AuthorsSettings } from "@/components/marketing/authors";
import dynamic from "next/dynamic";
import styles from "./settings-tabs.module.css";

const BillingWrapper = dynamic(
  () => import("@/app/settings/billing/billing-wrapper"),
  { ssr: false },
);

const sections = [
  { id: "profile", title: "Profile", icon: User },
  { id: "branding", title: "Branding", icon: Palette },
  { id: "authors", title: "Authors", icon: PenLine },
  { id: "domains", title: "Domains", icon: Globe },
  { id: "sending-rules", title: "Sending rules", icon: Clock },
  { id: "notifications", title: "Notifications", icon: Bell },
  { id: "billing", title: "Billing", icon: CreditCard },
];

export function SettingsTabs() {
  const [tab, setTab] = useState("profile");
  return (
    <Tabs value={tab} onValueChange={setTab} className={styles.settings}>
      <TabsList aria-label="Settings sections" className={styles.tabList}>
        {sections.map(({ id, title, icon: Icon }) => (
          <TabsTrigger key={id} value={id} className={styles.tab}><Icon size={15} strokeWidth={1.65} />{title}</TabsTrigger>
        ))}
      </TabsList>
      <TabsContent value="profile" className={styles.content}>
        <section className={styles.section} aria-labelledby="profile-settings-heading">
          <div className={styles.sectionIntro}>
            <h2 id="profile-settings-heading">Your profile</h2>
            <p>The details attached to your Xem account.</p>
          </div>
          <div className={styles.profileForm}><ProfileSettings /></div>
        </section>
      </TabsContent>
      <TabsContent value="branding" className={styles.content}><BrandingSettings /><TrackingConsentSettings /></TabsContent>
      <TabsContent value="authors" className={styles.content}><AuthorsSettings /></TabsContent>
      <TabsContent value="domains" className={styles.content}><CustomDomains /></TabsContent>
      <TabsContent value="sending-rules" className={styles.content}><SendingRulesSettings /></TabsContent>
      <TabsContent value="notifications" className={styles.content}><SummaryEmailSettings /></TabsContent>
      <TabsContent value="billing" className={styles.content}><BillingWrapper /></TabsContent>
    </Tabs>
  );
}
