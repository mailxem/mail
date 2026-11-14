"use client";
import { useEffect } from "react";
import { Rows3, TriangleAlert } from "lucide-react";
import { useMarketingQuery } from "@/lib/marketing/api";
import { describeRule, type SectionRule } from "@/lib/sections";

export type SectionReachResult = {
  sections: (SectionRule & { id?: string; name: string; readers: number })[];
  readers: number;
  empty: number;
};

// Lists the template's limited sections with how many of the list's readers
// get each one, and reports readers who would get an email with nothing in it
// (the send is refused for them on the server too).
export function SectionReach({
  templateId,
  listId,
  onEmpty,
}: {
  templateId?: string;
  listId?: string;
  onEmpty: (readers: number) => void;
}) {
  const reach = useMarketingQuery<SectionReachResult>(
    `marketing/templates/${templateId}/sections?listId=${listId}`,
    !!templateId && !!listId,
  );
  const segments = useMarketingQuery<{ data: { id: string; name: string }[] }>(
    "segments?limit=100",
    !!reach.data?.sections.some((s) => s.segments?.length),
  );
  const empty = templateId && listId ? (reach.data?.empty ?? 0) : 0;
  useEffect(() => onEmpty(empty), [empty, onEmpty]);
  if (!reach.data || (!reach.data.sections.length && !empty)) return null;
  const { sections, readers } = reach.data;
  const segmentName = (id: string) =>
    segments.data?.data.find((s) => s.id === id)?.name ?? "removed segment";

  return (
    <section
      aria-labelledby="section-reach-heading"
      className="rounded-[20px] border border-border bg-card p-5 shadow-sm"
    >
      <div className="flex items-center justify-between">
        <h2 id="section-reach-heading" className="text-sm font-medium">
          Sections
        </h2>
        <Rows3 size={16} className="text-muted-foreground" />
      </div>
      <ul className="mt-3 space-y-3">
        {sections.map((s, i) => (
          <li key={s.id || i} className="text-xs">
            <div className="flex items-baseline justify-between gap-3">
              <span className="font-medium">{s.name}</span>
              <span className="shrink-0 tabular-nums text-muted-foreground">
                {s.readers.toLocaleString()} of {readers.toLocaleString()}
              </span>
            </div>
            <p className="mt-0.5 text-muted-foreground">
              {describeRule(s, s.tags, (s.segments ?? []).map(segmentName))}
            </p>
          </li>
        ))}
      </ul>
      {empty > 0 && (
        <p
          role="alert"
          className="mt-4 flex gap-2 rounded-lg bg-destructive/10 p-3 text-xs text-destructive"
        >
          <TriangleAlert size={14} className="mt-0.5 shrink-0" />
          {empty === 1 ? "1 reader" : `${empty.toLocaleString()} readers`}{" "}
          would get this email with nothing in it, because none of its
          sections are for them. Change who the sections are for in the
          template before sending.
        </p>
      )}
    </section>
  );
}
