"use client";
import { useEffect, useState } from "react";
import Image from "next/image";
import { Check, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useApi } from "@/hooks/use-api";

type Params = Record<string, string | undefined>;

export function Consent({
  params,
  email,
  canWrite,
}: {
  params: Params;
  email: string;
  canWrite: boolean;
}) {
  const { apiFetch, status } = useApi();
  const [app, setApp] = useState<{ name: string; host: string } | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const {
    client_id,
    redirect_uri,
    code_challenge,
    code_challenge_method,
    state,
  } = params;
  const valid =
    params.response_type === "code" &&
    !!client_id &&
    !!redirect_uri &&
    !!code_challenge &&
    code_challenge_method === "S256";

  useEffect(() => {
    if (!valid || status !== "authenticated") return;
    apiFetch(
      `oauth/clients/${encodeURIComponent(client_id!)}?redirect_uri=${encodeURIComponent(redirect_uri!)}`,
      { requireAuth: false },
    )
      .then(async (r) => {
        const body = await r.json();
        if (!r.ok)
          throw new Error(body.message || "This app isn’t registered.");
        setApp(body);
      })
      .catch((e) => setError((e as Error).message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [valid, status, client_id, redirect_uri]);

  async function decide(allow: boolean) {
    setBusy(true);
    setError("");
    if (!allow) {
      const back = new URL(redirect_uri!);
      back.searchParams.set("error", "access_denied");
      if (state) back.searchParams.set("state", state);
      window.location.assign(back.toString());
      return;
    }
    try {
      const r = await apiFetch("oauth/authorize", {
        method: "POST",
        body: JSON.stringify({
          client_id,
          redirect_uri,
          code_challenge,
          code_challenge_method,
          state,
        }),
      });
      const body = await r.json();
      if (!r.ok) throw new Error(body.message || "Could not connect the app.");
      window.location.assign(body.redirect);
    } catch (e) {
      setError((e as Error).message);
      setBusy(false);
    }
  }

  return (
    <main className="grid min-h-screen place-items-center bg-muted/30 p-5">
      <div className="w-full max-w-md rounded-2xl border border-border bg-card p-7 shadow-sm">
        <Image
          src="/apple-touch-icon.png"
          alt="Xem"
          width={40}
          height={40}
          className="rounded-lg"
        />
        {!valid ? (
          <>
            <h1 className="mt-6 text-xl font-medium">
              This link isn’t complete
            </h1>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              Start the connection again from your AI app.
            </p>
          </>
        ) : error && !app ? (
          <>
            <h1 className="mt-6 text-xl font-medium">
              We can’t connect this app
            </h1>
            <p
              role="alert"
              className="mt-2 text-sm leading-6 text-muted-foreground"
            >
              {error}
            </p>
          </>
        ) : !app ? (
          <p
            role="status"
            className="mt-6 flex items-center gap-2 text-sm text-muted-foreground"
          >
            <Loader2
              size={16}
              className="animate-spin motion-reduce:animate-none"
            />
            Checking the app…
          </p>
        ) : (
          <>
            <h1 className="mt-6 text-xl font-medium tracking-tight">
              Connect {app.name} to your Xem workspace?
            </h1>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              Signed in as {email}. {app.name} will work as you, through Xem’s
              MCP server.
            </p>
            <ul className="mt-5 space-y-2.5 text-sm">
              {[
                "Read your contacts, lists, campaigns and results",
                ...(canWrite
                  ? [
                      "Draft campaigns and newsletters, import contacts",
                      "Schedule and send, when you ask it to",
                    ]
                  : []),
              ].map((item) => (
                <li key={item} className="flex gap-2.5">
                  <Check
                    size={16}
                    className="mt-0.5 shrink-0 text-emerald-600"
                  />
                  {item}
                </li>
              ))}
            </ul>
            <p className="mt-5 text-xs leading-5 text-muted-foreground">
              It can’t see your password, API keys, billing or team settings.
              Disconnect it any time in Settings → API keys. After you approve,
              you’ll return to <span className="break-all">{app.host}</span>.
            </p>
            {error && (
              <p role="alert" className="mt-4 text-sm text-destructive">
                {error}
              </p>
            )}
            <div className="mt-6 flex gap-3">
              <Button
                className="flex-1"
                disabled={busy}
                onClick={() => void decide(true)}
              >
                {busy ? "Connecting…" : "Allow"}
              </Button>
              <Button
                variant="outline"
                disabled={busy}
                onClick={() => void decide(false)}
              >
                Cancel
              </Button>
            </div>
          </>
        )}
      </div>
    </main>
  );
}
