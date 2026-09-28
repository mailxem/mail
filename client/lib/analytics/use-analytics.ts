"use client";
import { useQuery } from "@tanstack/react-query";
import { useTeam } from "@/app/providers/team-provider";
export function useAnalytics<T>(
  resource: string,
  params: URLSearchParams,
  enabled = true,
) {
  const { team } = useTeam();
  return useQuery<T, Error>({
    queryKey: ["analytics-v2", team?.id, resource, params.toString()],
    enabled: !!team?.id && enabled,
    staleTime: 60000,
    retry: 1,
    queryFn: async ({ signal }) => {
      const query = new URLSearchParams(params);
      query.set("teamId", team!.id);
      const response = await fetch(`/api/analytics/v2/${resource}?${query}`, {
        signal,
        cache: "no-store",
      });
      const body = await response.json();
      if (!response.ok)
        throw new Error(body.error || "Unable to load analytics");
      if (
        !body ||
        (resource === "report" && body.metricVersion !== "audience-v2") ||
        (resource === "email-overview" && body.metricVersion !== "email-v1")
      )
        throw new Error("Analytics returned an unsupported response");
      return body as T;
    },
  });
}
