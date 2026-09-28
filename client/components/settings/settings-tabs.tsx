"use client";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Bell, CreditCard, Globe, Palette, User } from "lucide-react";
import { ProfileSettings } from "./profile-settings";
import { BrandingSettings } from "./branding-settings";
import { CustomDomains } from "./custom-domains";
import dynamic from "next/dynamic";
import styles from "./settings-tabs.module.css";

const BillingWrapper = dynamic(
  () => import("@/app/settings/billing/billing-wrapper"),
  { ssr: false },
);

const sections = [
  { id: "profile", title: "Profile", icon: User },
  { id: "branding", title: "Branding", icon: Palette },
  { id: "domains", title: "Domains", icon: Globe },
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
      <TabsContent value="branding" className={styles.content}><BrandingSettings /></TabsContent>
      <TabsContent value="domains" className={styles.content}><CustomDomains /></TabsContent>
      <TabsContent value="notifications" className={styles.content}>
        <section className={styles.section} aria-labelledby="notification-settings-heading">
          <div className={styles.sectionIntro}>
            <h2 id="notification-settings-heading">Notifications</h2>
            <p>Account and campaign updates.</p>
          </div>
          <div className={styles.unavailable}>
            <Bell size={18} strokeWidth={1.65} />
            <div><h3>Preferences are not available yet</h3><p>Campaign delivery and engagement are available in Analytics. Delivery failures are recorded in Outbox.</p></div>
          </div>
        </section>
      </TabsContent>
      <TabsContent value="billing" className={styles.content}><BillingWrapper /></TabsContent>
    </Tabs>
  );
}
