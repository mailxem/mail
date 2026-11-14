"use client";
import { useEffect, useState } from "react";
import { Loader2 } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useMarketingQuery } from "@/lib/marketing/api";
import {
  describeRule,
  designRows,
  isLimited,
  ruleProblem,
  type SectionRule,
  type SectionRules,
} from "@/lib/sections";

type Named = { id: string; name: string };
const field =
  "h-9 w-full rounded-md border border-input bg-background px-2 text-sm";

// Lets each row of the design be kept for some readers, or for the email or
// the newsletter website only. The rules are saved with the template.
export function TemplateSections({
  readDesign,
  rules,
  onChange,
}: {
  readDesign: () => Promise<unknown>;
  rules: SectionRules;
  onChange: (rules: SectionRules) => void;
}) {
  const [rows, setRows] = useState<ReturnType<typeof designRows> | null>(null);
  const [error, setError] = useState("");
  const [open, setOpen] = useState("");
  const tags = useMarketingQuery<Named[]>("marketing/tags");
  const segments = useMarketingQuery<{ data: Named[] }>("segments?limit=100");
  const load = () => {
    setError("");
    readDesign()
      .then((design) => setRows(designRows(design)))
      .catch((e: Error) => setError(e.message));
  };
  useEffect(load, []); // eslint-disable-line react-hooks/exhaustive-deps

  const set = (id: string, patch: Partial<SectionRule>) =>
    onChange({ ...rules, [id]: { ...rules[id], ...patch } });
  const toggle = (list: string[] | undefined, value: string) =>
    list?.includes(value) ? list.filter((v) => v !== value) : [...(list ?? []), value];
  const segmentNames = (ids: string[] = []) =>
    ids.map((id) => segments.data?.data.find((s) => s.id === id)?.name ?? "removed segment");

  return (
    <div className="mx-auto max-w-3xl space-y-4 p-6">
      <div>
        <h2 className="text-base font-medium">Sections</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Keep a row for readers with a tag or in a segment, or for everyone
          except them. A row can also show in the email only or on the
          newsletter website only. Who gets what is decided when each email is
          sent, so readers never receive rows that aren’t for them. Save the
          template to keep changes.
        </p>
      </div>
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}{" "}
          <button type="button" className="underline" onClick={load}>
            Try again
          </button>
        </p>
      )}
      {!rows && !error && (
        <p className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin motion-reduce:animate-none" />
          Reading the design…
        </p>
      )}
      {rows?.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Add rows to the design first.
        </p>
      )}
      <ol className="divide-y divide-border rounded-xl border border-border">
        {rows?.map((row) => {
          const rule = rules[row.id] ?? {};
          const problem = ruleProblem(rule);
          return (
            <li key={row.id || row.index} className="p-4">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-sm font-medium">
                    {rule.name || `Row ${row.index + 1}`}
                  </p>
                  <p className="truncate text-xs text-muted-foreground">
                    {row.summary}
                  </p>
                  <p className="mt-1 text-xs">
                    {isLimited(rule)
                      ? describeRule(rule, rule.tags, segmentNames(rule.segments))
                      : "Everyone, in the email and on the website"}
                  </p>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={!row.id}
                  onClick={() => setOpen(open === row.id ? "" : row.id)}
                >
                  {open === row.id ? "Done" : "Change"}
                </Button>
              </div>
              {open === row.id && (
                <div className="mt-4 grid gap-3 sm:grid-cols-2">
                  <label className="space-y-1 text-xs text-muted-foreground">
                    Name (only you see it)
                    <Input
                      value={rule.name ?? ""}
                      maxLength={80}
                      placeholder={`Row ${row.index + 1}`}
                      onChange={(e) => set(row.id, { name: e.target.value })}
                    />
                  </label>
                  <label className="space-y-1 text-xs text-muted-foreground">
                    Shows
                    <select
                      className={field}
                      value={rule.show ?? ""}
                      onChange={(e) =>
                        set(row.id, { show: e.target.value as SectionRule["show"] })
                      }
                    >
                      <option value="">In the email and on the website</option>
                      <option value="email">In the email only</option>
                      <option value="web">On the website only</option>
                    </select>
                  </label>
                  <label className="space-y-1 text-xs text-muted-foreground sm:col-span-2">
                    Readers
                    <select
                      className={field}
                      value={rule.audience ?? ""}
                      onChange={(e) =>
                        set(row.id, {
                          audience: e.target.value as SectionRule["audience"],
                        })
                      }
                    >
                      <option value="">Everyone</option>
                      <option value="only">Only readers with these tags or segments</option>
                      <option value="except">Everyone except readers with these tags or segments</option>
                    </select>
                  </label>
                  {rule.audience && (
                    <>
                      <fieldset className="space-y-1">
                        <legend className="text-xs text-muted-foreground">Tags</legend>
                        <div className="max-h-40 space-y-1 overflow-y-auto rounded-md border border-border p-2">
                          {tags.data?.length ? (
                            tags.data.map((tag) => (
                              <label key={tag.id} className="flex items-center gap-2 text-sm">
                                <input
                                  type="checkbox"
                                  checked={!!rule.tags?.includes(tag.name)}
                                  onChange={() => set(row.id, { tags: toggle(rule.tags, tag.name) })}
                                />
                                {tag.name}
                              </label>
                            ))
                          ) : (
                            <p className="text-xs text-muted-foreground">No tags yet.</p>
                          )}
                        </div>
                      </fieldset>
                      <fieldset className="space-y-1">
                        <legend className="text-xs text-muted-foreground">Segments</legend>
                        <div className="max-h-40 space-y-1 overflow-y-auto rounded-md border border-border p-2">
                          {segments.data?.data?.length ? (
                            segments.data.data.map((segment) => (
                              <label key={segment.id} className="flex items-center gap-2 text-sm">
                                <input
                                  type="checkbox"
                                  checked={!!rule.segments?.includes(segment.id)}
                                  onChange={() => set(row.id, { segments: toggle(rule.segments, segment.id) })}
                                />
                                {segment.name}
                              </label>
                            ))
                          ) : (
                            <p className="text-xs text-muted-foreground">No segments yet.</p>
                          )}
                        </div>
                      </fieldset>
                    </>
                  )}
                  {problem && (
                    <p role="alert" className="text-xs text-destructive sm:col-span-2">
                      {problem}
                    </p>
                  )}
                </div>
              )}
            </li>
          );
        })}
      </ol>
    </div>
  );
}
