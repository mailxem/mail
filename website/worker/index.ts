// Newsletter websites. xem.email's own pages are static assets; any path
// they don't cover comes here, and xem.email/{slug} renders that
// newsletter's site from the public archive API.

interface Env {
  ASSETS: { fetch(request: Request | string): Promise<Response> };
  /** API origin; the archive lives under /public/newsletters. */
  API_URL: string;
}

type Issue = { id: string; subject: string; sentAt: string; cover?: string };
type Archive = {
  name: string;
  description: string;
  slug: string;
  theme: string;
  logo: string;
  readers: number;
  issues: Issue[];
};
type Edition = {
  newsletter: string;
  slug: string;
  theme: string;
  logo: string;
  issue: Issue;
  html: string;
};
type Site = { slug: string; base: string; origin: string };

const SLUG = /^[a-z0-9](?:[a-z0-9-]{1,46}[a-z0-9])$/;

const esc = (value: string) =>
  value.replace(
    /[<>&"']/g,
    (c) =>
      ({ "<": "&lt;", ">": "&gt;", "&": "&amp;", '"': "&quot;", "'": "&#39;" })[
        c
      ]!,
  );

const day = (iso: string) =>
  new Date(iso).toLocaleDateString("en", {
    month: "short",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });

const themes: Record<string, string> = {
  paper:
    "--bg:#fafaf9;--surface:#fff;--ink:#1c1917;--muted:#78716c;--line:#e7e5e4;--accent:#1c1917;--on-accent:#fff;--head:ui-serif,Georgia,'Times New Roman',serif",
  broadsheet:
    "--bg:#fbf7ef;--surface:#fffdf8;--ink:#1a1a1a;--muted:#6b6257;--line:#e6dccb;--accent:#b4232a;--on-accent:#fff;--head:'Times New Roman',Times,Georgia,serif",
  midnight:
    "--bg:#0b0f14;--surface:#121821;--ink:#e8edf1;--muted:#8d9aa6;--line:#223040;--accent:#4ade80;--on-accent:#06240f;--head:ui-sans-serif,system-ui,-apple-system,sans-serif",
};

const css = `*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--ink);font:16px/1.6 ui-sans-serif,system-ui,-apple-system,'Segoe UI',sans-serif;-webkit-font-smoothing:antialiased}
a{color:inherit}img{max-width:100%}
.bar{display:flex;align-items:center;justify-content:space-between;gap:16px;max-width:1100px;margin:0 auto;padding:18px 20px;border-bottom:1px solid var(--line)}
.brand{display:flex;align-items:center;gap:10px;text-decoration:none;font-weight:600}
.brand img{width:32px;height:32px;border-radius:8px;object-fit:cover}
.btn{display:inline-flex;align-items:center;justify-content:center;border:0;border-radius:999px;background:var(--accent);color:var(--on-accent);padding:10px 18px;font:600 15px/1 inherit;text-decoration:none;cursor:pointer;white-space:nowrap}
.hero{max-width:680px;margin:0 auto;padding:72px 20px 56px;text-align:center}
.hero .logo{width:76px;height:76px;border-radius:18px;object-fit:cover;box-shadow:0 8px 24px rgb(0 0 0/.12)}
h1{font-family:var(--head);font-weight:600;font-size:clamp(2.4rem,7vw,3.6rem);line-height:1.05;letter-spacing:-.02em;margin:22px 0 12px}
.lede{color:var(--muted);font-size:1.15rem;margin:0 auto 28px;max-width:560px}
form.sub{display:flex;gap:8px;max-width:460px;margin:0 auto;padding:6px;border:1px solid var(--line);border-radius:999px;background:var(--surface)}
form.sub input{flex:1;min-width:0;border:0;background:transparent;color:var(--ink);font:inherit;padding:8px 14px;outline-offset:4px}
.fine{color:var(--muted);font-size:14px;margin-top:12px}
.note{max-width:460px;margin:0 auto 18px;padding:12px 16px;border-radius:12px;background:var(--surface);border:1px solid var(--line);font-size:15px}
.issues{max-width:1100px;margin:0 auto;padding:0 20px 72px}
.issues h2{font-size:14px;font-weight:500;color:var(--muted);border-bottom:1px solid var(--line);padding-bottom:12px;margin:0 0 24px}
.grid{list-style:none;margin:0;padding:0;display:grid;gap:28px 24px;grid-template-columns:repeat(auto-fill,minmax(240px,1fr))}
.grid a{display:block;text-decoration:none}
.cover{aspect-ratio:16/10;border-radius:14px;overflow:hidden;background:var(--surface);border:1px solid var(--line);display:flex;align-items:center;justify-content:center;font:600 2.4rem var(--head);color:var(--accent)}
.cover img{width:100%;height:100%;object-fit:cover}
.grid time,.issue time{display:block;color:var(--muted);font-size:14px;margin-top:12px}
.grid span{display:block;font-weight:600;font-size:17px;line-height:1.35;margin-top:2px}
.grid a:hover span{text-decoration:underline}
.empty{color:var(--muted);text-align:center;padding:24px 0}
.issue{max-width:760px;margin:0 auto;padding:40px 20px 24px}
.issue h1{font-size:clamp(2rem,5vw,2.8rem);margin:8px 0 28px}
.back{color:var(--muted);text-decoration:none;font-size:14px}
iframe{display:block;width:100%;min-height:400px;border:1px solid var(--line);border-radius:14px;background:#fff}
.end{max-width:760px;margin:0 auto;padding:24px 20px 72px;text-align:center}
.end p{margin:0 0 16px;font-weight:600}
footer{max-width:1100px;margin:0 auto;padding:24px 20px 40px;border-top:1px solid var(--line);color:var(--muted);font-size:14px;display:flex;justify-content:space-between;gap:12px;flex-wrap:wrap}
@media (max-width:520px){form.sub{flex-direction:column;border-radius:18px}form.sub .btn{width:100%}.hero{padding-top:48px}}`;

function page(
  site: Site,
  theme: string,
  head: { title: string; description: string; path: string },
  body: string,
) {
  const nonce = crypto.randomUUID().replace(/-/g, "");
  const url = site.origin + site.base + head.path;
  const html = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>${esc(head.title)}</title><meta name="description" content="${esc(head.description)}">
<link rel="canonical" href="${esc(url)}"><link rel="alternate" type="application/rss+xml" href="${esc(site.base)}/feed.xml">
<meta property="og:title" content="${esc(head.title)}"><meta property="og:description" content="${esc(head.description)}"><meta property="og:url" content="${esc(url)}"><meta property="og:type" content="website">
<style>:root{${themes[theme] || themes.paper}}${css}</style></head><body>${body.replace(/<script>/g, `<script nonce="${nonce}">`)}</body></html>`;
  return new Response(html, {
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": "public, max-age=60",
      "Content-Security-Policy": `default-src 'none'; img-src https: data:; style-src 'unsafe-inline'; script-src 'nonce-${nonce}'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'`,
      "X-Content-Type-Options": "nosniff",
      "Referrer-Policy": "strict-origin-when-cross-origin",
    },
  });
}

