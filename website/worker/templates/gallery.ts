// Gallery: an image-led archive, covers in a staggered masonry.

import {
  bar,
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

const shapes = ["4/5", "1/1", "3/4", "4/3", "1/1", "4/5"];

function home(h: Home) {
  const tiles = h.issues
    .map(
      (i, n) =>
        `<a class="piece" href="${href(h, i)}" ${reveal(n)}><div class="frame" style="aspect-ratio:${shapes[n % shapes.length]}">${i.cover ? `<img src="${esc(i.cover)}" alt="" loading="lazy">` : `<span>${esc(i.subject)}</span>`}<div class="veil"><p class="mono">No. ${no(h, n)} · ${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></div></div><div class="below"><span class="mono">No. ${no(h, n)}</span><h3>${esc(i.subject)}</h3></div></a>`,
    )
    .join("");
  return `${bar(h.base, h.name, h.logo)}<main><section class="hero wrap" ${reveal(0)}><p class="mono kick">${esc(h.cadence || "A visual newsletter")}</p><h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}<div class="join">${note(h)}${subscribe(h.base, h)}</div></section>
<section class="wall wrap" id="issues">${tiles || `<p class="empty">The first issue is on its way.</p>`}</section></main>${foot(h.base, h.name, h.description)}`;
}

export const gallery: Template = {
  label: "Gallery",
  home,
  css: `:root{--bg:#0d0d0c;--surface:#171715;--field:rgb(255 255 255/.04);--ink:#f2efe9;--muted:#8d8a83;--line:rgb(255 255 255/.12);--accent:#f2efe9;--on-accent:#0d0d0c;--live:#f2efe9;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:"Instrument Serif",Georgia,serif}
.hero{display:flex;flex-direction:column;align-items:center;text-align:center;padding-top:96px;padding-bottom:96px}
.kick{color:var(--muted)}
.hero h1{font:italic 400 clamp(4rem,12vw,10rem)/.86 var(--head);letter-spacing:-.04em;margin:22px 0 22px;text-wrap:balance}
.lede{color:var(--muted);font-size:1.15rem;max-width:44ch;margin-bottom:36px}
.join{display:flex;flex-direction:column;align-items:center;width:100%}
.wall{columns:3 320px;column-gap:18px;padding-bottom:120px}
.piece{display:block;break-inside:avoid;margin-bottom:18px}
.frame{position:relative;border-radius:6px;overflow:hidden;background:var(--surface);display:grid;place-items:center}
.frame>span{padding:28px;font:400 2rem/1.05 var(--head);text-align:center;color:var(--muted)}
.frame img{position:absolute;inset:0;width:100%;height:100%;object-fit:cover;transition:transform 1.1s cubic-bezier(.2,.7,.2,1),filter .5s}
.veil h3{display:none}.veil{position:absolute;inset:0;display:flex;flex-direction:column;justify-content:flex-end;gap:8px;padding:26px;background:linear-gradient(180deg,transparent 35%,rgb(0 0 0/.78));opacity:0;transform:translateY(8px);transition:opacity .45s,transform .45s}
.veil .mono{color:rgb(255 255 255/.7)}
.veil h3{font:400 clamp(1.6rem,2.4vw,2.2rem)/1.02 var(--head);letter-spacing:-.015em;color:#fff;text-wrap:balance}
.piece:hover .veil{opacity:1;transform:none}
.piece:hover img{transform:scale(1.05);filter:brightness(.92)}
.below{padding:14px 2px 8px}
.below .mono{color:var(--muted)}
.below h3{font:400 1.5rem/1.1 var(--head);margin-top:4px}
.reader iframe{border-color:transparent}
@media (hover:none),(max-width:700px){.veil{display:none}}`,
};
