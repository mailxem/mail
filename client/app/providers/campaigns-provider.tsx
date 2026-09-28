"use client";
import { resourceEntity } from "@/lib/resource-response";

import { createContext, useContext, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTeam } from "./team-provider";
import { Campaign } from "@/lib";
import { useApi } from "@/hooks/use-api";
import { useResourcePage } from "@/hooks/use-resource-page";

interface CampaignsContextType {
  campaigns: Campaign[];
  loading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  getCampaign: (id: string) => Promise<Campaign>;
  campaign: Campaign | null;
  total: number;
  page: number;
  limit: number;
  setPage: (page: number) => void;
  setLimit: (limit: number) => void;
}
const Context = createContext<CampaignsContextType | null>(null);
export function CampaignsProvider({ children }: { children: React.ReactNode }) {
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [campaign, setCampaign] = useState<Campaign | null>(null);
  const { apiFetch } = useApi();
  const query = useResourcePage<Campaign>("campaigns", page, limit);
  const getCampaign = async (id: string) => {
    const response = await apiFetch(`campaigns/${id}?include=Template`);
    if (!response.ok) throw new Error("Unable to load this campaign.");
    const result = await response.json();
    const data = resourceEntity<Campaign>(result);
    setCampaign(data);
    return data as Campaign;
  };
  return <Context.Provider value={{
    campaigns: query.data?.data ?? [], loading: query.isLoading, error: query.error,
    total: query.data?.total ?? 0, page, limit, setPage, setLimit,
    refetch: async () => { await query.refetch(); }, getCampaign, campaign,
  }}>{children}</Context.Provider>;
}
export function useCampaigns() {
  const value = useContext(Context);
  if (!value) throw new Error("useCampaigns requires CampaignsProvider");
  return value;
}