function bar(site: Site, name: string, logo: string) {
  return `<header class="bar"><a class="brand" href="${esc(site.base || "/")}">${logo ? `<img src="${esc(logo)}" alt="">` : ""}${esc(name)}</a><a class="btn" href="${esc(site.base || "/")}#subscribe">Subscribe</a></header>`;
}

function subscribeForm(site: Site, readers: number) {
  const count =
    readers >= 10
      ? `Join ${readers.toLocaleString("en")} readers. Free, unsubscribe anytime.`
      : "Free, unsubscribe anytime.";
  return `<form class="sub" id="subscribe" method="post" action="${esc(site.base)}/subscribe"><input type="email" name="email" required autocomplete="email" placeholder="you@example.com" aria-label="Email address"><button class="btn">Subscribe</button></form><p class="fine">${count}</p>`;
}

const notes: Record<string, string> = {
  sent: "Almost there. Check your inbox for a link to confirm your subscription.",
  confirmed: "You're subscribed. The next issue will be in your inbox.",
  expired: "That confirmation link has expired. Subscribe again for a new one.",
  invalid: "We couldn't subscribe that address. Check it and try again.",
  busy: "Lots of people are signing up right now. Try again in a few minutes.",
};

function footer(site: Site) {
  return `<footer><span>Published with <a href="https://xem.email/?utm_source=newsletter-site">Xem</a></span><a href="${esc(site.base)}/feed.xml">RSS feed</a></footer>`;
}

async function api(env: Env, path: string, init?: RequestInit) {
  return fetch(
    `${env.API_URL.replace(/\/$/, "")}/public/newsletters/${path}`,
    init,
  );
}

