import { archiveResponse, type Archive } from "@/lib/newsletters/archive";

const xml = (value: string) =>
  value.replace(
    /[<>&"']/g,
    (c) =>
      ({ "<": "&lt;", ">": "&gt;", "&": "&amp;", '"': "&quot;", "'": "&apos;" })[c]!,
  );

/** RSS 2.0 feed of a published newsletter's archive, newest first. */
export async function GET(
  request: Request,
  { params }: { params: Promise<{ slug: string }> },
) {
  const { slug } = await params;
  const res = await archiveResponse(encodeURIComponent(slug));
  if (res.status === 404) return new Response("Not found", { status: 404 });
  if (!res.ok) return new Response("Archive unavailable", { status: 503 });
  const a: Archive = await res.json();
  const origin = (process.env.NEXT_PUBLIC_APP_URL || new URL(request.url).origin).replace(/\/$/, "");
  const home = `${origin}/n/${encodeURIComponent(a.slug)}`;
  const items = a.issues
    .slice(0, 50)
    .map(
      (i) =>
        `<item><title>${xml(i.subject)}</title><link>${home}/${i.id}</link><guid isPermaLink="true">${home}/${i.id}</guid><pubDate>${new Date(i.sentAt).toUTCString()}</pubDate></item>`,
    )
    .join("");
  return new Response(
    `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom"><channel><title>${xml(a.name)}</title><link>${home}</link><description>${xml(a.description || `Every issue of ${a.name}.`)}</description><language>en</language><atom:link href="${home}/feed.xml" rel="self" type="application/rss+xml"/>${items}</channel></rss>`,
    {
      headers: {
        "Content-Type": "application/rss+xml; charset=utf-8",
        "Cache-Control": "public, max-age=300",
      },
    },
  );
}
