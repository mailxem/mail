// The editorial templates: Paper, Broadsheet, Midnight, Atelier, Letter, Index.

import {
  arrow,
  back,
  count,
  day,
  esc,
  foot,
  bar,
  headline,
  href,
  img,
  no,
  note,
  part,
  readers,
  subscribe,
  fit,
  type Home,
  type Issue,
  type Template,
} from "./base";

/** Paper: an editorial journal. Statement headline, featured issue, grid. */
function paper(h: Home) {
  const [first, ...rest] = h.issues;
  const feature = first
    ? `<a class="feature" href="${href(h, first)}"><div class="art">${img(first) || `<span>${esc(h.name.charAt(0))}</span>`}</div><div class="copy"><p class="mono">No. ${no(h, 0)} · Latest</p><h3>${esc(first.subject)}</h3><p class="mono date">${day(first.sentAt)}</p><span class="read">Read the issue ${arrow}</span></div></a>`
    : `<p class="empty">The first issue is on its way.</p>`;
  const grid = rest
    .map(
      (i, n) =>
        `<li><a href="${href(h, i)}"><div class="art">${img(i) || `<span>${esc(i.subject.charAt(0))}</span>`}</div><p class="mono">No. ${no(h, n + 1)} · ${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></a></li>`,
    )
    .join("");
  return `${bar(h.base, h.name, h.logo)}<main><section class="hero wrap"><div><p class="mono kicker">${esc(h.name)}${h.cadence ? ` — ${esc(h.cadence)}` : ""}</p><h1>${headline(h)}</h1></div><div class="join">${note(h)}<p class="ask">Get each new issue in your inbox. No spam, unsubscribe anytime.</p>${subscribe(h.base, h)}</div></section>
<section class="archive wrap" id="issues"><header class="head"><h2 class="mono">Archive</h2><span class="mono">${count(h)}</span></header>${feature}${grid ? `<ol class="grid">${grid}</ol>` : ""}</section></main>${foot(h.base, h.name, h.description)}`;
}

/** Broadsheet: a newspaper front page. */
function broadsheet(h: Home) {
  const [lead, ...rest] = h.issues;
  const today = new Date().toLocaleDateString("en", {
    weekday: "long",
    month: "long",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });
  const leadHtml = lead
    ? `<a class="lead" href="${href(h, lead)}">${lead.cover ? `<div class="art"><img src="${esc(lead.cover)}" alt=""></div>` : ""}<div class="copy"><p class="mono kick">Latest edition · No. ${no(h, 0)}</p><h2>${esc(lead.subject)}</h2><p class="mono">${day(lead.sentAt)}</p><span class="read">Continue reading ${arrow}</span></div></a>`
    : `<p class="empty">The first edition is on its way.</p>`;
  const cols = rest
    .map(
      (i, n) =>
        `<a class="story" href="${href(h, i)}">${i.cover ? `<div class="art"><img src="${esc(i.cover)}" alt="" loading="lazy"></div>` : ""}<p class="mono">No. ${no(h, n + 1)} · ${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></a>`,
    )
    .join("");
  return `<header class="mast wrap"><div class="top mono"><span>${today}</span><span>${h.issues.length ? `No. ${no(h, 0)}` : "First edition soon"}${readers(h.readers) ? ` · ${readers(h.readers)}` : ""}</span><a href="#subscribe">Subscribe ${arrow}</a></div>
<h1 class="plate">${esc(h.name)}</h1><p class="line">${esc(h.description || `${h.cadence || "Every edition"}, delivered free.`)}</p></header>
<main class="wrap" id="issues">${leadHtml}${cols ? `<div class="cols">${cols}</div>` : ""}
<section class="desk"><div><p class="mono kick">Delivered free</p><h2>Read the next edition first.</h2></div><div>${note(h)}${subscribe(h.base, h)}</div></section></main>${foot(h.base, h.name, h.description)}`;
}

