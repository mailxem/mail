"use client";

import { createContext, useContext, useState } from "react";
import { EmailTemplate } from "@/lib";
import { useResourcePage } from "@/hooks/use-resource-page";

type Pagination = { page: number; limit: number; total: number };
interface ContextValue {
  templates: EmailTemplate[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  pagination: Pagination;
  setPagination: (pagination: Pagination) => void;
}
const Context = createContext<ContextValue | null>(null);
export function useTemplates() {
  const value = useContext(Context);
  if (!value) throw new Error("useTemplates requires TemplatesProvider");
  return value;
}
export function TemplatesProvider({ children }: { children: React.ReactNode }) {
  const [paging, setPaging] = useState({ page: 1, limit: 50 });
  const query = useResourcePage<EmailTemplate>("templates", paging.page, paging.limit);
  return <Context.Provider value={{
    templates: query.data?.data ?? [],
    isLoading: query.isLoading,
    error: query.error,
    pagination: { ...paging, total: query.data?.total ?? 0 },
    setPagination: (next) => setPaging({ page: Math.max(1, next.page), limit: next.limit }),
    refetch: async () => { await query.refetch(); },
  }}>{children}</Context.Provider>;
}
