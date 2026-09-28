"use client";
import { Suspense, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useAnalytics } from "@/lib/analytics/use-analytics";
import { AnalyticsPage } from "@/lib/analytics/types";
import { PageHeading, QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
const names: Record<string, string> = {
  recent: "Clicked in 30 days",
  earlier: "Clicked 31–90 days ago",
  no_clicks: "No clicks after 3+ tracked sends",
  not_contacted: "Not contacted in 90 days",
  insufficient: "Insufficient tracking evidence",
};
type Person = {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  listId: string;
  cohort: string;
};
export function AudiencePeople() {
  return (
    <Suspense fallback={<p>Loading audience…</p>}>
      <People />
    </Suspense>
  );
}
function People() {
  const search = useSearchParams();
  const [page, setPage] = useState(1);
  const params = new URLSearchParams(search);
  params.set("page", String(page));
  params.set("limit", "20");
  const query = useAnalytics<AnalyticsPage<Person>>("people", params);
  const data = query.data;
  const back = new URLSearchParams(search);
  back.delete("cohort");
  back.delete("asOf");
  return (
    <div className="space-y-6">
      <PageHeading
        title={names[search.get("cohort") || ""] || "Audience contacts"}
        description="Unique active recipients matching this cohort. The report cutoff and list/tag filters are preserved."
        action={
          <Button asChild variant="outline">
            <Link href={`/audience/dashboard?${back}`}>Back to audience</Link>
          </Button>
        }
      />
      <Card className="overflow-hidden shadow-none">
        {query.isPending || query.error ? (
          <div className="p-6">
            <QueryState
              loading={query.isPending}
              error={query.error}
              retry={() => void query.refetch()}
            />
          </div>
        ) : !data?.items.length ? (
          <p className="p-6 text-sm text-muted-foreground">
            No active recipients match this cohort.
          </p>
        ) : (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="pl-6">Contact</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Review</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.items.map((person) => (
                  <TableRow key={person.id}>
                    <TableCell className="pl-6 font-medium">
                      {[person.firstName, person.lastName]
                        .filter(Boolean)
                        .join(" ") || person.email}
                    </TableCell>
                    <TableCell>{person.email}</TableCell>
                    <TableCell>
                      <Link
                        className="text-violet-700 hover:underline"
                        href={`/crm?search=${encodeURIComponent(person.email)}`}
                      >
                        Open in CRM ↗
                      </Link>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            <div className="flex items-center justify-between border-t p-4 text-sm">
              <span>
                {data.totalCount.toLocaleString()} recipients · Page {data.page}
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1 || query.isFetching}
                  onClick={() => setPage(page - 1)}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={
                    page * data.pageSize >= data.totalCount || query.isFetching
                  }
                  onClick={() => setPage(page + 1)}
                >
                  Next
                </Button>
              </div>
            </div>
          </>
        )}
      </Card>
      <p className="text-xs text-muted-foreground">
        Review these contacts before taking action. Tracking signals alone never
        trigger a send or change subscription status.
      </p>
    </div>
  );
}
