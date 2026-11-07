// Cinema: the latest issue's cover fills the screen, like a film poster.

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
  const reel = rest
    .map(
      (i, n) =>
        `<a class="frame" href="${href(h, i)}" ${reveal(n)}><div class="still">${i.cover ? `<img src="${esc(i.cover)}" alt="" loading="lazy">` : ""}<span class="no mono">${no(h, n + 1)}</span></div><p class="mono">${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></a>`,
    )
    .join("");
  const poster = first
    ? `<section class="poster">${first.cover ? `<img class="bg" src="${esc(first.cover)}" alt="">` : ""}<div class="tint"></div>${bar(h.base, h.name, h.logo)}<a class="now wrap" href="${href(h, first)}"><p class="mono">Now showing · No. ${no(h, 0)} · ${day(first.sentAt)}</p><h1>${esc(first.subject)}</h1><span class="play"><span>${arrow}</span>Read the latest issue</span></a></section>`
    : `<section class="poster bare">${bar(h.base, h.name, h.logo)}<div class="now wrap"><h1>${esc(h.name)}</h1></div></section>`;
  return `${poster}<main><section class="about wrap" ${reveal(0)}><div><p class="mono">${esc(h.cadence || "The newsletter")}</p><h2>${esc(h.name)}</h2>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}</div><div>${note(h)}${subscribe(h.base, h)}</div></section>
${reel ? `<section class="reel wrap" id="issues"><header class="head"><h2>Previously</h2><span class="mono">${count(h)}</span></header><div class="strip">${reel}</div></section>` : ""}</main>${foot(h.base, h.name, h.description)}`;
}

export const cinema: Template = {
  label: "Cinema",
  home,
  css: `:root{--bg:#0a0a0b;--surface:#151517;--field:rgb(255 255 255/.06);--ink:#f5f3ee;--muted:#9a978f;--line:rgb(255 255 255/.12);--accent:#f5f3ee;--on-accent:#0a0a0b;--live:#ff5a3c;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:"Instrument Serif",Georgia,serif}
.poster{position:relative;min-height:min(100svh,920px);display:flex;flex-direction:column;overflow:hidden;color:#fff}
.poster .bg{position:absolute;inset:0;width:100%;height:100%;object-fit:cover;transform:scale(1.04);animation:drift 22s ease-in-out infinite alternate}
@keyframes drift{to{transform:scale(1.12) translate(-1.5%,-1%)}}
@media (prefers-reduced-motion:reduce){.poster .bg{animation:none}}
.tint{position:absolute;inset:0;background:linear-gradient(180deg,rgb(0 0 0/.55) 0,rgb(0 0 0/.1) 30%,rgb(0 0 0/.25) 55%,#0a0a0b 100%)}
.poster .nav{position:relative;z-index:1;--muted:rgb(255 255 255/.75)}
.now{position:relative;z-index:1;margin-top:auto;padding-bottom:72px}
.now .mono{color:rgb(255 255 255/.8)}
.now h1{font:400 clamp(3rem,8vw,7.4rem)/.92 var(--head);letter-spacing:-.03em;max-width:14ch;margin:18px 0 32px;text-wrap:balance;text-shadow:0 2px 30px rgb(0 0 0/.35)}
.play{display:inline-flex;align-items:center;gap:14px;font-size:15px}
.play span{display:grid;place-items:center;width:54px;height:54px;border-radius:50%;background:#fff;color:#000;transition:transform .3s}
.now:hover .play span{transform:scale(1.08)}
.bare{min-height:60vh;background:radial-gradient(circle at 30% 20%,#2a2620,#0a0a0b 70%)}
.about{display:grid;grid-template-columns:1.2fr 1fr;gap:56px;align-items:end;padding-top:64px;padding-bottom:96px;border-bottom:1px solid var(--line)}
.about .mono{color:var(--muted)}
.about h2{font:400 clamp(2.4rem,4.4vw,3.6rem)/1 var(--head);letter-spacing:-.025em;margin:14px 0 12px}
.lede{color:var(--muted);font-size:1.15rem;max-width:42ch}
.reel{padding-top:56px;padding-bottom:112px}
.head{display:flex;justify-content:space-between;align-items:baseline;margin-bottom:24px}
.head h2{font:400 2.2rem var(--head)}.head .mono{color:var(--muted)}
.strip{-webkit-mask-image:linear-gradient(90deg,#000 85%,transparent);mask-image:linear-gradient(90deg,#000 85%,transparent);display:grid;grid-auto-flow:column;grid-auto-columns:minmax(280px,1fr);gap:18px;overflow-x:auto;scroll-snap-type:x mandatory;padding-bottom:12px;scrollbar-width:thin}
.frame{scroll-snap-align:start}
.still{position:relative;aspect-ratio:2/3;border-radius:10px;overflow:hidden;background:var(--surface)}
.still img{width:100%;height:100%;object-fit:cover;transition:transform 1s cubic-bezier(.2,.7,.2,1),filter .4s;filter:saturate(.9)}
.frame:hover img{transform:scale(1.05);filter:none}
.no{position:absolute;top:12px;left:12px;padding:4px 8px;border-radius:6px;background:rgb(0 0 0/.55);backdrop-filter:blur(6px)}
.frame .mono{color:var(--muted);margin-top:14px}
.frame h3{font:400 1.6rem/1.05 var(--head);letter-spacing:-.015em;margin-top:6px;text-wrap:balance}
.reader iframe{border-color:transparent}
@media (max-width:820px){.about{grid-template-columns:1fr;gap:28px}.now{padding-bottom:48px}}`,
};
