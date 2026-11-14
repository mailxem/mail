import { AlertTriangle } from "lucide-react";
import { contentChecks } from "@/lib/email/content-checks";
import { cn } from "@/lib/utils";

/** Spam-filter and readability checks on the subject and HTML. Advisory: never blocks sending. */
export function ContentChecks({
  subject,
  html,
  className,
}: {
  subject?: string;
  html?: string;
  className?: string;
}) {
  const warnings = contentChecks({ subject, html });
  if (!warnings.length) return null;
  return (
    <div
      role="status"
      className={cn(
        "flex items-start gap-2 rounded-xl border border-amber-500/30 bg-amber-500/10 p-3 text-xs leading-relaxed",
        className,
      )}
    >
      <AlertTriangle size={14} className="mt-0.5 shrink-0 text-amber-600" />
      <ul className="space-y-2">
        {warnings.map((w) => (
          <li key={w.id}>
            <p className="font-medium">{w.title}</p>
            <p className="text-muted-foreground">{w.fix}</p>
          </li>
        ))}
      </ul>
    </div>
  );
}