/** Midnight: a dark product-journal on a dotted grid. */
function midnight(h: Home) {
  const cards = h.issues
    .map(
      (i, n) =>
        `<li><a href="${href(h, i)}"><div class="art">${img(i) || `<span class="mono">No. ${no(h, n)}</span>`}</div><div class="row mono"><span>No. ${no(h, n)}</span><span>${day(i.sentAt)}</span></div><h3>${esc(i.subject)}</h3></a></li>`,
    )
    .join("");
  return `${bar(h.base, h.name, h.logo)}<main><section class="hero wrap">${h.issues[0] ? `<a class="pill mono" href="${href(h, h.issues[0])}"><span class="dot"></span>New · No. ${no(h, 0)}<span class="sep"></span><span class="t">${esc(h.issues[0].subject)}</span>${arrow}</a>` : `<p class="pill mono"><span class="dot"></span>${esc(h.cadence || "Newsletter")}</p>`}<h1>${headline(h)}</h1>${h.description ? `<p class="lede">${esc(h.name)} lands in your inbox ${h.cadence ? esc(h.cadence.toLowerCase()) : "with every issue"}. Read the archive below.</p>` : ""}${note(h)}${subscribe(h.base, h, "Join")}</section>
<section class="archive wrap" id="issues"><header class="head mono"><span>Archive</span><span>${count(h)}</span></header>${cards ? `<ol class="cards">${cards}</ol>` : `<p class="empty">The first issue is on its way.</p>`}</section></main>${foot(h.base, h.name, h.description)}`;
}

/** Atelier: a magazine spread. Sticky introduction beside a numbered index. */
function atelier(h: Home) {
  const items = h.issues
    .map(
      (i, n) =>
        `<li><a href="${href(h, i)}"><span class="num">${no(h, n)}</span><div class="body">${i.cover ? `<div class="art"><img src="${esc(i.cover)}" alt="" loading="lazy"></div>` : ""}<p class="mono">${day(i.sentAt)}</p><h3>${esc(i.subject)}</h3></div></a></li>`,
    )
    .join("");
  return `<div class="spread wrap"><aside class="intro"><a class="brand" href="${esc(h.base || "/")}">${h.logo ? `<img src="${esc(h.logo)}" alt="">` : ""}<span class="mono">${esc(h.cadence || "A newsletter")}</span></a><h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}${note(h)}${subscribe(h.base, h)}<p class="mono rss"><a href="${esc(h.base)}/feed.xml">RSS</a> · ${count(h)}</p></aside>
<main class="index" id="issues">${items ? `<ol>${items}</ol>` : `<p class="empty">The first issue is on its way.</p>`}</main></div>${foot(h.base, h.name, h.description)}`;
}

/** Letter: one column of type, issues grouped by year. */
function letter(h: Home) {
  const years = new Map<string, Issue[]>();
  for (const i of h.issues) {
    const y = part(i.sentAt, { year: "numeric" });
    years.set(y, [...(years.get(y) || []), i]);
  }
  const archive = [...years]
    .map(
      ([y, list]) =>
        `<section><h2 class="mono">${y}</h2><ol>${list.map((i) => `<li><a href="${href(h, i)}"><span>${esc(i.subject)}</span><time class="mono" datetime="${esc(i.sentAt)}">${part(i.sentAt, { month: "short", day: "numeric" })}</time></a></li>`).join("")}</ol></section>`,
    )
    .join("");
  return `<main class="sheet"><header>${h.logo ? `<img class="logo" src="${esc(h.logo)}" alt="">` : ""}<h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}${note(h)}${subscribe(h.base, h)}</header>
<div class="archive" id="issues">${archive || `<p class="empty">The first letter is on its way.</p>`}</div></main>${foot(h.base, h.name, h.description)}`;
}

/** Index: Swiss grotesk. A huge condensed nameplate and dated rows. */
function index(h: Home) {
  const rows = h.issues
    .map(
      (i, n) =>
        `<li><a href="${href(h, i)}"><div class="when"><span class="d">${part(i.sentAt, { day: "2-digit" })}</span><span class="mono">${part(i.sentAt, { month: "short", year: "numeric" })}</span></div><h3>${esc(i.subject)}</h3><span class="mono no">No. ${no(h, n)}</span>${i.cover ? `<div class="thumb"><img src="${esc(i.cover)}" alt="" loading="lazy"></div>` : `<div class="thumb"></div>`}</a></li>`,
    )
    .join("");
  return `${bar(h.base, h.name, h.logo)}<main><section class="hero wrap"><h1 style="${fit(h.name, 150, 4)}">${esc(h.name)}</h1><div class="under"><p class="lede">${esc(h.description || `${h.cadence || "Every issue"}, in your inbox.`)}</p><div>${note(h)}${subscribe(h.base, h)}</div></div></section>
<section class="list wrap" id="issues"><header class="head mono"><span>Index</span><span>${count(h)}</span></header>${rows ? `<ol>${rows}</ol>` : `<p class="empty">The first issue is on its way.</p>`}</section></main>${foot(h.base, h.name, h.description)}`;
}

