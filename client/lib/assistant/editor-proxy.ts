import { z } from "zod";
import {
  authenticate,
  AssistantError,
  backendEndpoint,
  errorResponse,
} from "./server";
import { readJSON } from "./http";

// This is a fixed-purpose proxy, never a client-selected upstream URL.
export async function editorProxy(
  request: Request,
  operation: "draft" | "save",
  id?: string,
) {
  const controller = new AbortController();
  const abort = () => controller.abort();
  request.signal.addEventListener("abort", abort, { once: true });
  if (request.signal.aborted) abort();
  const timeout = setTimeout(abort, operation === "draft" ? 100_000 : 20_000);
  try {
    const input = await readJSON(
      request,
      operation === "draft" ? 100_000 : 2_000_000,
    );
    const scope = await authenticate();
    if (input.teamId !== scope.teamId)
      throw new AssistantError(
        409,
        "Your workspace changed. Reopen the template before continuing.",
      );
    if (id && !z.string().uuid().safeParse(id).success)
      throw new AssistantError(400, "Invalid template.");
    const { teamId: _, ...body } = input;
    const path =
      operation === "draft"
        ? "marketing/email-draft"
        : `marketing/templates${id ? `/${id}` : ""}`;
    const response = await fetch(`${backendEndpoint()}/${path}`, {
      method: operation === "draft" || !id ? "POST" : "PUT",
      headers: {
        Authorization: `Bearer ${scope.accessToken}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
      cache: "no-store",
      redirect: "error",
      signal: controller.signal,
    });
    if (!response.ok) {
      const messages: Record<number, string> = {
        400: "Check the template name, subject, category, and design, then try again.",
        401: "Please sign in again. Your draft is still here.",
        403: "You do not have permission to edit templates in this workspace.",
        404: "This template is no longer available. Your draft is still here.",
        409: "This template changed in another session. Reopen its latest version before saving your revision.",
        413: "This template is too large to save. Reduce its content and try again.",
        429: "Xem is receiving too many requests. Please wait a moment and try again.",
      };
      throw new AssistantError(
        messages[response.status] ? response.status : 503,
        messages[response.status] ||
          (operation === "draft"
            ? "The design service couldn’t finish this revision. Your template is unchanged. Try again shortly."
            : "Saving could not be confirmed. Your draft is still here; check the saved template before retrying."),
      );
    }
    return Response.json(await response.json(), {
      status: response.status,
      headers: { "Cache-Control": "no-store" },
    });
  } catch (error) {
    if (
      error instanceof Error &&
      ["TimeoutError", "AbortError", "TypeError"].includes(error.name)
    )
      return Response.json(
        {
          error:
            operation === "draft"
              ? "The design service did not respond. Try again; your template is unchanged."
              : "Saving could not be confirmed. Your draft is still here; check the saved template before retrying.",
        },
        { status: 503 },
      );
    return errorResponse(error);
  } finally {
    clearTimeout(timeout);
    request.signal.removeEventListener("abort", abort);
  }
}
