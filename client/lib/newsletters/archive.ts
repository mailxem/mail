import { notFound } from "next/navigation";

export type ArchiveIssue = { id: string; subject: string; sentAt: string };
export type Archive = {
  name: string;
  description: string;
  slug: string;
  subscribeForm: string;
  issues: ArchiveIssue[];
};
export type ArchiveEdition = {
  newsletter: string;
  slug: string;
  issue: ArchiveIssue;
  html: string;
};

/** Public archive endpoints live at the API's origin under /public, not /api/v1. */
export function archiveResponse(path: string) {
  const base = process.env.INTERNAL_API_URL?.replace(/\/api\/v1\/?$/, "");
  if (!base) throw new Error("INTERNAL_API_URL is not configured");
  return fetch(`${base}/public/newsletters/${path}`, {
    next: { revalidate: 300 },
  });
}

export async function archiveFetch<T>(path: string): Promise<T> {
  const res = await archiveResponse(path);
  if (res.status === 404) notFound();
  if (!res.ok) throw new Error(`Newsletter archive unavailable (${res.status})`);
  return res.json();
}

export const issueDate = (iso: string) =>
  new Date(iso).toLocaleDateString("en", {
    month: "short",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });

/** Plain text of an email for meta descriptions. */
export function htmlExcerpt(html: string, max = 160) {
  const text = html
    .replace(/<(style|script|head)[\s\S]*?<\/\1>/gi, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/\s+/g, " ")
    .trim();
  return text.length > max ? `${text.slice(0, max - 1)}…` : text;
}
