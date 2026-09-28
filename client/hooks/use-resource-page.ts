"use client";

import { useQuery } from "@tanstack/react-query";
import { useTeam } from "@/app/providers/team-provider";
import { useApi } from "@/hooks/use-api";

export interface ResourcePage<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
}

/** Keep server data in the query cache so it survives client-side navigation. */
export function useResourcePage<T>(resource: string, page: number, limit: number, filters: Record<string, string> = {}) {
  const { team, error: teamError } = useTeam();
  const { apiFetch, session } = useApi();
  const query = useQuery<ResourcePage<T>>({
    queryKey: [resource, team?.id, page, limit, filters],
    enabled: !!team?.id && !!session?.accessToken,
    queryFn: async ({ signal }) => {
      const params = new URLSearchParams({ ...filters, team_id: team!.id, page: String(page), limit: String(limit) });
      const response = await apiFetch(`${resource}?${params}`, { signal });
      if (!response.ok) throw new Error(`Unable to load ${resource.replaceAll("-", " ")}. Please try again.`);
      const result = await response.json();
      return { ...result, data: result.data ?? [] } as ResourcePage<T>;
    },
    staleTime: 30_000,
  });
  return { ...query, error: query.error || teamError, isLoading: !teamError && query.isPending };
}
