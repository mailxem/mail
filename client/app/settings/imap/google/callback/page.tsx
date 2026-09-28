"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useMarketing } from "@/lib/marketing/api";
import { Button } from "@/components/ui/button";

export default function GoogleMailboxCallback() {
  const { request, ready, refresh } = useMarketing();
  const started = useRef(false);
  const [message, setMessage] = useState(
    "Completing your Google mailbox connection…",
  );
  useEffect(() => {
    if (!ready || started.current) return;
    started.current = true;
    const params = new URLSearchParams(window.location.search);
    const code = params.get("code"),
      state = params.get("state");
    // Remove the one-time authorization code from browser history immediately.
    window.history.replaceState(null, "", window.location.pathname);
    if (params.has("error") || !code || !state) {
      setMessage(
        "Google authorization was canceled or incomplete. Return to settings to start again.",
      );
      return;
    }
    request("mail-connections/google/complete", "POST", { code, state })
      .then(async () => {
        await refresh();
        setMessage(
          "Google mailbox connected. You can now choose it in your inbox and composer.",
        );
      })
      .catch((error) => setMessage((error as Error).message));
  }, [ready, request, refresh]);
  return (
    <div className="max-w-xl p-6 space-y-4">
      <h1 className="text-xl font-semibold">Google mailbox</h1>
      <p role="status" className="text-sm text-muted-foreground">
        {message}
      </p>
      <Button asChild variant="outline">
        <Link href="/settings/imap">Back to mailbox settings</Link>
      </Button>
    </div>
  );
}
