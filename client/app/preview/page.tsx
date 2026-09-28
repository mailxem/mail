import { notFound } from "next/navigation";
import { WorkspacePreview } from "./workspace-preview";
// The preview is never exposed by a production build and never connects to a real account.
export default function PreviewPage() {
  if (process.env.NODE_ENV !== "development") notFound();
  return <WorkspacePreview />;
}
