"use client";
import { useState } from "react";
import { Check, Link2 } from "lucide-react";
export function ArticleActions() {
  const [status, setStatus] = useState("");
  return (
    <div className="flex flex-wrap items-center gap-3">
      <button
        onClick={async () => {
          try {
            await navigator.clipboard.writeText(
              window.location.href.split("#")[0],
            );
            setStatus("Link copied");
          } catch {
            setStatus(
              "Copy the address from your browser to share this article.",
            );
          }
        }}
        className="inline-flex min-h-10 items-center gap-2 rounded-full border border-ink/20 px-4 text-xs hover:bg-lavender"
      >
        {status === "Link copied" ? <Check size={14} /> : <Link2 size={14} />}
        Copy article link
      </button>
      <span role="status" className="text-xs text-muted">
        {status}
      </span>
    </div>
  );
}
