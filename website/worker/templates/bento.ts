// Bento: the whole newsletter as one tiled grid of rounded panels.

import {
  arrow,
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

const tints = ["#e8e2ff", "#dcf3e6", "#ffe7d6", "#dcecff", "#fff1b8"];

function home(h: Home) {
  const [first, ...rest] = h.issues;
  const issues = rest
    .map(
      (i, n) =>
        `<a class="tile issue ${n % 5 === 0 ? "wide" : ""}" href="${href(h, i)}" ${reveal(n, `--tint:${tints[n % tints.length]}`)}><div class="art">${i.cover ? `<img src="${esc(i.cover)}" alt="" loading="lazy">` : `<span>${esc(i.subject.charAt(0))}</span>`}</div><div class="cap"><p class="mono">No. ${no(h, n + 1)} · ${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></div></a>`,
    )
    .join("");
  return `${bar(h.base, h.name, h.logo)}<main class="wrap"><section class="board">
<div class="tile hero" ${reveal(0)}>${h.logo ? `<img class="logo" src="${esc(h.logo)}" alt="">` : ""}<h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}</div>
<div class="tile join" ${reveal(1)}><p class="mono">Subscribe</p><h2>Get the next issue in your inbox.</h2>${note(h)}${subscribe(h.base, h, "Join")}</div>
<div class="tile stat" ${reveal(2, `--tint:#e8e2ff`)}><span class="big">${h.readers >= 10 ? h.readers.toLocaleString("en") : h.issues.length}</span><p class="mono">${h.readers >= 10 ? "Readers" : "Issues so far"}</p></div>
<div class="tile stat" ${reveal(3, `--tint:#dcf3e6`)}><span class="big">${esc(h.cadence ? h.cadence.replace("Every ", "") : "Fresh")}</span><p class="mono">${h.cadence ? "New issue every" : "Newsletter"}</p></div>
${
  first
    ? `<a class="tile latest" href="${href(h, first)}" ${reveal(4)}>${first.cover ? `<img src="${esc(first.cover)}" alt="">` : ""}<div class="shade"></div><div class="cap"><span class="tag mono">Latest · No. ${no(h, 0)}</span><h3>${esc(first.subject)}</h3><span class="go">${arrow}</span></div></a>`
    : `<div class="tile latest empty">The first issue is on its way.</div>`
}
</section>${issues ? `<section class="board more" id="issues">${issues}</section>` : ""}</main>${foot(h.base, h.name, h.description)}`;
}

export const bento: Template = {
  label: "Bento",
  home,
  css: `:root{--bg:#f3f3ef;--surface:#fff;--field:#f6f6f3;--ink:#141413;--muted:#71716b;--line:rgb(20 20 19/.09);--accent:#141413;--on-accent:#fff;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:Geist,sans-serif;--head-weight:600;--head-track:-.035em}
.board{display:grid;grid-template-columns:repeat(12,1fr);gap:14px;padding-top:12px}
.more{padding-top:14px;padding-bottom:96px}
.tile{position:relative;border-radius:28px;background:var(--tint,var(--surface));padding:30px;overflow:hidden;box-shadow:0 1px 0 rgb(255 255 255/.8) inset,0 1px 2px rgb(0 0 0/.04)}
.hero{grid-column:span 7;grid-row:span 2;display:flex;flex-direction:column;justify-content:flex-end;min-height:420px;background:var(--surface)}
.hero .logo{position:absolute;top:30px;left:30px;width:52px;height:52px;border-radius:15px;object-fit:cover}
.hero h1{font:600 clamp(3rem,6vw,5.4rem)/.92 var(--sans);letter-spacing:-.05em}
.hero .lede{color:var(--muted);font-size:1.15rem;margin-top:16px;max-width:34ch}
.join{grid-column:span 5;grid-row:span 2;background:var(--ink);color:#f5f5f2;display:flex;flex-direction:column;justify-content:flex-end;--line:rgb(255 255 255/.14);--muted:#a3a39c}
.join .mono{color:var(--muted)}
.join h2{font:500 clamp(1.7rem,2.6vw,2.3rem)/1.05 var(--sans);letter-spacing:-.035em;margin:14px 0 28px;max-width:16ch}
.join form.sub{background:rgb(255 255 255/.06);--ink:#f5f5f2}
.join .btn{background:#f5f5f2;color:#141413}
.stat{grid-column:span 3;min-height:190px;display:flex;flex-direction:column;justify-content:space-between}
.big{font:600 clamp(2.4rem,4vw,3.6rem)/1 var(--sans);letter-spacing:-.05em;text-transform:capitalize}
.stat .mono{color:rgb(20 20 19/.6)}
.latest{grid-column:span 6;grid-row:span 1;min-height:190px;padding:0;background:#1d1d1b;color:#fff}
.latest img{position:absolute;inset:0;width:100%;height:100%;object-fit:cover;transition:transform 1s cubic-bezier(.2,.7,.2,1)}
.latest:hover img{transform:scale(1.04)}
.shade{position:absolute;inset:0;background:linear-gradient(180deg,transparent 30%,rgb(0 0 0/.72))}
.latest .cap{position:absolute;inset:auto 0 0 0;padding:24px 26px;display:grid;grid-template-columns:1fr auto;align-items:end;gap:6px 16px}
.tag{grid-column:1/-1;color:rgb(255 255 255/.75)}
.latest h3{font:600 1.45rem/1.15 var(--sans);letter-spacing:-.03em;text-wrap:balance}
.go{display:grid;place-items:center;width:40px;height:40px;border-radius:50%;background:#fff;color:#111;transition:transform .3s}
.latest:hover .go{transform:rotate(-45deg)}
.issue{grid-column:span 4;padding:10px;display:flex;flex-direction:column;transition:transform .35s cubic-bezier(.2,.7,.2,1)}
.issue.wide{grid-column:span 8}
.issue:hover{transform:translateY(-4px)}
.issue .art{aspect-ratio:16/10;border-radius:20px;overflow:hidden;display:grid;place-items:center;background:rgb(255 255 255/.55);font:600 3rem var(--sans);color:rgb(20 20 19/.35)}
.issue.wide .art{aspect-ratio:21/9}
.issue img{width:100%;height:100%;object-fit:cover}
.issue .cap{padding:18px 14px 12px}
.issue .mono{color:rgb(20 20 19/.55)}
.issue h3{font:600 1.15rem/1.3 var(--sans);letter-spacing:-.025em;margin-top:6px;text-wrap:pretty}
.empty{display:grid;place-items:center;color:#aaa}
@media (max-width:900px){.hero,.join,.latest,.issue.wide{grid-column:span 12}.stat{grid-column:span 6}.issue{grid-column:span 6}.hero{min-height:340px}}
@media (max-width:560px){.issue{grid-column:span 12}.tile{border-radius:22px;padding:24px}.issue{padding:8px}}`,
};
