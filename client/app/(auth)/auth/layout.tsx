"use client";
import Intercom from "@intercom/messenger-js-sdk";
import { useEffect } from "react";
import { AuthShell } from "@/components/auth/auth-shell";
export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  useEffect(() => {
    Intercom({ app_id: "ts43f4k1" });
  }, []);
  return <AuthShell>{children}</AuthShell>;
}
