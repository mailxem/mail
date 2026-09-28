"use client";

import { workspaceClassName } from "@/lib/workspace-styles";
import { SMTPProvider as SmtpContextProvider } from "@/app/providers/smtp-provider";
import { SMTPSettings } from "@/components/settings/smtp-settings";
import { useSearchParams } from "next/navigation";
import { useRouter } from "next/navigation";
export default function SMTPPage() {  
  const searchParams = useSearchParams();
  const isDialogOpen = searchParams.get("dialog") === "true";
  const router = useRouter();
  const setIsDialogOpen = (open: boolean) => {
    router.push(`/settings/smtp?dialog=${open ? "true" : "false"}`);
  };
  return (
    <div className="flex-1 space-y-4">
      <div className={workspaceClassName("workspace-page-body")}>
        <SmtpContextProvider>
          <SMTPSettings
            isDialogOpen={isDialogOpen}
            setIsDialogOpen={setIsDialogOpen}
          />
        </SmtpContextProvider>
      </div>
    </div>
  );
}
