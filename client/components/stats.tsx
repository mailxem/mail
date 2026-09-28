"use client";
import { useAnalytics } from "@/lib/analytics/use-analytics";
import { Report, formatRate } from "@/lib/analytics/types";
import { Card, CardContent } from "@/components/ui/card";
export function Stats() {
  const query = useAnalytics<Report>("report", new URLSearchParams());
  const s = query.data?.summary;
  const metrics = [
    [
      "Messages accepted",
      s?.accepted.toLocaleString(),
      "SMTP accepted, last 30 days",
    ],
    [
      "Recipients reached",
      s?.recipientsReached.toLocaleString(),
      "Unique recipients",
    ],
    [
      "Message click rate",
      s ? formatRate(s.clickRate) : undefined,
      "Clicked / accepted messages",
    ],
    [
      "Known bounce rate",
      s ? formatRate(s.bounceRate) : undefined,
      "Captured bounces / accepted",
    ],
  ];
  return (
    <>
      {metrics.map(([title, value, detail]) => (
        <Card key={title} className="shadow-none">
          <CardContent className="p-5">
            <p className="text-sm text-muted-foreground">{title}</p>
            <p className="mt-3 text-3xl font-semibold tabular-nums">
              {value ?? "—"}
            </p>
            <p className="mt-2 text-xs text-muted-foreground">
              {query.error
                ? "Analytics unavailable"
                : query.isPending
                  ? "Loading analytics…"
                  : detail}
            </p>
            {query.error && (
              <button
                onClick={() => void query.refetch()}
                className="mt-2 text-xs underline"
              >
                Try again
              </button>
            )}
          </CardContent>
        </Card>
      ))}
    </>
  );
}
