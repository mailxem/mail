// Newsletter websites. xem.email's own pages are static assets; any path
// they don't cover comes here, and xem.email/{slug} renders that
// newsletter's site from the public archive API. Any other host is a
// newsletter's own domain, connected through Cloudflare for SaaS.

import {
  back as backLink,
  bar,
  day,
  esc,
  foot,
  grainy,
  stylesheet,
  subscribe,
  template,
  type Issue,
} from "./templates";

interface Env {
  ASSETS: { fetch(request: Request | string): Promise<Response> };
  /** API origin; the archive lives under /public/newsletters. */
  API_URL: string;
}

type Archive = {
  name: string;
  description: string;
  slug: string;
  theme: string;
  logo: string;
  readers: number;
  cadence?: string;
  issues: Issue[];
};

const cadences: Record<string, string> = {
  DAILY: "Every day",
  WEEKLY: "Every week",
  MONTHLY: "Every month",
};
type Edition = {
  newsletter: string;
  slug: string;
  theme: string;
  logo: string;
  issue: Issue;
  html: string;
};
/** api is the API origin; the page's tracker posts views and clicks there. */
type Site = { slug: string; base: string; origin: string; api: string };

// Fades content in as it scrolls into view; [data-r] marks what moves.
const REVEAL = `const o=new IntersectionObserver(es=>es.forEach(e=>{if(e.isIntersecting){e.target.classList.add("in");o.unobserve(e.target)}}),{rootMargin:"0px 0px -6% 0px"});document.querySelectorAll("[data-r]").forEach(el=>o.observe(el))`;

// Counts views and clicks for the newsletter owner's analytics. No cookies:
// the visit id lives in sessionStorage. Skipped inside frames, so the owner's
// heatmap preview doesn't count itself.
const tracker = (site: Site) =>
  `(()=>{if(window.top!==window||!navigator.sendBeacon)return;const E=${JSON.stringify(`${site.api}/public/newsletters/${encodeURIComponent(site.slug)}/events`)},b=${JSON.stringify(site.base)};let s;try{s=sessionStorage.getItem("xem-v")}catch{}if(!s){s=(Math.random().toString(36).slice(2)+Date.now().toString(36)).replace(/[^a-z0-9]/g,"").slice(0,20);try{sessionStorage.setItem("xem-v",s)}catch{}}const d=document.documentElement,send=e=>{e.p=location.pathname.slice(b.length)||"/";e.s=s;e.w=innerWidth;e.h=d.scrollHeight;navigator.sendBeacon(E,JSON.stringify(e))};let r="";try{if(document.referrer&&new URL(document.referrer).host!==location.host)r=document.referrer}catch{}send({t:"view",r});addEventListener("click",ev=>{const a=ev.target instanceof Element?ev.target.closest("a"):null;send({t:"click",x:Math.min(1,Math.max(0,+(ev.pageX/d.scrollWidth).toFixed(4))),y:Math.round(ev.pageY),l:a?a.href:""})},{capture:true,passive:true})})()`;

const SLUG = /^[a-z0-9](?:[a-z0-9-]{1,46}[a-z0-9])$/;
const XEM_HOST =
  /^(?:www\.)?xem\.email$|\.workers\.dev$|^localhost$|^127\.0\.0\.1$/;

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
<style>${stylesheet(theme)}</style><script nonce="${nonce}">document.documentElement.classList.add("js")</script></head><body class="t-${esc(theme)}${grainy(theme) ? " grain" : ""}">${body.replace(/<script>/g, `<script nonce="${nonce}">`)}<script nonce="${nonce}">${REVEAL}</script><script nonce="${nonce}">${tracker(site)}</script></body></html>`;
  return new Response(html, {
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": "public, max-age=60",
      "Content-Security-Policy": `default-src 'none'; img-src https: data:; style-src 'unsafe-inline'; font-src 'self'; script-src 'nonce-${nonce}'; connect-src ${site.api}; form-action 'self'; frame-ancestors https://app.xem.email; base-uri 'none'`,
      "X-Content-Type-Options": "nosniff",
      "Referrer-Policy": "strict-origin-when-cross-origin",
    },
  });
}

const notes: Record<string, string> = {
  sent: "Almost there. Check your inbox for a link to confirm your subscription.",
  confirmed: "You're subscribed. The next issue will be in your inbox.",
  expired: "That confirmation link has expired. Subscribe again for a new one.",
  invalid: "We couldn't subscribe that address. Check it and try again.",
  busy: "Lots of people are signing up right now. Try again in a few minutes.",
};

