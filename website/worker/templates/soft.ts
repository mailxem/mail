// Soft: a pastel color mesh with frosted-glass cards.

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
  const cards = h.issues
    .map(
      (i, n) =>
        `<li ${reveal(n)}><a href="${href(h, i)}"><div class="art">${i.cover ? `<img src="${esc(i.cover)}" alt="" loading="lazy">` : `<span class="orb"></span>`}</div><div class="cap"><p class="mono">No. ${no(h, n)} · ${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></div><span class="go">${arrow}</span></a></li>`,
    )
    .join("");
  return `<div class="mesh" aria-hidden="true"><i></i><i></i><i></i></div>${bar(h.base, h.name, h.logo)}<main><section class="hero wrap" ${reveal(0)}><p class="pill mono"><span class="dot"></span>${esc(h.cadence || "Newsletter")}${h.readers >= 10 ? ` · ${h.readers.toLocaleString("en")} readers` : ""}</p><h1>${esc(h.description || h.name)}</h1>${h.description ? `<p class="lede">${esc(h.name)}, in your inbox.</p>` : ""}<div class="glass">${note(h)}${subscribe(h.base, h)}</div></section>
<section class="wrap list" id="issues"><header class="head"><h2>Latest issues</h2><span class="mono">${count(h)}</span></header>${cards ? `<ol class="cards">${cards}</ol>` : `<p class="empty">The first issue is on its way.</p>`}</section></main>${foot(h.base, h.name, h.description)}`;
}

export const soft: Template = {
  label: "Soft",
  home,
  css: `:root{--bg:#f7f5fc;--surface:rgb(255 255 255/.6);--field:rgb(255 255 255/.85);--ink:#1d1a2b;--muted:#6d6883;--line:rgb(29 26 43/.08);--accent:#1d1a2b;--on-accent:#fff;--live:#8b5cf6;--ring:rgb(139 92 246/.15);--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:Geist,sans-serif;--head-weight:600;--head-track:-.04em}
.mesh{position:fixed;inset:0;z-index:-1;overflow:hidden;filter:blur(70px) saturate(1.15)}
.mesh i{position:absolute;border-radius:50%;opacity:.75;animation:float 18s ease-in-out infinite alternate}
.mesh i:nth-child(1){width:46vw;height:46vw;left:-8vw;top:-12vw;background:#c4b5fd}
.mesh i:nth-child(2){width:40vw;height:40vw;right:-10vw;top:-6vw;background:#a5f3fc;animation-delay:-6s}
.mesh i:nth-child(3){width:38vw;height:38vw;left:30vw;top:22vw;background:#fbcfe8;animation-delay:-12s}
@keyframes float{to{transform:translate(4vw,3vw) scale(1.08)}}
@media (prefers-reduced-motion:reduce){.mesh i{animation:none}}
.nav nav .btn{box-shadow:0 8px 20px rgb(29 26 43/.18)}
.hero{display:flex;flex-direction:column;align-items:center;text-align:center;padding-top:96px;padding-bottom:104px}
.pill{display:inline-flex;align-items:center;gap:10px;padding:8px 14px;border-radius:999px;background:var(--surface);border:1px solid rgb(255 255 255/.8);backdrop-filter:blur(12px);color:var(--muted);box-shadow:0 4px 18px rgb(29 26 43/.06)}
.hero h1{font:600 clamp(2.8rem,6.6vw,5.4rem)/.98 var(--sans);letter-spacing:-.05em;margin:28px 0 18px;max-width:15ch;text-wrap:balance}
.lede{font-size:1.2rem;color:var(--muted);margin-bottom:38px}
.glass{display:flex;flex-direction:column;align-items:center;width:100%;max-width:520px;padding:22px;border-radius:28px;background:var(--surface);border:1px solid rgb(255 255 255/.85);backdrop-filter:blur(18px);box-shadow:0 24px 60px rgb(76 56 140/.14)}
.glass .fine{display:none}
.glass form.sub{max-width:100%;border-color:transparent;box-shadow:0 1px 2px rgb(0 0 0/.05)}
.list{padding-bottom:120px}
.head{display:flex;justify-content:space-between;align-items:baseline;margin-bottom:22px}
.head h2{font:600 1.6rem var(--sans);letter-spacing:-.035em}.head .mono{color:var(--muted)}
.cards{list-style:none;margin:0;padding:0;display:grid;grid-template-columns:repeat(3,1fr);gap:18px}
.cards a{position:relative;display:block;height:100%;padding:10px 10px 22px;border-radius:26px;background:var(--surface);border:1px solid rgb(255 255 255/.85);backdrop-filter:blur(16px);box-shadow:0 10px 30px rgb(76 56 140/.08);transition:transform .35s cubic-bezier(.2,.7,.2,1),box-shadow .35s}
.cards a:hover{transform:translateY(-5px);box-shadow:0 24px 50px rgb(76 56 140/.16)}
.art{aspect-ratio:16/10;border-radius:18px;overflow:hidden;display:grid;place-items:center;background:linear-gradient(135deg,#ede9fe,#e0f2fe)}
.art img{width:100%;height:100%;object-fit:cover}
.orb{width:38%;aspect-ratio:1;border-radius:50%;background:radial-gradient(circle at 30% 30%,#fff,#c4b5fd 60%,#8b5cf6);box-shadow:0 20px 40px rgb(139 92 246/.35)}
.cap{padding:16px 12px 0}
.cap .mono{color:var(--muted)}
.cap h3{font:600 1.15rem/1.3 var(--sans);letter-spacing:-.03em;margin-top:6px;padding-right:40px;text-wrap:pretty}
.go{position:absolute;right:20px;bottom:22px;display:grid;place-items:center;width:34px;height:34px;border-radius:50%;background:#fff;box-shadow:0 4px 12px rgb(29 26 43/.1);transition:transform .3s}
.cards a:hover .go{transform:rotate(-45deg)}
@media (max-width:900px){.cards{grid-template-columns:1fr 1fr}}@media (max-width:560px){.cards{grid-template-columns:1fr}.hero{padding-top:64px}}`,
};