export const editorial: Record<string, Template> = {
  paper: {
    label: "Paper",
    grain: true,
    home: paper,
    css: `:root{--bg:#f2efe8;--surface:#fbfaf6;--field:#fbfaf6;--ink:#1c1b18;--muted:#74706a;--line:rgb(28 27 24/.12);--accent:#1c1b18;--on-accent:#f7f5ef;--grain:.07;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:"Instrument Serif",Georgia,serif}
.hero{display:grid;grid-template-columns:1.5fr 1fr;gap:64px;align-items:end;padding-top:96px;padding-bottom:104px}
.kicker{color:var(--muted);margin-bottom:28px}
.hero h1{font:400 clamp(3rem,6.4vw,5.6rem)/.96 var(--head);letter-spacing:-.025em;text-wrap:balance}
.join{padding-bottom:10px}.ask{color:var(--muted);margin-bottom:18px;max-width:34ch}
.archive{padding-bottom:112px}
.head{display:flex;justify-content:space-between;padding:14px 0;border-top:1px solid var(--ink);color:var(--muted)}.head h2{color:var(--ink)}
.feature{display:grid;grid-template-columns:7fr 5fr;gap:48px;align-items:center;padding:32px 0 56px;border-bottom:1px solid var(--line)}
.art{aspect-ratio:3/2;border-radius:6px;overflow:hidden;background:var(--surface);display:grid;place-items:center;font:400 4rem var(--head);color:var(--muted);box-shadow:0 0 0 1px var(--line)}
.art img{width:100%;height:100%;object-fit:cover;transition:transform .8s cubic-bezier(.2,.7,.2,1)}
a:hover .art img{transform:scale(1.025)}
.feature .copy .mono{color:var(--muted)}
.feature h3{font:400 clamp(2rem,3.6vw,3.1rem)/1.02 var(--head);letter-spacing:-.02em;margin:16px 0 18px;text-wrap:balance}
.read{display:inline-flex;align-items:center;gap:8px;margin-top:28px;padding-bottom:3px;border-bottom:1px solid var(--ink);font-size:14px}
.grid{list-style:none;margin:0;padding:48px 0 0;display:grid;grid-template-columns:repeat(3,1fr);gap:56px 32px}
.grid .mono{color:var(--muted);margin:18px 0 8px}
.grid h3{font:450 1.15rem/1.35 var(--sans);letter-spacing:-.012em;text-wrap:pretty}
.grid a:hover h3,.feature:hover h3{text-decoration:underline;text-decoration-thickness:1px;text-underline-offset:4px}
@media (max-width:900px){.hero,.feature{grid-template-columns:1fr;gap:32px}.hero{padding-top:56px;padding-bottom:64px}.grid{grid-template-columns:1fr 1fr}}
@media (max-width:560px){.grid{grid-template-columns:1fr}}`,
  },
  broadsheet: {
    label: "Broadsheet",
    grain: true,
    home: broadsheet,
    css: `:root{--bg:#ede6d8;--surface:#f6f1e6;--field:#f8f4ea;--ink:#171411;--muted:#6a6155;--line:rgb(23 20 17/.18);--accent:#7d1a1a;--on-accent:#f8f4ea;--radius:0;--grain:.09;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:"Instrument Serif",Georgia,serif}
body{font-family:Newsreader,Georgia,serif;font-size:18px}
.mast{text-align:center;padding-top:22px}
.top{display:flex;justify-content:space-between;gap:12px;padding-bottom:16px;border-bottom:1px solid var(--ink);color:var(--muted)}
.top a{display:inline-flex;align-items:center;gap:6px;color:var(--accent)}
.plate{font:400 clamp(3.6rem,13vw,10.5rem)/.82 var(--head);letter-spacing:-.035em;margin:44px 0 30px}
.line{font:italic 400 1.2rem Newsreader,Georgia,serif;color:var(--muted);padding:13px 0;border-top:1px solid var(--ink);border-bottom:1px solid var(--ink);box-shadow:0 -4px 0 -3px var(--ink),0 4px 0 -3px var(--ink)}
.lead{display:grid;grid-template-columns:1.25fr 1fr;gap:48px;align-items:center;padding:52px 0;border-bottom:1px solid var(--ink)}
.art{aspect-ratio:4/3;overflow:hidden;background:var(--surface)}.art img{width:100%;height:100%;object-fit:cover;filter:saturate(.82) contrast(1.04)}
.kick{color:var(--accent);margin-bottom:18px}
.lead h2{font:400 clamp(2.4rem,4.6vw,4rem)/.98 var(--head);letter-spacing:-.025em;margin:0 0 20px;text-wrap:balance}
.lead .mono{color:var(--muted)}
.read{display:inline-flex;align-items:center;gap:8px;margin-top:26px;font:500 14px var(--sans);color:var(--accent)}
.cols{display:grid;grid-template-columns:repeat(3,1fr);border-bottom:1px solid var(--ink)}
.story{padding:32px 28px 40px;border-left:1px solid var(--line)}
.story:nth-child(3n+1){border-left:0;padding-left:0}.story:nth-child(3n){padding-right:0}
.story .art{aspect-ratio:3/2;margin-bottom:18px}
.story .mono{color:var(--muted)}
.story h3{font:400 1.75rem/1.06 var(--head);letter-spacing:-.015em;margin-top:10px;text-wrap:balance}
.lead:hover h2,.story:hover h3{text-decoration:underline;text-decoration-thickness:1px;text-underline-offset:5px}
.desk{display:grid;grid-template-columns:1fr 1fr;gap:40px;align-items:center;margin:72px 0 96px;padding:44px 48px;border:1px solid var(--ink);background:var(--surface)}
.desk h2{font:400 clamp(2rem,3.4vw,2.8rem)/1 var(--head);letter-spacing:-.02em}
.foot{border-top:1px solid var(--ink);font-family:var(--sans)}
.reader{font-family:var(--sans)}
@media (max-width:860px){.lead,.desk{grid-template-columns:1fr}.desk{padding:32px 24px}.cols{grid-template-columns:1fr}.story{border-left:0;padding:28px 0;border-top:1px solid var(--line)}.top span:nth-child(2){display:none}}`,
  },
  midnight: {
    label: "Midnight",
    home: midnight,
    css: `:root{--bg:#09090b;--surface:#111113;--field:rgb(255 255 255/.03);--ink:#f4f4f5;--muted:#8f8f99;--line:rgb(255 255 255/.09);--accent:#e8ff6b;--on-accent:#0b0d02;--live:#e8ff6b;--ring:rgb(232 255 107/.12);--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:Geist,ui-sans-serif,sans-serif;--head-weight:500;--head-track:-.04em}
body{background-image:radial-gradient(circle at 1px 1px,rgb(255 255 255/.07) 1px,transparent 0);background-size:22px 22px;background-attachment:fixed}
main{-webkit-mask-image:none}
.hero{display:flex;flex-direction:column;align-items:center;text-align:center;padding-top:112px;padding-bottom:112px;position:relative}
.hero::before{content:"";position:absolute;inset:-80px 10% auto;height:420px;background:radial-gradient(closest-side,rgb(232 255 107/.10),transparent);pointer-events:none;z-index:-1}
.pill{display:inline-flex;align-items:center;gap:10px;max-width:100%;padding:7px 14px;border:1px solid var(--line);border-radius:999px;color:var(--muted);background:rgb(255 255 255/.02);transition:border-color .2s}
.pill:hover{border-color:rgb(255 255 255/.22);color:var(--ink)}.pill .sep{width:1px;height:12px;background:var(--line)}.pill .t{text-transform:none;letter-spacing:0;font-family:var(--sans);font-size:13px;color:var(--ink);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;max-width:34ch}
.hero h1{font:500 clamp(2.6rem,6vw,4.75rem)/1.02 var(--sans);letter-spacing:-.045em;margin:30px 0 22px;max-width:16ch;text-wrap:balance}
.lede{color:var(--muted);font-size:1.15rem;max-width:46ch;margin-bottom:36px}
.hero form.sub{background:rgb(255 255 255/.03);backdrop-filter:blur(8px)}
.archive{padding-bottom:120px}
.head{display:flex;justify-content:space-between;padding:0 0 18px;color:var(--muted)}
.cards{list-style:none;margin:0;padding:0;display:grid;grid-template-columns:repeat(3,1fr);gap:16px}
.cards a{display:flex;flex-direction:column;height:100%;padding:10px 10px 22px;border:1px solid var(--line);border-radius:18px;background:linear-gradient(180deg,rgb(255 255 255/.03),rgb(255 255 255/.01));transition:border-color .25s,background .25s}
.cards a:hover{border-color:rgb(255 255 255/.2);background:rgb(255 255 255/.04)}
.art{aspect-ratio:16/10;border-radius:11px;overflow:hidden;display:grid;place-items:center;background:radial-gradient(circle at 30% 20%,rgb(232 255 107/.14),transparent 60%),#141417;color:var(--muted)}
.art img{width:100%;height:100%;object-fit:cover;opacity:.92;transition:opacity .3s,transform .8s cubic-bezier(.2,.7,.2,1)}
.cards a:hover .art img{opacity:1;transform:scale(1.03)}
.row{display:flex;justify-content:space-between;color:var(--muted);margin:18px 8px 8px}
.cards h3{font:500 1.08rem/1.38 var(--sans);letter-spacing:-.015em;margin:0 8px;text-wrap:pretty}
.reader iframe{border-color:transparent}
@media (max-width:900px){.cards{grid-template-columns:1fr 1fr}}@media (max-width:560px){.cards{grid-template-columns:1fr}.hero{padding-top:72px;padding-bottom:80px}}`,
  },
  atelier: {
    label: "Atelier",
    grain: true,
    home: atelier,
    css: `:root{--bg:#e8e1d5;--surface:#f2ece2;--field:transparent;--ink:#221b16;--muted:#7a6d61;--line:rgb(34 27 22/.18);--accent:#221b16;--on-accent:#f2ece2;--radius:0;--grain:.08;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:"Instrument Serif",Georgia,serif}
.spread{display:grid;grid-template-columns:5fr 7fr;gap:88px;padding-top:56px;padding-bottom:96px}
.intro{position:sticky;top:56px;align-self:start;display:flex;flex-direction:column}
.intro .brand{gap:12px;color:var(--muted)}.intro .brand img{width:34px;height:34px;border-radius:50%}
.intro h1{font:italic 400 clamp(3.6rem,7vw,6.4rem)/.9 var(--head);letter-spacing:-.03em;margin:88px 0 26px;text-wrap:balance}
.lede{font:400 1.28rem/1.5 Newsreader,Georgia,serif;color:var(--muted);margin-bottom:40px;max-width:30ch}
.intro form.sub{padding:0;border:0;border-bottom:1px solid var(--ink);border-radius:0;background:transparent;box-shadow:none}
.intro form.sub input{padding:0}
.intro form.sub .btn{background:transparent;color:var(--ink);padding:0 0 0 12px;height:44px}
.rss{color:var(--muted);margin-top:36px}.rss a{text-decoration:underline;text-underline-offset:3px}
.index ol{list-style:none;margin:0;padding:0}
.index li{border-top:1px solid var(--line)}.index li:first-child{border-top:1px solid var(--ink)}
.index a{display:grid;grid-template-columns:96px 1fr;gap:24px;padding:36px 0 44px}
.num{font:400 3.6rem/1 var(--head);color:var(--muted);letter-spacing:-.02em;transition:color .3s}
.index a:hover .num{color:var(--ink)}
.art{aspect-ratio:3/2;overflow:hidden;margin-bottom:22px}.art img{width:100%;height:100%;object-fit:cover;filter:saturate(.88);transition:transform 1s cubic-bezier(.2,.7,.2,1)}
.index a:hover .art img{transform:scale(1.03)}
.body .mono{color:var(--muted)}
.index h3{font:400 clamp(1.8rem,3vw,2.5rem)/1.04 var(--head);letter-spacing:-.02em;margin-top:10px;text-wrap:balance}
@media (max-width:960px){.spread{grid-template-columns:1fr;gap:40px}.intro{position:static}.intro h1{margin-top:48px}.index a{grid-template-columns:56px 1fr;gap:16px}.num{font-size:2.4rem}}`,
  },
  letter: {
    label: "Letter",
    home: letter,
    css: `:root{--bg:#fbfaf6;--surface:#fff;--field:#fff;--ink:#26231f;--muted:#8a8279;--line:rgb(38 35 31/.12);--accent:#26231f;--on-accent:#fff;--radius:12px;--live:#3f7d63;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:"Instrument Serif",Georgia,serif}
body{font:400 19px/1.7 Newsreader,Georgia,serif}
.sheet{max-width:640px;margin:0 auto;padding:112px 28px 64px}
.sheet .logo{width:48px;height:48px;border-radius:50%;object-fit:cover;margin-bottom:36px}
.sheet h1{font:400 clamp(2.8rem,7vw,3.8rem)/1 var(--head);letter-spacing:-.02em;margin-bottom:20px}
.sheet .lede{font-size:1.3rem;line-height:1.55;color:var(--muted);margin-bottom:36px}
.sheet form.sub{max-width:100%}
.archive{margin-top:96px}
.archive section+section{margin-top:48px}
.archive h2{color:var(--muted);padding-bottom:12px;border-bottom:1px solid var(--line)}
.archive ol{list-style:none;margin:0;padding:0}
.archive a{display:flex;align-items:baseline;justify-content:space-between;gap:24px;padding:15px 0;border-bottom:1px solid var(--line)}
.archive span{font-size:1.15rem;line-height:1.4;text-wrap:pretty}
.archive a:hover span{text-decoration:underline;text-decoration-thickness:1px;text-underline-offset:4px}
.archive time{color:var(--muted);white-space:nowrap}
.foot,.nav{max-width:696px}
.reader h1{font-size:clamp(2.2rem,5vw,3rem)}`,
  },
  index: {
    label: "Index",
    grain: true,
    home: index,
    css: `:root{--bg:#e9e7e2;--surface:#f3f2ee;--field:#f6f5f2;--ink:#111110;--muted:#6d6a64;--line:rgb(17 17 16/.16);--accent:#ff4a1c;--on-accent:#fff;--radius:4px;--live:#ff4a1c;--grain:.06;--sans:Geist,ui-sans-serif,system-ui,sans-serif;--mono:"Geist Mono",ui-monospace,monospace;--head:Bricolage,Geist,sans-serif;--head-weight:600;--head-track:-.03em}
.hero{padding-top:40px;padding-bottom:72px}
.hero h1{font-weight:700;line-height:.8;font-family:var(--head);font-stretch:75%;letter-spacing:-.035em;text-transform:uppercase;margin-left:-.04em;overflow-wrap:anywhere}
.under{display:grid;grid-template-columns:1fr 1fr;gap:48px;align-items:end;margin-top:44px;padding-top:24px;border-top:1px solid var(--ink)}
.lede{font:500 clamp(1.25rem,2vw,1.6rem)/1.3 var(--sans);letter-spacing:-.02em;max-width:24ch;text-wrap:balance}
.list{padding-bottom:120px}
.head{display:flex;justify-content:space-between;padding:12px 0;border-top:1px solid var(--ink);border-bottom:1px solid var(--ink)}
.list ol{list-style:none;margin:0;padding:0}
.list a{position:relative;display:grid;grid-template-columns:150px 1fr auto 168px;gap:32px;align-items:center;padding:26px 0;border-bottom:1px solid var(--line)}
.when{display:flex;flex-direction:column}
.d{font:700 3.8rem/.85 var(--head);font-stretch:75%;letter-spacing:-.03em}
.when .mono{color:var(--muted);margin-top:8px}
.list h3{font:600 clamp(1.4rem,2.4vw,2.1rem)/1.05 var(--head);font-stretch:85%;letter-spacing:-.025em;text-wrap:balance;transition:color .2s}
.list a:hover h3{color:var(--accent)}
.no{color:var(--muted)}
.thumb{aspect-ratio:3/2;border-radius:4px;overflow:hidden;background:var(--surface)}
.thumb img{width:100%;height:100%;object-fit:cover;filter:grayscale(1) contrast(1.05);transition:filter .3s}
.list a:hover .thumb img{filter:none}
@media (max-width:860px){.under{grid-template-columns:1fr}.list a{grid-template-columns:84px 1fr;gap:18px}.d{font-size:2.6rem}.no,.thumb{display:none}}`,
  },
};