// Previews may run without the configured vars; fall back to production.
const api = (env: Env, path: string, init?: RequestInit) =>
  fetch(
    `${(env.API_URL || "https://api.xem.email").replace(/\/$/, "")}/public/newsletters/${path}`,
    init,
  );

async function notFound(env: Env, request: Request) {
  const res = await env.ASSETS.fetch(new URL("/404", request.url).toString());
  return new Response(res.body, { status: 404, headers: res.headers });
}

async function archive(env: Env, request: Request, site: Site) {
  const res = await api(env, encodeURIComponent(site.slug));
  if (res.status === 404) return notFound(env, request);
  if (!res.ok)
    return new Response("This newsletter is unavailable right now.", {
      status: 503,
    });
  return (await res.json()) as Archive;
}

async function home(env: Env, request: Request, site: Site) {
  const a = await archive(env, request, site);
  if (a instanceof Response) return a;
  // Only fixed messages: nothing from the URL is shown on the page.
  const note =
    notes[new URL(request.url).searchParams.get("subscribe") || ""] || "";
  const body = template(a.theme).home({
    base: site.base,
    name: a.name,
    description: a.description,
    logo: a.logo,
    readers: a.readers,
    cadence: cadences[a.cadence || ""] || "",
    issues: a.issues,
    note,
  });
  return page(
    site,
    a.theme,
    {
      title: a.name,
      description: a.description || `Every issue of ${a.name}.`,
      path: "",
    },
    body,
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
    `${bar(site.base, e.newsletter, e.logo)}<main class="reader">${backLink(site.base)}<time class="meta mono" datetime="${esc(e.issue.sentAt)}">${day(e.issue.sentAt)}</time><h1>${esc(e.issue.subject)}</h1>
<iframe title="${esc(e.issue.subject)}" sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox" referrerpolicy="no-referrer" srcdoc="${esc(FRAME_POLICY + e.html)}"></iframe></main>
<section class="end"><h2>Get the next issue in your inbox.</h2>${subscribe(site.base, { readers: 0, cadence: "" })}</section>${foot(site.base, e.newsletter)}${FIT}`,
  );
}

// Opening the emailed link only shows a button: mail scanners open links on
// their own, and a subscription needs the reader's click.
async function confirmPage(env: Env, request: Request, site: Site) {
  const a = await archive(env, request, site);
  if (a instanceof Response) return a;
  const token = new URL(request.url).searchParams.get("token") || "";
  return page(
    site,
    a.theme,
    {
      title: `Confirm · ${a.name}`,
      description: `Confirm your subscription to ${a.name}.`,
      path: "/confirm",
    },
    `${bar(site.base, a.name, a.logo)}<main class="panel"><h1>One more click</h1><p>Confirm that you'd like to receive ${esc(a.name)}.</p><form method="post" action="${esc(site.base)}/confirm"><input type="hidden" name="token" value="${esc(token)}"><button class="btn">Yes, subscribe me</button></form></main>${foot(site.base, a.name, a.description)}`,
  );
}

async function feed(env: Env, request: Request, site: Site) {
  const a = await archive(env, request, site);
  if (a instanceof Response) return a;
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

async function subscribeTo(env: Env, request: Request, site: Site) {
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
    if (rest === "subscribe") return subscribeTo(env, request, site);
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
    try {
      return await serve(request, env);
    } catch (err) {
      // A failed API call shows a plain error page, never Cloudflare's crash page.
      console.error("newsletter site", request.url, err);
      return new Response(
        "This page is unavailable right now. Try again shortly.",
        {
          status: 503,
          headers: {
            "Content-Type": "text/plain; charset=utf-8",
            "Retry-After": "30",
          },
        },
      );
    }
  },
};

async function serve(request: Request, env: Env): Promise<Response> {
  {
    const url = new URL(request.url);
    const apiOrigin = new URL(env.API_URL || "https://api.xem.email").origin;
    const parts = url.pathname.split("/").filter(Boolean);
    if (!XEM_HOST.test(url.hostname)) {
      // Template fonts are served on customer domains too, same origin.
      if (parts[0] === "site-fonts") return env.ASSETS.fetch(request);
      const slug = await slugForDomain(env, url.hostname);
      if (!slug || parts.length > 1)
        return new Response("Not found", { status: 404 });
      return route(
        env,
        request,
        { slug, base: "", origin: url.origin, api: apiOrigin },
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
      { slug, base: `/${slug}`, origin: url.origin, api: apiOrigin },
      rest,
    );
  }
}
