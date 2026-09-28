"use client";

import { workspaceClassName } from "@/lib/workspace-styles";
import { IMAPProvider } from "@/app/providers/imap-provider";
import { IMAPSettings } from "@/components/settings/imap-settings";
import { useRouter, useSearchParams } from "next/navigation";

export default function SMTPPage() {
  const searchParams = useSearchParams();
  const isDialogOpen = searchParams.get("dialog") === "true";
  const router = useRouter();
  
  const setIsDialogOpen = (open: boolean) => {
    router.push(`/settings/imap?dialog=${open ? "true" : "false"}`);
  };

  return (
    <div className="flex-1 space-y-4">
      <div className={workspaceClassName("workspace-page-body")}>
        <IMAPProvider>
          <IMAPSettings
            isDialogOpen={isDialogOpen}
            setIsDialogOpen={setIsDialogOpen}
          />
        </IMAPProvider>
      </div>
    </div>
  );
}
