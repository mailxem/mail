"use client";

import { TeamProvider } from "@/app/providers/team-provider";
import BillingPage from "./page";

export default function BillingWrapper() {
  return (
    <TeamProvider>
      <BillingPage />
    </TeamProvider>
  );
}
