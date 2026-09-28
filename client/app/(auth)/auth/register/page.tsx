"use client";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { signIn, useSession } from "next-auth/react";
import { redirect } from "next/navigation";
import Link from "next/link";
import { toast } from "sonner";
import GoogleIcon from "@/components/icon/GoogleIcon";
import { useApi } from "@/hooks/use-api";
import { AuthField, AuthPasswordField } from "@/components/auth/auth-fields";
import styles from "@/components/auth/auth.module.css";
const schema = z
  .object({
    email: z.string().email(),
    first_name: z.string().min(1, "Enter your first name"),
    last_name: z.string().min(1, "Enter your last name"),
    password: z.string().min(8, "Use at least 8 characters"),
    confirmPassword: z.string().min(8, "Use at least 8 characters"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don’t match",
    path: ["confirmPassword"],
  });
export default function RegisterPage() {
  const form = useForm<z.infer<typeof schema>>({
    defaultValues: {
      email: "",
      first_name: "",
      last_name: "",
      password: "",
      confirmPassword: "",
    },
    mode: "onBlur",
    resolver: zodResolver(schema),
  });
  const { data: session } = useSession();
  const { apiFetch } = useApi();
  const [error, setError] = useState("");
  if (session) redirect("/");
  const submit = async (data: z.infer<typeof schema>) => {
    setError("");
    try {
      const response = await apiFetch("auth/register", {
        method: "POST",
        requireAuth: false,
        body: JSON.stringify(data),
      });
      if (!response.ok)
        throw new Error(
          "Registration failed. Please check your details and try again.",
        );
      toast.success("Registration successful", {
        description: "Please check your email for verification",
      });
      await signIn("credentials", {
        email: data.email,
        password: data.password,
        redirect: true,
        callbackUrl: "/onboarding",
      });
    } catch {
      setError(
        "We couldn’t create your account. Please check your details and try again.",
      );
    }
  };
  const errors = form.formState.errors;
  const busy = form.formState.isSubmitting;
  return (
    <div className={styles.page}>
      <header className={styles.heading}>
        <h1>Your next chapter.</h1>
        <p>Create your Xem account.</p>
      </header>
      <form className={styles.form} onSubmit={form.handleSubmit(submit)}>
        {error && (
          <p className={styles.error} role="alert">
            {error}
          </p>
        )}
        <div className={styles.row}>
          <AuthField
            label="First name"
            autoComplete="given-name"
            placeholder="Alex"
            {...form.register("first_name")}
            error={errors.first_name?.message}
            disabled={busy}
          />
          <AuthField
            label="Last name"
            autoComplete="family-name"
            placeholder="Morgan"
            {...form.register("last_name")}
            error={errors.last_name?.message}
            disabled={busy}
          />
        </div>
        <AuthField
          label="Email"
          type="email"
          autoComplete="username"
          placeholder="you@example.com"
          {...form.register("email")}
          error={errors.email?.message}
          disabled={busy}
        />
        <AuthPasswordField
          label="Password"
          autoComplete="new-password"
          placeholder="At least 8 characters"
          {...form.register("password")}
          error={errors.password?.message}
          disabled={busy}
        />
        <AuthPasswordField
          label="Confirm password"
          autoComplete="new-password"
          placeholder="Repeat your password"
          {...form.register("confirmPassword")}
          error={errors.confirmPassword?.message}
          disabled={busy}
        />
        <div className={styles.actions}>
          <Button type="submit" className={styles.primary} disabled={busy}>
            {busy ? "Creating account…" : "Create account"}
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
            Already have an account? <Link href="/auth/login">Log in</Link>.
          </p>
        </div>
      </form>
    </div>
  );
}
