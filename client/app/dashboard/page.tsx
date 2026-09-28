import { Suspense } from "react";
import { DashboardHome } from "@/components/analytics/dashboard-home";

export default function Home() {
  return (
    <Suspense fallback={<p role="status">Loading dashboard…</p>}>
      <DashboardHome />
    </Suspense>
  );
}
