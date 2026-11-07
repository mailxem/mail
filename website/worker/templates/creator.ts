// Creator: a personal page. Big avatar, soft color wash, warm cards.

import {
  arrow,
  bar,
  count,
  day,
  esc,
  foot,
  href,
  no,
  note,
  reveal,
  subscribe,
  type Home,
  type Template,
} from "./base";

function home(h: Home) {
  const [first, ...rest] = h.issues;
  const cards = rest
    .map(
      (i, n) =>
        `<a class="card" href="${href(h, i)}" ${reveal(n)}>${i.cover ? `<div class="art"><img src="${esc(i.cover)}" alt="" loading="lazy"></div>` : ""}<div class="txt"><p class="mono">${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></div><span class="go">${arrow}</span></a>`,
    )
    .join("");
  const stats = [
    h.readers >= 10 ? `<b>${h.readers.toLocaleString("en")}</b> readers` : "",
    `<b>${h.issues.length}</b> ${h.issues.length === 1 ? "issue" : "issues"}`,
    h.cadence
      ? `<b>${esc({ "Every day": "Daily", "Every week": "Weekly", "Every month": "Monthly" }[h.cadence] || h.cadence)}</b>`
      : "",
  ]
    .filter(Boolean)
    .map((s) => `<li>${s}</li>`)
    .join("");
  return `<div class="wash" aria-hidden="true"></div>${bar(h.base, h.name, h.logo)}<main><section class="hero wrap" ${reveal(0)}><div class="avatar">${h.logo ? `<img src="${esc(h.logo)}" alt="">` : `<span>${esc(h.name.charAt(0))}</span>`}</div><h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}<ul class="stats">${stats}</ul>${note(h)}${subscribe(h.base, h)}</section>
<section class="wrap feed" id="issues">${
    first
      ? `<a class="lead" href="${href(h, first)}" ${reveal(1)}>${first.cover ? `<div class="art"><img src="${esc(first.cover)}" alt=""></div>` : ""}<div class="txt"><span class="chip mono">New · No. ${no(h, 0)}</span><h2>${esc(first.subject)}</h2><p class="mono">${day(first.sentAt)}</p><span class="read">Read it ${arrow}</span></div></a>`
      : `<p class="empty">The first issue is on its way.</p>`
  }${cards ? `<div class="head"><h2>More issues</h2><span class="mono">${count(h)}</span></div><div class="cards">${cards}</div>` : ""}</section></main>${foot(h.base, h.name, h.description)}`;
}

export const creator: Template = {
  label: "Creator",
  home,
  css: `:root{--bg:#fff8f1;--surface:#fff;--field:#fff;--ink:#2b1d14;--muted:#8c7566;--line:rgb(43 29 20/.1);--accent:#2b1d14;--on-accent:#fff;--live:#f26b3a;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:"Instrument Serif",Georgia,serif}
.wash{position:absolute;inset:0 0 auto;height:620px;z-index:-1;background:radial-gradient(40% 60% at 20% 10%,#ffd2b8 0,transparent 70%),radial-gradient(40% 55% at 80% 0,#ffc9de 0,transparent 70%),radial-gradient(50% 50% at 50% 40%,#fff1c9 0,transparent 70%);filter:blur(10px)}
.hero{display:flex;flex-direction:column;align-items:center;text-align:center;padding-top:64px;padding-bottom:80px}
.avatar{width:112px;height:112px;border-radius:50%;padding:5px;background:linear-gradient(135deg,#f26b3a,#e8458b);box-shadow:0 18px 40px rgb(232 69 139/.25)}
.avatar img,.avatar span{width:100%;height:100%;border-radius:50%;object-fit:cover;background:#fff;border:4px solid #fff;display:grid;place-items:center;font:400 3rem var(--head)}
.hero h1{font:400 clamp(3rem,7vw,5rem)/1 var(--head);letter-spacing:-.025em;margin:28px 0 12px}
.lede{font-size:1.2rem;color:var(--muted);max-width:40ch}
.hero .fine{display:none}
.stats{list-style:none;display:flex;gap:10px;flex-wrap:wrap;justify-content:center;padding:0;margin:26px 0 30px}
.stats li{padding:7px 14px;border-radius:999px;background:rgb(255 255 255/.7);border:1px solid var(--line);font-size:14px;color:var(--muted);backdrop-filter:blur(6px)}
.stats b{color:var(--ink);font-weight:600;text-transform:capitalize}
.feed{max-width:960px;padding-bottom:112px}
.lead{display:grid;grid-template-columns:1.1fr 1fr;align-items:center;gap:36px;padding:14px;border-radius:30px;background:var(--surface);box-shadow:0 2px 4px rgb(43 29 20/.04),0 24px 60px rgb(43 29 20/.08)}
.lead .art{aspect-ratio:4/3;border-radius:20px;overflow:hidden}.lead img,.card img{width:100%;height:100%;object-fit:cover}
.lead .txt{padding:18px 18px 18px 0}
.chip{display:inline-block;padding:6px 11px;border-radius:999px;background:#ffe3d6;color:#b4461d}
.lead h2{font:400 clamp(2rem,3.6vw,2.9rem)/1.02 var(--head);letter-spacing:-.02em;margin:18px 0 14px;text-wrap:balance}
.lead .mono{color:var(--muted)}
.read{display:inline-flex;align-items:center;gap:8px;margin-top:24px;padding:10px 16px;border-radius:999px;background:var(--ink);color:#fff;font-size:14px}
.head{display:flex;justify-content:space-between;align-items:baseline;margin:72px 0 20px}
.head h2{font:400 2rem var(--head);letter-spacing:-.02em}.head .mono{color:var(--muted)}
.cards{display:flex;flex-direction:column;gap:12px}
.card{display:grid;grid-template-columns:120px 1fr auto;align-items:center;gap:22px;padding:12px 22px 12px 12px;border-radius:22px;background:rgb(255 255 255/.75);border:1px solid var(--line);transition:transform .3s,box-shadow .3s,background .3s}
.card:not(:has(.art)){grid-template-columns:1fr auto;padding-left:24px}
.card:hover{transform:translateY(-2px);background:#fff;box-shadow:0 18px 40px rgb(43 29 20/.08)}
.card .art{aspect-ratio:4/3;border-radius:14px;overflow:hidden}
.card .mono{color:var(--muted)}
.card h3{font:500 1.15rem/1.3 var(--sans);letter-spacing:-.02em;margin-top:4px}
.go{display:grid;place-items:center;width:38px;height:38px;border-radius:50%;border:1px solid var(--line);transition:background .3s,color .3s}
.card:hover .go{background:var(--ink);color:#fff}
@media (max-width:760px){.lead{grid-template-columns:1fr}.lead .txt{padding:6px 12px 14px}.card{grid-template-columns:88px 1fr}.go{display:none}}`,
};
