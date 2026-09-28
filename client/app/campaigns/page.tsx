"use client";

import { CampaignsList } from "@/components/campaigns/campaigns-list";
import { CampaignsProvider, useCampaigns } from "@/app/providers/campaigns-provider";
import { useState } from "react";
import { FullScreenCalendar } from "@/components/ui/full-screen-calendar";
import { Metric, FilterTabs } from "@/components/marketing/shared";
import { workspaceClassName } from "@/lib/workspace-styles";
import { Mail, CalendarClock, FilePenLine, CheckCheck } from "lucide-react";

function CampaignCollection() {
  const { campaigns, total, loading } = useCampaigns();
  const [view, setView] = useState("List");
  const calendar = campaigns.flatMap(campaign => {
    if (!campaign.scheduledFor) return [];
    const date = new Date(campaign.scheduledFor);
    if (Number.isNaN(date.getTime())) return [];
    return [{ day: date, events: [{ id: campaign.id, name: campaign.name, time: date.toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" }), datetime: date.toISOString(), href: `/campaigns/${campaign.id}` }] }];
  });
  const count = (status: string) => loading ? "—" : campaigns.filter(c => c.status === status).length;
  return <>
    <div className="mb-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <Metric label="Total campaigns" value={loading ? "—" : total} icon={<Mail size={18}/>}/>
      <Metric label="Scheduled on this page" value={count("SCHEDULED")} icon={<CalendarClock size={18}/>}/>
      <Metric label="Drafts on this page" value={count("DRAFT")} icon={<FilePenLine size={18}/>}/>
      <Metric label="Completed on this page" value={count("COMPLETED")} icon={<CheckCheck size={18}/>}/>
    </div>
    <section className={workspaceClassName("product-panel")}>
      <div className={workspaceClassName("panel-toolbar")}><h2>All campaigns</h2><FilterTabs items={["List", "Calendar"]} value={view} onChange={setView}/></div>
      {view === "List" ? <CampaignsList/> : <><p className="mb-4 text-xs text-muted-foreground">Schedule for the {campaigns.length} campaigns on the current list page.</p><FullScreenCalendar data={calendar}/></>}
    </section>
  </>;
}
export default function CampaignsPage() {
  return <CampaignsProvider><CampaignCollection/></CampaignsProvider>;
}
