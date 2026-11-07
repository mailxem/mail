// Shared pieces of the newsletter website templates: types, helpers, fonts
// and the base stylesheet every template builds on.

export type Issue = {
  id: string;
  subject: string;
  sentAt: string;
  cover?: string;
};

/** What a home page needs. Strings are raw; templates escape them. */
export type Home = {
  base: string;
  name: string;
  description: string;
  logo: string;
  readers: number;
  /** "Every week", "Every month", … or "". */
  cadence: string;
  issues: Issue[];
  /** A fixed status message, or "". */
  note: string;
};

export const esc = (value: string) =>
  value.replace(
    /[<>&"']/g,
    (c) =>
      ({ "<": "&lt;", ">": "&gt;", "&": "&amp;", '"': "&quot;", "'": "&#39;" })[
        c
      ]!,
  );

export const day = (iso: string) =>
  new Date(iso).toLocaleDateString("en", {
    month: "short",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });

export const part = (iso: string, opts: Intl.DateTimeFormatOptions) =>
  new Date(iso).toLocaleDateString("en", { ...opts, timeZone: "UTC" });

const face = (family: string, file: string, weight: string, extra = "") =>
  `@font-face{font-family:"${family}";src:url(/site-fonts/${file}.woff2) format("woff2");font-weight:${weight};font-display:swap;${extra}}`;

export const fonts =
  face("Geist", "geist", "100 900") +
  face("Geist Mono", "geist-mono", "100 900") +
  face("Instrument Serif", "instrument-serif", "400") +
  face(
    "Instrument Serif",
    "instrument-serif-italic",
    "400",
    "font-style:italic",
  ) +
  face("Bricolage", "bricolage", "200 800", "font-stretch:75% 100%") +
  face("Newsreader", "newsreader-wght-normal", "200 800") +
  face("Newsreader", "newsreader-wght-italic", "200 800", "font-style:italic");

// Fine paper grain, drawn once as an SVG tile.
export const grain = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='180' height='180'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.85' numOctaves='3' stitchTiles='stitch'/%3E%3CfeColorMatrix values='0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 .07 0'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E")`;

export const arrow = `<svg width="14" height="14" viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8h10M9 4l4 4-4 4" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/></svg>`;

/** Shared structure; templates restyle it through variables and overrides. */
export const base = `${fonts}
*{box-sizing:border-box}html{-webkit-text-size-adjust:100%;scroll-behavior:smooth}
body{margin:0;background:var(--bg);color:var(--ink);font:400 16px/1.6 var(--sans);-webkit-font-smoothing:antialiased;text-rendering:optimizeLegibility;font-feature-settings:"ss01","cv11"}
body.grain{background-image:${grain}}
a{color:inherit;text-decoration:none}img{display:block;max-width:100%}
h1,h2,h3,p{margin:0}
:focus-visible{outline:1.5px solid var(--ink);outline-offset:3px;border-radius:2px}
::selection{background:var(--accent);color:var(--on-accent)}
.wrap{width:100%;max-width:1200px;margin:0 auto;padding:0 28px}
.mono{font:500 11.5px/1.4 var(--mono);letter-spacing:.06em;text-transform:uppercase}
.nav{display:flex;align-items:center;justify-content:space-between;gap:20px;height:72px}
.brand{display:flex;align-items:center;gap:10px;font:500 15px var(--sans);letter-spacing:-.01em}
.brand img{width:26px;height:26px;border-radius:7px;object-fit:cover}
.nav nav{display:flex;align-items:center;gap:26px;font-size:14px;color:var(--muted)}
.nav nav a:not(.btn):hover{color:var(--ink)}
.btn{display:inline-flex;align-items:center;justify-content:center;gap:8px;height:40px;border:0;border-radius:var(--radius,999px);background:var(--accent);color:var(--on-accent);padding:0 18px;font:500 14px/1 var(--sans);letter-spacing:-.005em;cursor:pointer;white-space:nowrap;transition:opacity .2s,transform .2s}
.btn:hover{opacity:.88}.btn:active{transform:scale(.98)}
.btn.sm{height:34px;padding:0 14px;font-size:13px}
form.sub{display:flex;align-items:center;gap:6px;width:100%;max-width:440px;padding:5px 5px 5px 18px;border:1px solid var(--line);border-radius:var(--radius,999px);background:var(--field,var(--surface));transition:border-color .2s,box-shadow .2s}
form.sub:focus-within{border-color:var(--ink);box-shadow:0 0 0 4px var(--ring,rgb(0 0 0/.05))}
form.sub input{flex:1;min-width:0;height:38px;border:0;background:transparent;color:var(--ink);font:400 15px var(--sans);outline:none}
form.sub input::placeholder{color:var(--muted)}
.fine{display:flex;align-items:center;gap:8px;color:var(--muted);margin-top:14px}
.dot{width:6px;height:6px;border-radius:50%;background:var(--live,#22c55e);box-shadow:0 0 0 3px color-mix(in srgb,var(--live,#22c55e) 22%,transparent)}
.note{max-width:440px;margin:0 0 16px;padding:12px 16px;border-radius:12px;background:var(--surface);border:1px solid var(--line);font-size:14.5px}
.empty{color:var(--muted);padding:40px 0}
.reader{max-width:760px;margin:0 auto;padding:40px 28px 0}
.reader .back{display:inline-flex;align-items:center;gap:6px;color:var(--muted)}
.reader .back svg{transform:rotate(180deg)}
.reader .meta{display:block;color:var(--muted);margin-top:44px}
.reader h1{font:var(--head-weight,400) clamp(2.2rem,5vw,3.4rem)/1.04 var(--head);letter-spacing:var(--head-track,-.02em);margin:14px 0 36px;text-wrap:balance}
.reader iframe{display:block;width:100%;min-height:420px;border:1px solid var(--line);border-radius:14px;background:#fff}
.end,.panel{max-width:760px;margin:0 auto;padding:72px 28px 96px;display:flex;flex-direction:column;align-items:center;text-align:center}
.end h2,.panel h1{font:var(--head-weight,400) clamp(1.9rem,4vw,2.6rem)/1.08 var(--head);letter-spacing:var(--head-track,-.02em);margin:0 0 24px;text-wrap:balance}
.panel{padding-top:120px;min-height:60vh}.panel p{color:var(--muted);margin:0 0 28px}
.foot{display:grid;grid-template-columns:1.4fr 1fr auto;gap:32px;padding-top:40px;padding-bottom:48px;border-top:1px solid var(--line);font-size:14px;color:var(--muted)}
.foot strong{display:block;color:var(--ink);font-weight:500;margin-bottom:6px}
.foot p{max-width:36ch}
.foot .links{display:flex;flex-direction:column;gap:6px}
.foot a:hover{color:var(--ink)}
.foot .made{align-self:end}
.js [data-r]{opacity:0;transform:translateY(18px);transition:opacity .9s cubic-bezier(.2,.7,.2,1),transform .9s cubic-bezier(.2,.7,.2,1);transition-delay:var(--d,0s)}
.js [data-r].in{opacity:1;transform:none}
@media (prefers-reduced-motion:reduce){.js [data-r]{opacity:1;transform:none;transition:none}html{scroll-behavior:auto}}
@media (max-width:720px){.nav nav a:not(.btn){display:none}.foot{grid-template-columns:1fr}form.sub{padding-left:14px}}`;

export const readers = (n: number) =>
  n >= 10 ? `${n.toLocaleString("en")} readers` : "";

export const subscribe = (
  base: string,
  h: Pick<Home, "readers" | "cadence">,
  button = "Subscribe",
) => {
  const meta =
    [h.cadence, readers(h.readers)].filter(Boolean).join(" · ") ||
    "Free, unsubscribe anytime";
  return `<form class="sub" id="subscribe" method="post" action="${esc(base)}/subscribe"><input type="email" name="email" required autocomplete="email" placeholder="you@example.com" aria-label="Email address"><button class="btn">${button} ${arrow}</button></form><p class="fine mono"><span class="dot"></span>${esc(meta)}</p>`;
};

export const bar = (base: string, name: string, logo: string) =>
  `<header class="nav wrap"><a class="brand" href="${esc(base || "/")}">${logo ? `<img src="${esc(logo)}" alt="">` : ""}<span>${esc(name)}</span></a><nav><a href="${esc(base || "/")}#issues">Archive</a><a href="${esc(base)}/feed.xml">RSS</a><a class="btn sm" href="${esc(base || "/")}#subscribe">Subscribe</a></nav></header>`;

export const foot = (base: string, name = "", description = "") =>
  `<footer class="foot wrap"><div>${name ? `<strong>${esc(name)}</strong>` : ""}${description ? `<p>${esc(description)}</p>` : ""}</div><div class="links"><a href="${esc(base || "/")}#subscribe">Subscribe</a><a href="${esc(base || "/")}#issues">Archive</a><a href="${esc(base)}/feed.xml">RSS feed</a></div><p class="made">Published with <a href="https://xem.email/?utm_source=newsletter-site" style="color:var(--ink)">Xem</a></p></footer>`;

export const back = (base: string) =>
  `<a class="back mono" href="${esc(base || "/")}">${arrow} All issues</a>`;

export const note = (h: Home) =>
  h.note ? `<p class="note" role="status">${esc(h.note)}</p>` : "";
export const href = (h: Home, i: Issue) => `${esc(h.base)}/${esc(i.id)}`;
export const no = (h: Home, n: number) =>
  String(h.issues.length - n).padStart(2, "0");
export const img = (i: Issue) =>
  i.cover ? `<img src="${esc(i.cover)}" alt="" loading="lazy">` : "";
export const headline = (h: Home) => esc(h.description || h.name);
export const count = (h: Home) =>
  `${h.issues.length} ${h.issues.length === 1 ? "issue" : "issues"}`;

/** A newsletter website template. */
export type Template = {
  label: string;
  home: (h: Home) => string;
  css: string;
  /** Paper-stock templates get a fine grain on the page. */
  grain?: boolean;
};

/** Big display type sized so the name fills the line, by its length. */
export const fit = (text: string, max: number, min = 3) =>
  `font-size:clamp(${min}rem,${(max / Math.max(text.length, 4)).toFixed(2)}vw,${(max * 0.12).toFixed(1)}rem)`;

/** Staggered reveal for the n-th item in a list. */
export const reveal = (n = 0, style = "") =>
  `data-r style="--d:${Math.min(n % 6, 5) * 0.06}s;${style}"`;
