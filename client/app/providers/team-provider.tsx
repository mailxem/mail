"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { useSession } from "next-auth/react";
import { logger } from "@/app/lib/logger";
import { Team, User, TeamInvite, Domain } from "@/lib";

interface TeamWithUsers extends Team {
  users: User[];
  id: string;
  firstName: string;
  lastName: string;
  invites: TeamInvite[];
  customDomains: Domain[];
  emailTemplateId: string;
}

interface TeamContextType {
  team: TeamWithUsers | null;
  loading: boolean;
  error: Error | null;
  refreshTeam: () => Promise<void>;
}


const TeamContext = createContext<TeamContextType | undefined>(undefined);

export function TeamProvider({ children }: { children: React.ReactNode }) {
  const { data: session } = useSession();
  const [team, setTeam] = useState<Team | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchTeamData = async () => {
    try {
      if (!session?.user?.teamId) {
        setTeam(null);
        return;
      }

      const response = await fetch(`/api/team`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });
      if (!response.ok) throw new Error("Failed to fetch team data");

      const teamData = await response.json();

      setTeam(teamData);
      setError(null);

      logger.info({
        fileName: "team-provider.tsx",
        action: "fetchTeamData",
        label: "teamData",
        value: { id: teamData.id },
        emoji: "✅",
        message: "Team data fetched successfully",
      });
    } catch (err) {
      logger.error({
        fileName: "team-provider.tsx",
        action: "fetchTeamData",
        label: "error",
        value: err,
        emoji: "❌",
        message: "Error fetching team data",
      });
      setError(
        err instanceof Error ? err : new Error("Failed to fetch team data")
      );
    } finally {
      setLoading(false);
    }
  };

  const refreshTeam = async () => {
    setLoading(true);
    await fetchTeamData();
  };

  useEffect(() => {
    fetchTeamData();
  }, [session?.user?.teamId]);

  return (
    <TeamContext.Provider
      value={{ team: team as TeamWithUsers, loading, error, refreshTeam }}
    >
      {children}
    </TeamContext.Provider>
  );
}

export function useTeam() {
  const context = useContext(TeamContext);
  if (context === undefined) {
    throw new Error("useTeam must be used within a TeamProvider");
  }
  return context;
}
