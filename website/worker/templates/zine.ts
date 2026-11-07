// Zine: loud, hand-made print energy. Thick ink, offset shadows, stickers.

import {
  arrow,
  day,
  esc,
  fit,
  foot,
  href,
  no,
  note,
  reveal,
  subscribe,
  type Home,
  type Template,
} from "./base";

const paper = ["#ff8fc7", "#8ed1ff", "#b8f28a", "#fff", "#ffb35c", "#c5a8ff"];
const tilt = [-1.2, 0.8, -0.6, 1.1, -0.9, 0.5];

function home(h: Home) {
  const cards = h.issues
    .map(
      (i, n) =>
        `<li ${reveal(n)}><a class="card" href="${href(h, i)}" style="--bg-card:${paper[n % paper.length]};--tilt:${tilt[n % tilt.length]}deg">${n === 0 ? `<span class="sticker">NEW!</span>` : ""}${i.cover ? `<div class="art"><img src="${esc(i.cover)}" alt="" loading="lazy"></div>` : ""}<p class="meta">#${no(h, n)} — ${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3><span class="go">Read ${arrow}</span></a></li>`,
    )
    .join("");
  return `<header class="nav wrap"><a class="brand" href="${esc(h.base || "/")}">${h.logo ? `<img src="${esc(h.logo)}" alt="">` : ""}<span>${esc(h.name)}</span></a><nav><a href="#issues">Issues</a><a class="btn sm" href="#subscribe">Subscribe</a></nav></header>
<main class="wrap"><section class="hero" ${reveal(0)}><p class="badge">${esc(h.cadence || "Fresh issues")} ✶ ${h.readers >= 10 ? `${h.readers.toLocaleString("en")} readers` : "free"}</p><h1 style="${fit(h.name, 140, 3.4)}">${esc(h.name)}</h1>${h.description ? `<p class="lede"><mark>${esc(h.description)}</mark></p>` : ""}<div class="box">${note(h)}${subscribe(h.base, h, "Sign me up")}</div></section>
<section id="issues"><h2 class="strip"><span>The back issues ✶</span><span>The back issues ✶</span><span>The back issues ✶</span><span>The back issues ✶</span><span>The back issues ✶</span><span>The back issues ✶</span><span>The back issues ✶</span><span>The back issues ✶</span></h2>${cards ? `<ol class="cards">${cards}</ol>` : `<p class="empty">Issue one is at the printer.</p>`}</section></main>${foot(h.base, h.name, h.description)}`;
}

export const zine: Template = {
  label: "Zine",
  home,
  css: `:root{--bg:#ffe24a;--surface:#fff;--field:#fff;--ink:#111;--muted:#3b3a33;--line:#111;--accent:#111;--on-accent:#ffe24a;--radius:14px;--live:#ff3d7f;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:Bricolage,Geist,sans-serif;--head-weight:800;--head-track:-.04em}
.nav .brand{font-weight:700}
.nav nav{color:var(--ink);font-weight:600}
.btn{border:2.5px solid var(--ink);box-shadow:4px 4px 0 var(--ink);font-weight:700;transition:transform .15s,box-shadow .15s}
.btn:hover{opacity:1;transform:translate(-2px,-2px);box-shadow:6px 6px 0 var(--ink)}
.btn:active{transform:translate(2px,2px);box-shadow:1px 1px 0 var(--ink)}
.hero{padding:56px 0 80px}
.badge{display:inline-block;padding:8px 14px;border:2.5px solid var(--ink);border-radius:999px;background:#fff;font:700 13px var(--mono);text-transform:uppercase;transform:rotate(-2deg);box-shadow:3px 3px 0 var(--ink)}
.hero h1{font-family:var(--head);font-weight:800;font-stretch:80%;line-height:.82;letter-spacing:-.04em;text-transform:uppercase;margin:28px 0 22px;-webkit-text-stroke:0}
.lede{font:600 clamp(1.25rem,2.4vw,1.8rem)/1.35 var(--sans);letter-spacing:-.02em;max-width:30ch}
mark{background:#fff;padding:0 .2em;box-decoration-break:clone;-webkit-box-decoration-break:clone;box-shadow:3px 3px 0 var(--ink)}
.box{margin-top:36px;max-width:520px}
form.sub{border:2.5px solid var(--ink);box-shadow:5px 5px 0 var(--ink)}
form.sub:focus-within{box-shadow:7px 7px 0 var(--ink);border-color:var(--ink)}
.fine{color:var(--ink)}
.strip{display:flex;gap:40px;overflow:hidden;white-space:nowrap;margin:0 -28px 48px;padding:14px 0;background:var(--ink);color:var(--bg);font:800 1.6rem var(--head);font-stretch:80%;text-transform:uppercase;letter-spacing:-.02em;transform:rotate(-1.2deg)}
.strip span{animation:slide 18s linear infinite}
@keyframes slide{to{transform:translateX(-100%)}}
@media (prefers-reduced-motion:reduce){.strip span{animation:none}}
.cards{list-style:none;margin:0;padding:0 0 112px;display:grid;align-items:start;grid-template-columns:repeat(auto-fill,minmax(300px,1fr));gap:34px 26px}
.card{position:relative;display:block;padding:16px 16px 22px;background:var(--bg-card);border:2.5px solid var(--ink);border-radius:18px;box-shadow:6px 6px 0 var(--ink);transform:rotate(var(--tilt));transition:transform .25s cubic-bezier(.2,.7,.2,1),box-shadow .25s}
.card:hover{transform:rotate(0) translate(-3px,-3px);box-shadow:10px 10px 0 var(--ink)}
.art{aspect-ratio:16/10;border:2.5px solid var(--ink);border-radius:10px;overflow:hidden;margin-bottom:16px}.art img{width:100%;height:100%;object-fit:cover}
.meta{font:700 13px var(--mono);text-transform:uppercase}
.card h3{font:800 1.7rem/1 var(--head);font-stretch:85%;letter-spacing:-.03em;margin:8px 0 16px;text-wrap:balance}
.go{display:inline-flex;align-items:center;gap:6px;font-weight:700;border-bottom:2.5px solid var(--ink)}
.sticker{position:absolute;top:-16px;right:-12px;z-index:1;display:grid;place-items:center;width:72px;height:72px;border-radius:50%;background:#ff3d7f;color:#fff;border:2.5px solid var(--ink);font:800 15px var(--head);transform:rotate(14deg)}
.foot{border-top:2.5px solid var(--ink);color:var(--ink)}
.reader iframe{border:2.5px solid var(--ink);box-shadow:6px 6px 0 var(--ink)}`,
};
