// Almanac: a classic book page. Centered, roman numerals, small caps, fleurons.

import {
  esc,
  foot,
  href,
  note,
  part,
  reveal,
  subscribe,
  type Home,
  type Template,
} from "./base";

const roman = (n: number) => {
  let out = "";
  for (const [v, s] of [
    [1000, "M"],
    [900, "CM"],
    [500, "D"],
    [400, "CD"],
    [100, "C"],
    [90, "XC"],
    [50, "L"],
    [40, "XL"],
    [10, "X"],
    [9, "IX"],
    [5, "V"],
    [4, "IV"],
    [1, "I"],
  ] as const)
    while (n >= v) ((out += s), (n -= v));
  return out;
};

function home(h: Home) {
  const total = h.issues.length;
  const entries = h.issues
    .map(
      (i, n) =>
        `<li ${reveal(n)}><a href="${href(h, i)}"><span class="num">${roman(total - n)}.</span><span class="title">${esc(i.subject)}</span><span class="lead"></span><span class="date">${part(i.sentAt, { month: "long", day: "numeric", year: "numeric" })}</span></a></li>`,
    )
    .join("");
  return `<main class="page"><header class="title-page" ${reveal(0)}>${h.logo ? `<img class="seal" src="${esc(h.logo)}" alt="">` : `<p class="fleuron">❦</p>`}<p class="sc">${esc(h.cadence ? `Published ${h.cadence.toLowerCase()}` : "A newsletter")}</p><h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}<p class="rule">⁂</p>${note(h)}${subscribe(h.base, h)}</header>
<section class="contents" id="issues"><h2 class="sc">Contents</h2>${entries ? `<ol>${entries}</ol>` : `<p class="empty">The first chapter is being set in type.</p>`}<p class="fleuron end">❧</p></section></main>${foot(h.base, h.name, h.description)}`;
}

export const almanac: Template = {
  label: "Almanac",
  home,
  grain: true,
  css: `:root{--bg:#f4eee2;--surface:#faf6ec;--field:#faf6ec;--ink:#2a2118;--muted:#86786a;--line:rgb(42 33 24/.16);--accent:#2a2118;--on-accent:#f4eee2;--radius:2px;--live:#9a3b2a;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:Newsreader,Georgia,serif;--head-weight:400}
body{font:400 19px/1.65 Newsreader,Georgia,serif}
.page{max-width:720px;margin:0 auto;padding:96px 28px 40px;text-align:center}
.seal{width:72px;height:72px;border-radius:50%;object-fit:cover;margin:0 auto 28px;box-shadow:0 0 0 1px var(--line),0 0 0 6px var(--surface),0 0 0 7px var(--line)}
.fleuron{font-size:2rem;color:var(--live);margin-bottom:22px}
.sc{font:500 14px/1.4 Newsreader,Georgia,serif;font-variant:small-caps;letter-spacing:.14em;text-transform:lowercase;color:var(--muted)}
.title-page h1{font:300 clamp(3.2rem,9vw,6rem)/1 Newsreader,Georgia,serif;letter-spacing:-.025em;margin:18px 0 22px;font-variation-settings:"opsz" 72}
.lede{font-style:italic;font-size:1.35rem;color:var(--muted);max-width:32ch;margin:0 auto}
.rule{margin:40px 0 34px;font-size:1.4rem;color:var(--muted);letter-spacing:.4em}
.title-page form.sub{margin:0 auto;border-color:var(--ink)}
.title-page .fine{justify-content:center}
.contents{margin-top:120px;text-align:left}
.contents h2{text-align:center;margin-bottom:34px}
.contents ol{list-style:none;margin:0;padding:0}
.contents a{display:flex;align-items:baseline;gap:14px;padding:11px 0}
.num{min-width:3.4em;font-variant:small-caps;color:var(--muted);text-align:right}
.title{font-size:1.2rem;text-wrap:pretty}
.contents a:hover .title{font-style:italic;color:var(--live)}
.lead{flex:1;border-bottom:1.5px dotted var(--line);transform:translateY(-6px);min-width:24px}
.date{font-size:15px;color:var(--muted);white-space:nowrap;font-variant-numeric:oldstyle-nums}
.end{text-align:center;margin-top:48px}
.foot,.nav{font-family:var(--sans)}
@media (max-width:600px){.lead,.date{display:none}.num{min-width:2.6em}}`,
};
