"use client";

import { createContext, useContext, useState } from "react";
import { useTeam } from "./team-provider";
import { SMTPConfig } from "@/lib/validations/smtp-provider";
import { useApi } from "@/hooks/use-api";
import { useResourcePage } from "@/hooks/use-resource-page";

type SMTPContextType = {
  configs: SMTPConfig[];
  isLoading: boolean;
  error: Error | null;
  refresh: () => void;
};

const SMTPContext = createContext<SMTPContextType>({
  configs: [],
  isLoading: true,
  error: null,
  refresh: () => {},
});

export function useSMTP() {
  return useContext(SMTPContext);
}

export function SMTPProvider({ children }: { children: React.ReactNode }) {
  const query = useResourcePage<SMTPConfig>("smtp-configs", 1, 100);
  return <SMTPContext.Provider value={{
    configs: query.data?.data ?? [], isLoading: query.isLoading, error: query.error,
    refresh: () => { void query.refetch(); },
  }}>{children}</SMTPContext.Provider>;
}
