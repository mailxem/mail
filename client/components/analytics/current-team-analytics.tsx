"use client";
import Link from "next/link";
import { Stats } from "@/components/stats";
export function CurrentTeamAnalytics() {
  return (
    <div className="space-y-4">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <Stats />
      </div>
      <Link
        href="/analytics"
        className="text-sm text-violet-700 hover:underline"
      >
        Open workspace analytics ↗
      </Link>
    </div>
  );
}
