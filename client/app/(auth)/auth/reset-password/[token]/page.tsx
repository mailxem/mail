"use client";
import { Button } from "@/components/ui/button";
import { use, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import { AuthPasswordField } from "@/components/auth/auth-fields";
import styles from "@/components/auth/auth.module.css";
const schema = z
  .object({
    password: z.string().min(8, "Use at least 8 characters"),
    confirmPassword: z.string().min(8, "Use at least 8 characters"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don’t match",
    path: ["confirmPassword"],
  });
export default function ResetPasswordPage({
  params,
}: {
  params: Promise<{ token: string }>;
}) {
  const { token } = use(params);
  const router = useRouter();
  const { apiFetch } = useApi();
  const [error, setError] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    defaultValues: { password: "", confirmPassword: "" },
    mode: "onBlur",
    resolver: zodResolver(schema),
  });
  const submit = async (data: z.infer<typeof schema>) => {
    setError("");
    try {
      const response = await apiFetch("auth/password-reset/verify", {
        method: "POST",
        requireAuth: false,
        body: JSON.stringify({ code: token, new_password: data.password }),
      });
      if (!response.ok) throw new Error();
      toast.success("Password reset successful");
      router.push("/auth/login");
    } catch {
      setError(
        "We couldn’t reset your password. The link may have expired; request a new one below.",
      );
    }
  };
  return (
    <div className={styles.page}>
      <header className={styles.heading}>
        <h1>A fresh start.</h1>
        <p>Choose a new password for your Xem account.</p>
      </header>
      <form className={styles.form} onSubmit={form.handleSubmit(submit)}>
        {error && (
          <p role="alert" className={styles.error}>
            {error} <Link href="/auth/forgot-password">Request a new link</Link>
            .
          </p>
        )}
        <AuthPasswordField
          label="New password"
          autoComplete="new-password"
          placeholder="At least 8 characters"
          {...form.register("password")}
          error={form.formState.errors.password?.message}
          disabled={form.formState.isSubmitting}
        />
        <AuthPasswordField
          label="Confirm password"
          autoComplete="new-password"
          placeholder="Repeat your password"
          {...form.register("confirmPassword")}
          error={form.formState.errors.confirmPassword?.message}
          disabled={form.formState.isSubmitting}
        />
        <div className={styles.actions}>
          <Button
            type="submit"
            className={styles.primary}
            disabled={form.formState.isSubmitting}
          >
            {form.formState.isSubmitting
              ? "Updating password…"
              : "Reset password"}
          </Button>
          <p className={styles.footnote}>
            <Link href="/auth/login">Back to log in</Link>
          </p>
        </div>
      </form>
    </div>
  );
}
