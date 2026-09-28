"use client";
import { createContext, useContext } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "@/hooks/use-api";
export type Transport = <T>(
  path: string,
  method?: string,
  body?: unknown,
) => Promise<T>;
export const PreviewTransport = createContext<Transport | null>(null);
export class MarketingRequestError extends Error {
  constructor(message: string, public readonly status: number) {
    super(message);
    this.name = "MarketingRequestError";
  }
}
export function useMarketing() {
  const { apiFetch, status, session } = useApi();
  const preview = useContext(PreviewTransport);
  const queryClient = useQueryClient();
  const request: Transport = async (path, method = "GET", body) => {
    if (preview) return preview(path, method, body);
    const response = await apiFetch(path, {
      method,
      ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
    });
    const data = response.status === 204 ? null : await response.json().catch((error) => {
      if (response.ok) throw error;
      return null;
    });
    if (!response.ok)
      throw new MarketingRequestError(
        typeof data?.error === "string"
          ? data.error
          : typeof data?.message === "string" ? data.message
          : "The request could not be completed. Please try again.",
        response.status,
      );
    return data;
  };
  return {
    request,
    ready: !!preview || status === "authenticated",
    scope: preview ? "preview" : session?.user?.teamId,
    refresh: () => queryClient.invalidateQueries({ queryKey: ["marketing"] }),
  };
}
export function useMarketingQuery<T>(path: string, enabled = true, refetchInterval?: number) {
  const { request, ready, scope } = useMarketing();
  return useQuery<T, Error>({
    queryKey: ["marketing", scope, path],
    queryFn: () => request<T>(path),
    enabled: ready && enabled,
    retry: 1,
    refetchInterval,
    refetchIntervalInBackground: false,
  });
}