async function notFound(env: Env, request: Request) {
  const res = await env.ASSETS.fetch(new URL("/404", request.url).toString());
  return new Response(res.body, { status: 404, headers: res.headers });
}

async function home(env: Env, request: Request, site: Site) {
  const res = await api(env, encodeURIComponent(site.slug));
  if (res.status === 404) return notFound(env, request);
  if (!res.ok)
    return new Response("This newsletter is unavailable right now.", {
      status: 503,
    });
  const a: Archive = await res.json();
  const params = new URL(request.url).searchParams;
  // Only fixed messages: nothing from the URL is shown on the page.
  const note = notes[params.get("subscribe") || ""] || "";
  const cards = a.issues
    .map(
      (i) =>
        `<li><a href="${esc(site.base)}/${esc(i.id)}"><div class="cover">${i.cover ? `<img src="${esc(i.cover)}" alt="" loading="lazy">` : esc(i.subject.trim().charAt(0).toUpperCase())}</div><time datetime="${esc(i.sentAt)}">${day(i.sentAt)}</time><span>${esc(i.subject)}</span></a></li>`,
    )
    .join("");
  return page(
    site,
    a.theme,
    {
      title: a.name,
      description: a.description || `Every issue of ${a.name}.`,
      path: "",
    },
    `${bar(site, a.name, a.logo)}<main><section class="hero">${a.logo ? `<img class="logo" src="${esc(a.logo)}" alt="">` : ""}<h1>${esc(a.name)}</h1>${a.description ? `<p class="lede">${esc(a.description)}</p>` : ""}${note ? `<p class="note" role="status">${esc(note)}</p>` : ""}${subscribeForm(site, a.readers)}</section>
<section class="issues"><h2>Latest issues</h2>${cards ? `<ol class="grid">${cards}</ol>` : `<p class="empty">The first issue is on its way.</p>`}</section></main>${footer(site)}`,
  );
}

// The email sits in a frame with scripts disabled, so its HTML can't touch
// this page. Same origin only so the page can size the frame to fit.
const FRAME_POLICY = `<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src https: data:; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'"><base target="_blank">`;
const FIT = `<script>const f=document.querySelector('iframe');const fit=()=>{const b=f.contentDocument&&f.contentDocument.body;if(!b)return;const h=Math.max(0,...Array.from(b.children,e=>e.getBoundingClientRect().bottom));if(h)f.style.height=Math.ceil(h)+16+'px'};f.addEventListener('load',fit);fit();</script>`;

async function issue(env: Env, request: Request, site: Site, id: string) {
  const res = await api(
    env,
    `${encodeURIComponent(site.slug)}/issues/${encodeURIComponent(id)}`,
  );
  if (res.status === 404) return notFound(env, request);
  if (!res.ok)
    return new Response("This issue is unavailable right now.", {
      status: 503,
    });
  const e: Edition = await res.json();
  return page(
    site,
    e.theme,
    {
      title: `${e.issue.subject} · ${e.newsletter}`,
      description: `An issue of ${e.newsletter}.`,
      path: `/${id}`,
    },
    `${bar(site, e.newsletter, e.logo)}<main><article class="issue"><a class="back" href="${esc(site.base || "/")}">← All issues</a><time datetime="${esc(e.issue.sentAt)}">${day(e.issue.sentAt)}</time><h1>${esc(e.issue.subject)}</h1>
<iframe title="${esc(e.issue.subject)}" sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox" referrerpolicy="no-referrer" srcdoc="${esc(FRAME_POLICY + e.html)}"></iframe></article>
<section class="end"><p>Get the next issue in your inbox.</p>${subscribeForm(site, 0)}</section></main>${footer(site)}${FIT}`,
  );
}

async function feed(env: Env, request: Request, site: Site) {
  const res = await api(env, encodeURIComponent(site.slug));
  if (res.status === 404) return notFound(env, request);
  if (!res.ok) return new Response("Feed unavailable", { status: 503 });
  const a: Archive = await res.json();
  const home = site.origin + site.base;
  const items = a.issues
    .slice(0, 50)
    .map(
      (i) =>
        `<item><title>${esc(i.subject)}</title><link>${home}/${esc(i.id)}</link><guid isPermaLink="true">${home}/${esc(i.id)}</guid><pubDate>${new Date(i.sentAt).toUTCString()}</pubDate></item>`,
    )
    .join("");
  return new Response(
    `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom"><channel><title>${esc(a.name)}</title><link>${home}</link><description>${esc(a.description || `Every issue of ${a.name}.`)}</description><language>en</language><atom:link href="${home}/feed.xml" rel="self" type="application/rss+xml"/>${items}</channel></rss>`,
    {
      headers: {
        "Content-Type": "application/rss+xml; charset=utf-8",
        "Cache-Control": "public, max-age=300",
      },
    },
  );
}

