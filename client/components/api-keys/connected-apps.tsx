"use client";
import { useCallback, useEffect, useState } from "react";
import { format } from "date-fns";
import { Bot } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useApi } from "@/hooks/use-api";
import { workspaceClassName } from "@/lib/workspace-styles";

type Grant = {
  id: string;
  clientName: string;
  createdAt: string;
  lastUsedAt: string | null;
};

// AI apps (Claude, ChatGPT, Grok…) this person connected by signing in.
export function ConnectedApps() {
  const { apiFetch, status } = useApi();
  const [apps, setApps] = useState<Grant[] | null>(null);
  const load = useCallback(async () => {
    const r = await apiFetch("oauth/grants");
    if (r.ok) setApps(await r.json());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status]);
  useEffect(() => {
    if (status === "authenticated") void load();
  }, [status, load]);
  if (!apps?.length) return null;
  async function disconnect(app: Grant) {
    const r = await apiFetch(`oauth/grants/${app.id}`, { method: "DELETE" });
    if (!r.ok) return void toast.error("Couldn’t disconnect the app");
    toast.success(`${app.clientName} is disconnected`);
    await load();
  }
  return (
    <section className={`${workspaceClassName("product-panel")} mt-6`}>
      <div className={workspaceClassName("panel-toolbar")}>
        <h2>Connected AI apps</h2>
      </div>
      <ul className="divide-y divide-border rounded-xl border border-border">
        {apps.map((app) => (
          <li
            key={app.id}
            className="flex items-center justify-between gap-4 p-4"
          >
            <span className="flex min-w-0 items-center gap-3">
              <Bot size={18} className="shrink-0 text-muted-foreground" />
              <span className="min-w-0">
                <span className="block truncate text-sm font-medium">
                  {app.clientName}
                </span>
                <span className="block text-xs text-muted-foreground">
                  Connected {format(new Date(app.createdAt), "MMM d, yyyy")}
                  {app.lastUsedAt
                    ? ` · last used ${format(new Date(app.lastUsedAt), "MMM d")}`
                    : ""}
                </span>
              </span>
            </span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void disconnect(app)}
            >
              Disconnect
            </Button>
          </li>
        ))}
      </ul>
    </section>
  );
}
