"use client";

import { UserRound } from "lucide-react";
import { useMarketingQuery } from "@/lib/marketing/api";
import { CONTACT_FIELDS, usedFields } from "@/lib/email/fields";
import { cn } from "@/lib/utils";

type EmptyFields = { total: number; empty: Record<string, number> };

const label = (key: string) =>
  (
    CONTACT_FIELDS.find((f) => f.key === key)?.label ?? key.replace(/_/g, " ")
  ).toLowerCase();

/** How many readers of the list lack each field the email uses, and what they see instead. */
export function FieldFallbacks({
  listId,
  subject,
  html,
  data,
  className,
}: {
  listId?: string;
  subject?: string;
  html?: string;
  data?: Record<string, string>;
  className?: string;
}) {
  const counts = useMarketingQuery<EmptyFields>(
    `marketing/lists/${listId}/empty-fields`,
    !!listId,
  ).data;
  if (!counts?.total) return null;
  // A value set on the campaign goes to every reader.
  const gaps = usedFields(subject, html).filter(
    (f) => counts.empty[f.key] && !data?.[f.key]?.trim(),
  );
  if (!gaps.length) return null;
  const total = counts.total.toLocaleString();
  return (
    <div
      role="status"
      className={cn(
        "flex items-start gap-2 rounded-xl border border-border bg-muted/20 p-3 text-xs leading-relaxed",
        className,
      )}
    >
      <UserRound size={14} className="mt-0.5 shrink-0 text-muted-foreground" />
      <ul className="space-y-1">
        {gaps.map(({ key, fallback }) => (
          <li key={key}>
            {counts.empty[key].toLocaleString()} of {total} have no {label(key)}{" "}
            {fallback ? (
              `and will see “${fallback}”.`
            ) : fallback === "" ? (
              "and will see a blank."
            ) : (
              <>
                and will see a blank. Add a fallback like{" "}
                <span className="font-mono">{`{{${key}|there}}`}</span>.
              </>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
