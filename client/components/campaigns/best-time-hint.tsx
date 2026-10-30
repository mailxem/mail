"use client";

import { useMemo } from "react";
import { format } from "date-fns";
import { Clock3 } from "lucide-react";
import { useAnalytics } from "@/lib/analytics/use-analytics";
import type { Report } from "@/lib/analytics/types";
import {
  formatHour,
  nextOccurrence,
  peakHour,
} from "@/lib/analytics/best-time";
import { Button } from "@/components/ui/button";

const day = (offset: number) => {
  const d = new Date();
  d.setUTCDate(d.getUTCDate() + offset);
  return d.toISOString().slice(0, 10);
};

/** Suggests when to send from when readers clicked over the last 90 days. */
export function BestTimeHint({
  timezone,
  onPick,
}: {
  timezone: string;
  onPick: (date: Date) => void;
}) {
  const params = useMemo(
    () => new URLSearchParams({ from: day(-89), to: day(1), timezone }),
    [timezone],
  );
  const report = useAnalytics<Report>("report", params);
  const peak = report.data?.charts?.clickHours
    ? peakHour(report.data.charts.clickHours)
    : null;
  if (!peak) return null;
  const next = nextOccurrence(peak.day, peak.hour);
  return (
    <div className="flex flex-wrap items-center gap-3 rounded-xl border border-border bg-muted/20 p-3">
      <Clock3 size={16} className="shrink-0 text-primary" />
      <p className="min-w-0 flex-1 text-xs leading-relaxed">
        Your readers click most on{" "}
        <strong>
          {peak.dayName}s around {formatHour(peak.hour)}
        </strong>
        <span className="text-muted-foreground">
          {" "}
          ({peak.total.toLocaleString()} clicks in the last 90 days,{" "}
          {timezone}).
        </span>
      </p>
      <Button
        type="button"
        size="sm"
        variant="outline"
        onClick={() => onPick(next)}
      >
        Use {format(next, "EEE, MMM d · h:mm a")}
      </Button>
    </div>
  );
}
