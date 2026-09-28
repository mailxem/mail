"use client";

import { useTeam } from "@/app/providers/team-provider";
import { Button } from "@/components/ui/button";
import { useApi } from "@/hooks/use-api";
import { useCallback } from "react";
import { toast } from "sonner";

export function TeamInvites() {
  const { team, refreshTeam } = useTeam();
  const { apiFetch } = useApi();

  const handleCancelInvite = async (code: string) => {
    try {
      const response = await apiFetch(`/users/invite/${code}`, {
        method: "DELETE",
      });

      if (!response.ok) throw new Error("Failed to cancel invite");

      toast.success("Invite cancelled successfully");
      refreshTeam();
    } catch (error) {
      toast.error("Failed to cancel invite");
    }
  };

  const resendInvite = async (code: string) => {
    try {
      const response = await apiFetch(`/users/invite/resend/${code}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({}),
      });

      if (!response.ok) throw new Error("Failed to resend invite");

      toast.success("Invite resent successfully");
    } catch (error) {
      toast.error("Failed to resend invite");
    }
  };

  const pendingInvites = team?.invites?.filter(
    (invite) => invite.status === "PENDING" && invite.expiresAt > new Date(),
  );

  const getInviter = useCallback(() => {
    return (
      team?.users.find((user) => user.id === pendingInvites?.[0]?.inviterId) ??
      null
    );
  }, [pendingInvites, team?.users]);

  if (!pendingInvites?.length) {
    return <p className="text-sm text-muted-foreground">No pending invites</p>;
  }

  return (
    <div className="space-y-4">
      {pendingInvites.map((invite) => (
        <div
          key={invite.id}
          className="flex items-center justify-between p-4 border border-muted rounded-lg"
        >
          <div>
            <p className="font-medium">
              {" "}
              {invite.name} &lt;{invite.email}&gt;
            </p>
            <p className="text-sm text-muted-foreground">
              Invited by:{" "}
              {getInviter()?.firstName + " " + getInviter()?.lastName}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => resendInvite(invite.code)}
              disabled={invite.status !== "PENDING"}
            >
              Resend Invite
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => handleCancelInvite(invite.code)}
            >
              Cancel Invite
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
