import { auth } from "@/auth";
import { redirect } from "next/navigation";
import { Consent } from "./consent";

// The authorization endpoint for MCP clients (Claude, ChatGPT, Grok…): a
// signed-in person approves the app for their workspace.
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | undefined>>;
}) {
  const params = await searchParams;
  const session = await auth();
  if (!session?.user) {
    const query = new URLSearchParams(
      Object.entries(params).filter((e): e is [string, string] => !!e[1]),
    );
    redirect(
      `/auth/login?next=${encodeURIComponent(`/oauth/authorize?${query}`)}`,
    );
  }
  return (
    <Consent
      params={params}
      email={session.user.email ?? ""}
      canWrite={["ADMIN", "SUPER_ADMIN"].includes(session.user.role || "")}
    />
  );
}
