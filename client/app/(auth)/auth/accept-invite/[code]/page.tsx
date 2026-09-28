"use client";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { useApi } from "@/hooks/use-api";
import { useQuery } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { QueryState } from "@/components/marketing/shared";
import { toast } from "sonner";
import styles from "@/components/auth/auth.module.css";
import { AuthPasswordField } from "@/components/auth/auth-fields";
interface Invite {
  valid: boolean;
  expires_at: string;
  name: string;
  team_name: string;
}
export default function AcceptInvitePage() {
  const { code } = useParams<{ code: string }>();
  const router = useRouter();
  const { apiFetch } = useApi();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [busy, setBusy] = useState(false);
  const query = useQuery<Invite>({
    queryKey: ["invite", code],
    enabled: !!code,
    retry: 1,
    queryFn: async ({ signal }) => {
      const response = await apiFetch(
        `auth/invite/${encodeURIComponent(code)}`,
        { requireAuth: false, signal },
      );
      if (!response.ok)
        throw new Error("This invitation is unavailable or has expired.");
      return response.json();
    },
  });
  const accept = async (event: React.FormEvent) => {
    event.preventDefault();
    if (password !== confirm) {
      toast.error("Passwords do not match");
      return;
    }
    setBusy(true);
    try {
      const response = await apiFetch(
        `auth/accept/${encodeURIComponent(code)}`,
        {
          method: "POST",
          requireAuth: false,
          body: JSON.stringify({ password }),
        },
      );
      const result = await response.json();
      if (!response.ok)
        throw new Error(result.error || "Unable to accept invitation");
      toast.success("Invitation accepted. Sign in to your workspace.");
      router.push("/auth/login");
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setBusy(false);
    }
  };
  if (query.isPending || query.error)
    return (
      <div className={styles.page}>
        <header className={styles.heading}>
          <h1>Your invitation.</h1>
          <p>Join your team on Xem.</p>
        </header>
        <QueryState
          loading={query.isPending}
          error={query.error}
          retry={() => void query.refetch()}
        />
        <div className={styles.actions}>
          <Link href="/auth/login">Back to log in</Link>
        </div>
      </div>
    );
  if (!query.data.valid || Date.parse(query.data.expires_at) <= Date.now())
    return (
      <div className={styles.page}>
        <header className={styles.heading}>
          <h1>Invitation expired.</h1>
          <p>Ask your workspace administrator for a new invitation.</p>
        </header>
        <div className={styles.actions}>
          <Link href="/auth/login">Back to log in</Link>
        </div>
      </div>
    );
  return (
    <div className={styles.page}>
      <header className={styles.heading}>
        <h1>Join {query.data.team_name}.</h1>
        <p>
          Welcome, {query.data.name}. Set a password to accept your invitation.
        </p>
      </header>
      <form onSubmit={accept} className={styles.form}>
        <AuthPasswordField
          label="Password"
          id="invite-password"
          autoComplete="new-password"
          required
          minLength={8}
          maxLength={72}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          disabled={busy}
        />
        <AuthPasswordField
          label="Confirm password"
          id="invite-confirm"
          autoComplete="new-password"
          required
          minLength={8}
          maxLength={72}
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          disabled={busy}
        />
        <div className={styles.actions}>
          <Button className={styles.primary} type="submit" disabled={busy}>
            {busy ? "Joining workspace…" : "Accept invitation"}
          </Button>
          <p className={styles.footnote}>
            <Link href="/auth/login">Back to log in</Link>
          </p>
        </div>
      </form>
    </div>
  );
}
