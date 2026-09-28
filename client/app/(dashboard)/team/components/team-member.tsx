"use client";

import { useTeam } from "@/app/providers/team-provider";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";
import { toast } from "sonner";
import { logger } from "@/app/lib/logger";
import { useSession } from "next-auth/react";

interface TeamMember {
  id: string;
  email: string;
  name: string | null;
  role: "ADMIN" | "USER";
}

export function TeamMemberCard({ member }: { member: TeamMember }) {
  const { refreshTeam } = useTeam();
  const { data: session } = useSession();

  const handleRemoveMember = async () => {
    try {
      logger.info({
        fileName: "team-member.tsx",
        action: "handleRemoveMember",
        label: "memberId",
        value: member.id,
        emoji: "👤",
        message: "Removing team member",
      });

      const response = await fetch(`/api/team/member/${member.id}`, {
        method: "DELETE",
      });

      if (!response.ok) throw new Error("Failed to remove team member");

      toast.success("Team member removed successfully");
      refreshTeam();
    } catch (error) {
      logger.error({
        fileName: "team-member.tsx",
        action: "handleRemoveMember",
        label: "error",
        value: error,
        emoji: "❌",
        message: "Failed to remove team member",
      });
      toast.error("Failed to remove team member");
    }
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>{member.name || member.email}</CardTitle>
        <CardDescription>{member.email}</CardDescription>
      </CardHeader>
      <CardContent>
        <p>Role: {member.role}</p>
      </CardContent>
      <CardFooter>
        {session?.user?.email !== member.email && (
          <Button variant="destructive" onClick={handleRemoveMember}>
            Remove
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
