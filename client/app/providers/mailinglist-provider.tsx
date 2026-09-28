"use client";

import { createContext, useContext, useState } from "react";
import { MailingList } from "@/lib";
import { useResourcePage } from "@/hooks/use-resource-page";

type Pagination = { page: number; limit: number; total: number };
interface ContextValue {
  lists: MailingList[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  pagination: Pagination;
  setPagination: (pagination: Pagination) => void;
}
const Context = createContext<ContextValue | null>(null);
export function useMailingLists() {
  const value = useContext(Context);
  if (!value) throw new Error("useMailingLists requires MailingListProvider");
  return value;
}
export function MailingListProvider({ children }: { children: React.ReactNode }) {
  const [paging, setPaging] = useState({ page: 1, limit: 20 });
  const query = useResourcePage<MailingList>("mailing-lists", paging.page, paging.limit, { sort: "subscribers_count", order: "desc" });
  return <Context.Provider value={{
    lists: query.data?.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    pagination: { ...paging, total: query.data?.total ?? 0 },
    setPagination: (next) => setPaging({ page: Math.max(1, next.page), limit: next.limit }),
    refetch: async () => { await query.refetch(); },
  }}>{children}</Context.Provider>;
}
