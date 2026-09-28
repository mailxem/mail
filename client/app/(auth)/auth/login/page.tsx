"use client";
import { Button } from "@/components/ui/button";
import { useState, useEffect, use } from "react";
import Link from "next/link";
import { signIn, useSession } from "next-auth/react";
import { redirect, useRouter } from "next/navigation";
import { Eye, EyeOff } from "lucide-react";
import GoogleIcon from "@/components/icon/GoogleIcon";
import styles from "@/components/auth/auth.module.css";

export default function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string; code?: string }>;
}) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const router = useRouter();
  const { error, code } = use(searchParams);
  const { data: session } = useSession();
  useEffect(() => {
    if (error) {
      setMessage(
        error === "CredentialsSignin" && code === "credentials"
          ? "Invalid email or password. Please try again."
          : "We couldn’t sign you in. Please try again.",
      );
      router.replace("/auth/login");
    }
  }, [error, code, router]);
  if (session) redirect("/");
  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (busy) return;
    setBusy(true);
    setMessage("");
    try {
      await signIn("credentials", {
        email,
        password,
        redirect: true,
        callbackUrl: "/",
      });
    } catch {
      setMessage("We couldn’t sign you in. Please try again.");
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className={styles.page}>
      <header className={styles.heading}>
        <h1>Welcome to Xem.</h1>
        <p>Email built for better conversations.</p>
      </header>
      <form className={styles.form} onSubmit={submit}>
        {message && (
          <p role="alert" className={styles.error}>
            {message}
          </p>
        )}
        <div className={styles.field}>
          <label htmlFor="email">Email</label>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="username"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={busy}
          />
        </div>
        <div className={styles.field}>
          <div className={styles.labelRow}>
            <label htmlFor="password">Password</label>
            <Link href="/auth/forgot-password">Forgot password</Link>
          </div>
          <div className={styles.password}>
            <input
              id="password"
              name="password"
              type={showPassword ? "text" : "password"}
              autoComplete="current-password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={busy}
            />
            <button
              type="button"
              className={styles.reveal}
              aria-label={showPassword ? "Hide password" : "Show password"}
              aria-pressed={showPassword}
              onClick={() => setShowPassword(!showPassword)}
            >
              {showPassword ? <EyeOff size={15} /> : <Eye size={15} />}
            </button>
          </div>
        </div>
        <div className={styles.actions}>
          <Button type="submit" className={styles.primary} disabled={busy}>
            {busy ? "Signing in…" : "Log in"}
          </Button>
          <Button
            type="button"
            className={styles.google}
            disabled={busy}
            variant="outline"
            onClick={() => signIn("google")}
          >
            <GoogleIcon />
            Continue with Google
          </Button>
          <p className={styles.footnote}>
            New to Xem? <Link href="/auth/register">Create an account</Link> to
            get started.
          </p>
        </div>
      </form>
    </div>
  );
}
