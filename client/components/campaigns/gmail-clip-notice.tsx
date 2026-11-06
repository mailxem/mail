import { AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  formatKB,
  gmailClipState,
  sentEmailBytes,
} from "@/lib/email/size";

const HINT =
  "This usually comes from styles pasted from another tool, images embedded as data URLs, or very large tables. Host images and trim unused styles to shrink it.";

/** Email size against Gmail's clipping limit. Advisory: never blocks sending. */
export function GmailClipNotice({
  html,
  className,
}: {
  html?: string;
  className?: string;
}) {
  if (!html) return null;
  const bytes = sentEmailBytes(html);
  const state = gmailClipState(bytes);
  const size = `About ${formatKB(bytes)} when sent`;
  if (state === "ok")
    return (
      <p
        data-state={state}
        className={cn("text-xs text-muted-foreground", className)}
      >
        {size}
      </p>
    );
  return (
    <div
      role="status"
      data-state={state}
      className={cn(
        "flex items-start gap-2 rounded-xl border p-3 text-xs leading-relaxed",
        state === "clipped"
          ? "border-destructive/40 bg-destructive/10"
          : "border-amber-500/30 bg-amber-500/10",
        className,
      )}
    >
      <AlertTriangle
        size={14}
        className={cn(
          "mt-0.5 shrink-0",
          state === "clipped" ? "text-destructive" : "text-amber-600",
        )}
      />
      <div>
        <p className="font-medium">
          {state === "clipped"
            ? "Gmail will clip this email and hide your unsubscribe link"
            : "Close to Gmail's 102 KB limit"}
        </p>
        <p className="text-muted-foreground">
          {size}. {HINT}
        </p>
      </div>
    </div>
  );
}
