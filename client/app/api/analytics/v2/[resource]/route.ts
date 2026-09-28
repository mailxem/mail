import { auth } from "@/auth";
import { NextResponse } from "next/server";

const resources = new Set([
  "report",
  "email-overview",
  "breakdown",
  "people",
  "options",
]);
export async function GET(
  request: Request,
  context: { params: Promise<{ resource: string }> },
) {
  const session = await auth();
  if (!session?.user?.teamId || !session.accessToken)
    return NextResponse.json(
      { error: "Sign in to view analytics" },
      { status: 401 },
    );
  const { resource } = await context.params;
  if (!resources.has(resource))
    return NextResponse.json({ error: "Report not found" }, { status: 404 });
  const query = new URL(request.url).searchParams;
  if (query.has("teamId") && query.get("teamId") !== session.user.teamId)
    return NextResponse.json({ error: "Workspace mismatch" }, { status: 403 });
  query.set("teamId", session.user.teamId);
  const base = process.env.INTERNAL_API_URL?.replace(/\/$/, "");
  if (!base)
    return NextResponse.json(
      { error: "Analytics API is not configured" },
      { status: 503 },
    );
  const controller = new AbortController();
  const cancel = () => controller.abort();
  request.signal.addEventListener("abort", cancel, { once: true });
  const timeout = setTimeout(cancel, 25000);
  if (request.signal.aborted) cancel();
  try {
    const response = await fetch(`${base}/analytics/${resource}?${query}`, {
      headers: { Authorization: `Bearer ${session.accessToken}` },
      cache: "no-store",
      signal: controller.signal,
    });
    if (!response.ok) {
      const data = await response.json().catch(() => null);
      return NextResponse.json(
        {
          error:
            response.status === 404
              ? "Analytics is unavailable. The updated backend must be deployed."
              : response.status >= 500
                ? "Unable to load analytics. Please try again."
                : data?.message || "Unable to load analytics",
        },
        { status: response.status },
      );
    }
    return NextResponse.json(await response.json(), {
      headers: { "Cache-Control": "private, no-store" },
    });
  } catch {
    return NextResponse.json(
      { error: "Analytics did not respond. Please try again." },
      { status: 502 },
    );
  } finally {
    clearTimeout(timeout);
    request.signal.removeEventListener("abort", cancel);
  }
}
