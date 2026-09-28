"use client";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useApi } from "@/hooks/use-api";
import { AuthField } from "@/components/auth/auth-fields";
import styles from "@/components/auth/auth.module.css";
const schema = z.object({ email: z.string().email() });
export default function ForgotPasswordPage() {
  const form = useForm<z.infer<typeof schema>>({
    defaultValues: { email: "" },
    mode: "onBlur",
    resolver: zodResolver(schema),
  });
  const { apiFetch } = useApi();
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");
  const submit = async (data: z.infer<typeof schema>) => {
    setError("");
    try {
      const response = await apiFetch("auth/password-reset", {
        method: "POST",
        requireAuth: false,
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error();
      setSent(true);
    } catch {
      setError("We couldn’t request a reset link. Please try again.");
    }
  };
  return (
    <div className={styles.page}>
      <header className={styles.heading}>
        <h1>{sent ? "Check your inbox." : "Forgot your password?"}</h1>
        <p>
          {sent
            ? "Your next step is on its way."
            : "We’ll help you get back to your workspace."}
        </p>
      </header>
      {sent ? (
        <>
          <p role="status" className={styles.success}>
            If an account exists for this email, you’ll receive a password reset
            link. Check your spam folder too.
          </p>
          <div className={styles.actions}>
            <p className={styles.footnote}>
              <Link href="/auth/login">Back to log in</Link>
            </p>
          </div>
        </>
      ) : (
        <form className={styles.form} onSubmit={form.handleSubmit(submit)}>
          {error && (
            <p role="alert" className={styles.error}>
              {error}
            </p>
          )}
          <AuthField
            label="Email"
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            {...form.register("email")}
            error={form.formState.errors.email?.message}
            disabled={form.formState.isSubmitting}
          />
          <div className={styles.actions}>
            <Button
              type="submit"
              className={styles.primary}
              disabled={form.formState.isSubmitting}
            >
              {form.formState.isSubmitting
                ? "Sending link…"
                : "Send reset link"}
            </Button>
            <p className={styles.footnote}>
              Remember your password? <Link href="/auth/login">Log in</Link>.
            </p>
          </div>
        </form>
      )}
    </div>
  );
}