const back = (site: Site, query: string) =>
  new Response(null, {
    status: 303,
    headers: { Location: `${site.base || "/"}?${query}#subscribe` },
  });

async function subscribe(env: Env, request: Request, site: Site) {
  const form = await request.formData().catch(() => null);
  const email = String(form?.get("email") || "").slice(0, 320);
  const res = await api(env, `${encodeURIComponent(site.slug)}/subscribe`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  if (res.ok) return back(site, "subscribe=sent");
  return back(
    site,
    res.status === 429 ? "subscribe=busy" : "subscribe=invalid",
  );
}

// Opening the emailed link only shows a button: mail scanners open links on
// their own, and a subscription needs the reader's click.
async function confirmPage(env: Env, request: Request, site: Site) {
  const res = await api(env, encodeURIComponent(site.slug));
  if (res.status === 404) return notFound(env, request);
  if (!res.ok)
    return new Response("This newsletter is unavailable right now.", {
      status: 503,
    });
  const a: Archive = await res.json();
  const token = new URL(request.url).searchParams.get("token") || "";
  return page(
    site,
    a.theme,
    {
      title: `Confirm · ${a.name}`,
      description: `Confirm your subscription to ${a.name}.`,
      path: "/confirm",
    },
    `${bar(site, a.name, a.logo)}<main><section class="hero"><h1>One more click</h1><p class="lede">Confirm that you'd like to receive ${esc(a.name)}.</p><form method="post" action="${esc(site.base)}/confirm"><input type="hidden" name="token" value="${esc(token)}"><button class="btn">Yes, subscribe me</button></form></section></main>${footer(site)}`,
  );
}

async function confirm(env: Env, request: Request, site: Site) {
  const form = await request.formData().catch(() => null);
  const token = String(form?.get("token") || "").slice(0, 200);
  const res = await api(env, `${encodeURIComponent(site.slug)}/confirm`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token }),
  });
  if (res.ok) return back(site, "subscribe=confirmed");
  if (res.status === 404) return notFound(env, request);
  return back(site, "subscribe=expired");
}

// xem.email itself; every other host reaching this Worker is a newsletter's
// own domain, connected through Cloudflare for SaaS.
const XEM_HOST =
  /^(?:www\.)?xem\.email$|\.workers\.dev$|^localhost$|^127\.0\.0\.1$/;

async function slugForDomain(env: Env, host: string) {
  const res = await api(env, `by-domain/${encodeURIComponent(host)}`, {
    cf: { cacheTtl: 300, cacheEverything: true },
  } as RequestInit);
  if (!res.ok) return "";
  return ((await res.json()) as { slug?: string }).slug || "";
}

function route(
  env: Env,
  request: Request,
  site: Site,
  rest: string | undefined,
) {
  if (request.method === "POST") {
    if (rest === "subscribe") return subscribe(env, request, site);
    if (rest === "confirm") return confirm(env, request, site);
    return new Response("Method not allowed", { status: 405 });
  }
  if (request.method !== "GET" && request.method !== "HEAD")
    return new Response("Method not allowed", { status: 405 });
  if (!rest) return home(env, request, site);
  if (rest === "feed.xml") return feed(env, request, site);
  if (rest === "confirm") return confirmPage(env, request, site);
  return issue(env, request, site, rest);
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const parts = url.pathname.split("/").filter(Boolean);
    if (!XEM_HOST.test(url.hostname)) {
      const slug = await slugForDomain(env, url.hostname);
      if (!slug || parts.length > 1)
        return new Response("Not found", { status: 404 });
      return route(
        env,
        request,
        { slug, base: "", origin: url.origin },
        parts[0],
      );
    }
    // xem.email's own pages come first; newsletters fill the gaps.
    const asset = await env.ASSETS.fetch(request);
    if (asset.status !== 404) return asset;
    const [slug, rest, ...extra] = parts;
    if (!slug || !SLUG.test(slug) || extra.length)
      return notFound(env, request);
    return route(
      env,
      request,
      { slug, base: `/${slug}`, origin: url.origin },
      rest,
    );
  },
};
