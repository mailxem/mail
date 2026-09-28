"use client";

import { createContext, useContext, useState } from "react";
import { useTeam } from "./team-provider";
import { IMAPConfig } from "@/lib/validations/imap-provider";
import { useApi } from "@/hooks/use-api";
import { useResourcePage } from "@/hooks/use-resource-page";

type IMAPContextType = {
  configs: IMAPConfig[];
  isLoading: boolean;
  error: Error | null;
  refresh: () => void;
};

const IMAPContext = createContext<IMAPContextType>({
  configs: [],
  isLoading: true,
  error: null,
  refresh: () => {},
});

export function useIMAP() {
  return useContext(IMAPContext);
}

export function IMAPProvider({ children }: { children: React.ReactNode }) {
  const query = useResourcePage<IMAPConfig>("imap", 1, 100);
  return <IMAPContext.Provider value={{
    configs: query.data?.data ?? [], isLoading: query.isLoading, error: query.error,
    refresh: () => { void query.refetch(); },
  }}>{children}</IMAPContext.Provider>;
}
