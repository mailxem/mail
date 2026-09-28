import { workspaceClassName } from "@/lib/workspace-styles";
import { Metadata } from "next";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TeamInfo } from "./components/team-info";
import { InviteTeamMember } from "./components/invite-member";
import { TeamInvites } from "./components/team-invites";
import { Users, Mail, BarChart2 } from "lucide-react";
import { CurrentTeamAnalytics } from "@/components/analytics/current-team-analytics";

export const metadata: Metadata = {
  title: "Team | Xem",
  description: "Manage your team members and settings",
};

export default function TeamPage() {
  return (
    <div className={workspaceClassName("workspace-page")}>
      <div className={workspaceClassName("workspace-page-body space-y-6")}>
        <Tabs defaultValue="members" className="space-y-4">
          <TabsList className="border-b border-muted rounded-none w-full justify-start gap-6 bg-transparent h-auto p-0">
            <TabsTrigger
              value="members"
              className="flex items-center gap-2 data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-12 px-4"
            >
              <Users className="h-4 w-4" />
              Members
            </TabsTrigger>
            <TabsTrigger
              value="invites"
              className="flex items-center gap-2 data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-12 px-4"
            >
              <Mail className="h-4 w-4" />
              Invites
            </TabsTrigger>
            <TabsTrigger
              value="analytics"
              className="flex items-center gap-2 data-[state=active]:border-b-2 data-[state=active]:border-primary rounded-none h-12 px-4"
            >
              <BarChart2 className="h-4 w-4" />
              Analytics
            </TabsTrigger>
          </TabsList>

          <TabsContent value="members" className="space-y-4">
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7">
              <div className="col-span-4 border border-muted rounded-lg p-6">
                <h3 className="font-medium mb-4">Current Team Members</h3>
                <TeamInfo />
              </div>
              <div className="col-span-3 border border-muted rounded-lg p-6">
                <h3 className="font-medium mb-4">Invite New Member</h3>
                <InviteTeamMember />
              </div>
            </div>
          </TabsContent>

          <TabsContent value="invites" className="space-y-4">
            <div className="border border-muted rounded-lg p-6">
              <h3 className="font-medium mb-4">Pending Invites</h3>
              <TeamInvites />
            </div>
          </TabsContent>

          <TabsContent value="analytics" className="space-y-4">
            <div className="border border-muted rounded-lg p-6">
              <CurrentTeamAnalytics/>
            </div>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
