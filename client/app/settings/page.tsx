import { workspaceClassName } from "@/lib/workspace-styles";
import { PageHeader } from "@/components/page-header";
import { SettingsTabs } from "@/components/settings/settings-tabs";

export default function SettingsPage() {
  return (
    <div className={workspaceClassName("workspace-page-body")}>
      <SettingsTabs />
    </div>
  );
}
