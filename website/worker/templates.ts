// The looks a newsletter website can take. Each template has its own home
// page layout and stylesheet; issue and confirmation pages share markup that
// every stylesheet dresses.

export type Issue = {
  id: string;
  subject: string;
  sentAt: string;
  cover?: string;
};

/** What a home page needs, already escaped where it is markup. */
export type Home = {
  base: string;
  name: string;
  description: string;
  logo: string;
  readers: number;
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

const fonts = ["inter", "fraunces", "newsreader"]
  .flatMap((f) =>
    ["normal", "italic"].map(
      (s) =>
        `@font-face{font-family:${f[0].toUpperCase() + f.slice(1)};src:url(/site-fonts/${f}-wght-${s}.woff2) format("woff2");font-weight:100 900;font-style:${s};font-display:swap}`,
    ),
  )
  .join("");

/** Shared structure; templates restyle it through variables and overrides. */
const base = `${fonts}
*{box-sizing:border-box}html{-webkit-text-size-adjust:100%}
body{margin:0;background:var(--bg);color:var(--ink);font:400 16px/1.6 var(--sans);-webkit-font-smoothing:antialiased;text-rendering:optimizeLegibility}
a{color:inherit;text-decoration:none}img{display:block;max-width:100%}
:focus-visible{outline:2px solid var(--accent);outline-offset:3px}
.wrap{width:100%;max-width:1160px;margin:0 auto;padding:0 24px}
.bar{display:flex;align-items:center;justify-content:space-between;gap:16px;padding-top:20px;padding-bottom:20px}
.brand{display:flex;align-items:center;gap:10px;font-weight:600;letter-spacing:-.01em}
.brand img{width:30px;height:30px;border-radius:8px;object-fit:cover}
.btn{display:inline-flex;align-items:center;justify-content:center;gap:8px;border:0;border-radius:var(--radius,999px);background:var(--accent);color:var(--on-accent);padding:11px 20px;font:600 15px/1 var(--sans);cursor:pointer;white-space:nowrap;transition:transform .15s,filter .15s}
.btn:hover{filter:brightness(1.08)}.btn:active{transform:translateY(1px)}
form.sub{display:flex;gap:6px;width:100%;max-width:460px;padding:6px;border:1px solid var(--line);border-radius:var(--radius,999px);background:var(--surface)}
form.sub input{flex:1;min-width:0;border:0;background:transparent;color:var(--ink);font:400 16px var(--sans);padding:8px 14px}
form.sub input::placeholder{color:var(--muted)}form.sub input:focus{outline:none}
form.sub:focus-within{border-color:var(--accent)}
.fine{color:var(--muted);font-size:14px;margin:12px 0 0}
.note{max-width:460px;margin:0 0 18px;padding:12px 16px;border-radius:12px;background:var(--surface);border:1px solid var(--line);font-size:15px}
.label{font:600 12px/1 var(--sans);letter-spacing:.14em;text-transform:uppercase;color:var(--muted)}
.empty{color:var(--muted)}
.reader{max-width:780px;margin:0 auto;padding:32px 24px 8px}
.reader .back{color:var(--muted);font-size:14px}
.reader time{display:block;color:var(--muted);font-size:14px;margin-top:28px}
.reader h1{font:600 clamp(2rem,5vw,3rem)/1.1 var(--display);letter-spacing:-.02em;margin:8px 0 32px}
.reader iframe{display:block;width:100%;min-height:420px;border:1px solid var(--line);border-radius:16px;background:#fff}
.end,.panel{max-width:780px;margin:0 auto;padding:48px 24px 80px;display:flex;flex-direction:column;align-items:center;text-align:center}
.end h2,.panel h1{font:600 clamp(1.6rem,4vw,2.2rem)/1.15 var(--display);letter-spacing:-.02em;margin:0 0 20px}
.panel{padding-top:96px}.panel p{color:var(--muted);margin:0 0 28px}
footer.foot{display:flex;justify-content:space-between;gap:12px;flex-wrap:wrap;padding-top:28px;padding-bottom:40px;border-top:1px solid var(--line);color:var(--muted);font-size:14px}
footer.foot a:hover{color:var(--ink)}
@media (max-width:560px){form.sub{flex-direction:column;border-radius:18px}form.sub .btn{width:100%}}`;

const count = (readers: number) =>
  readers >= 10
    ? `Join ${readers.toLocaleString("en")} readers. Free, unsubscribe anytime.`
    : "Free, unsubscribe anytime.";

export const subscribe = (
  base: string,
  readers: number,
  button = "Subscribe",
) =>
  `<form class="sub" id="subscribe" method="post" action="${esc(base)}/subscribe"><input type="email" name="email" required autocomplete="email" placeholder="you@example.com" aria-label="Email address"><button class="btn">${button}</button></form><p class="fine">${count(readers)}</p>`;

export const bar = (base: string, name: string, logo: string) =>
  `<header class="bar wrap"><a class="brand" href="${esc(base || "/")}">${logo ? `<img src="${esc(logo)}" alt="">` : ""}<span>${esc(name)}</span></a><a class="btn" href="${esc(base || "/")}#subscribe">Subscribe</a></header>`;

export const foot = (base: string) =>
  `<footer class="foot wrap"><span>Published with <a href="https://xem.email/?utm_source=newsletter-site"><b>Xem</b></a></span><a href="${esc(base)}/feed.xml">RSS feed</a></footer>`;

const note = (h: Home) =>
  h.note ? `<p class="note" role="status">${esc(h.note)}</p>` : "";
const href = (h: Home, i: Issue) => `${esc(h.base)}/${esc(i.id)}`;
const cover = (i: Issue, fallback: string) =>
  i.cover
    ? `<img src="${esc(i.cover)}" alt="" loading="lazy">`
    : `<span>${esc(fallback)}</span>`;
const initial = (i: Issue) => i.subject.trim().charAt(0).toUpperCase();

/** Centered hero and a grid of cover cards: Paper, Midnight and Signal. */
function cards(h: Home) {
  const items = h.issues
    .map(
      (i) =>
        `<li><a href="${href(h, i)}"><div class="cover">${cover(i, initial(i))}</div><time datetime="${esc(i.sentAt)}">${day(i.sentAt)}</time><h3>${esc(i.subject)}</h3></a></li>`,
    )
    .join("");
  return `${bar(h.base, h.name, h.logo)}<div class="band"><section class="hero wrap">${h.logo ? `<img class="logo" src="${esc(h.logo)}" alt="">` : ""}<h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}${note(h)}${subscribe(h.base, h.readers)}</section></div>
<main class="wrap issues"><h2 class="label">Latest issues</h2>${items ? `<ol class="cards">${items}</ol>` : `<p class="empty">The first issue is on its way.</p>`}</main>${foot(h.base)}`;
}

/** A newspaper front page: nameplate, lead story, columns. */
function newspaper(h: Home) {
  const [lead, ...rest] = h.issues;
  const today = new Date().toLocaleDateString("en", {
    weekday: "long",
    month: "long",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });
  const leadHtml = lead
    ? `<article class="lead"><a href="${href(h, lead)}">${lead.cover ? `<div class="lead-art"><img src="${esc(lead.cover)}" alt=""></div>` : ""}<div class="lead-copy"><p class="kicker">Latest issue</p><h2>${esc(lead.subject)}</h2><time datetime="${esc(lead.sentAt)}">${day(lead.sentAt)}</time><span class="more">Read the issue →</span></div></a></article>`
    : `<p class="empty">The first issue is on its way.</p>`;
  const cols = rest
    .map(
      (i) =>
        `<article><a href="${href(h, i)}">${i.cover ? `<img src="${esc(i.cover)}" alt="" loading="lazy">` : ""}<time datetime="${esc(i.sentAt)}">${day(i.sentAt)}</time><h3>${esc(i.subject)}</h3></a></article>`,
    )
    .join("");
  return `<header class="masthead wrap"><div class="topline"><span>${today}</span><span>${h.readers >= 10 ? `${h.readers.toLocaleString("en")} readers` : "Free to read"}</span><a href="#subscribe">Subscribe</a></div>
${h.logo ? `<img class="seal" src="${esc(h.logo)}" alt="">` : ""}<h1 class="nameplate">${esc(h.name)}</h1>${h.description ? `<p class="dateline">${esc(h.description)}</p>` : ""}</header>
<main class="wrap">${leadHtml}${cols ? `<div class="columns">${cols}</div>` : ""}
<section class="desk"><div><p class="kicker">Delivered free</p><h2>Get every edition in your inbox.</h2></div><div>${note(h)}${subscribe(h.base, h.readers)}</div></section></main>${foot(h.base)}`;
}

/** A magazine spread: sticky introduction beside a numbered index. */
function magazine(h: Home) {
  const total = h.issues.length;
  const items = h.issues
    .map(
      (i, n) =>
        `<li><a href="${href(h, i)}"><div class="thumb">${cover(i, initial(i))}</div><div class="meta"><span class="no">N° ${total - n}</span><time datetime="${esc(i.sentAt)}">${day(i.sentAt)}</time></div><h3>${esc(i.subject)}</h3></a></li>`,
    )
    .join("");
  return `<div class="spread wrap"><aside class="intro">${h.logo ? `<img class="logo" src="${esc(h.logo)}" alt="">` : ""}<p class="label">A newsletter</p><h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}${note(h)}${subscribe(h.base, h.readers)}</aside>
<main class="index">${items ? `<ol>${items}</ol>` : `<p class="empty">The first issue is on its way.</p>`}</main></div>${foot(h.base)}`;
}

/** A personal letter: one column of type, issues as dated lines. */
function letter(h: Home) {
  const rows = h.issues
    .map(
      (i) =>
        `<li><a href="${href(h, i)}"><span>${esc(i.subject)}</span><i></i><time datetime="${esc(i.sentAt)}">${day(i.sentAt)}</time></a></li>`,
    )
    .join("");
  return `<main class="sheet"><header>${h.logo ? `<img class="logo" src="${esc(h.logo)}" alt="">` : ""}<h1>${esc(h.name)}</h1>${h.description ? `<p class="lede">${esc(h.description)}</p>` : ""}${note(h)}${subscribe(h.base, h.readers, "Subscribe")}</header>
<section><h2>Archive</h2>${rows ? `<ol class="rows">${rows}</ol>` : `<p class="empty">The first letter is on its way.</p>`}</section></main>${foot(h.base)}`;
}

type Template = { label: string; home: (h: Home) => string; css: string };

export const templates: Record<string, Template> = {
  paper: {
    label: "Paper",
    home: cards,
    css: `:root{--bg:#f7f6f2;--surface:#fff;--ink:#1a1917;--muted:#6f6b64;--line:#e6e2da;--accent:#1a1917;--on-accent:#fff;--sans:Inter,ui-sans-serif,system-ui,sans-serif;--display:Fraunces,Georgia,serif}
.hero{display:flex;flex-direction:column;align-items:center;text-align:center;padding-top:80px;padding-bottom:72px}
.hero .logo{width:80px;height:80px;border-radius:20px;object-fit:cover;box-shadow:0 1px 2px rgb(0 0 0/.08),0 12px 32px rgb(0 0 0/.10)}
.hero h1{font:600 clamp(2.8rem,7.5vw,4.6rem)/1 var(--display);letter-spacing:-.035em;margin:26px 0 14px;font-variation-settings:"opsz" 144}
.hero .lede{color:var(--muted);font-size:1.2rem;max-width:560px;margin:0 0 32px}
.issues{padding-bottom:88px}.issues .label{padding-bottom:16px;border-bottom:1px solid var(--line);margin:0 0 28px}
.cards{list-style:none;margin:0;padding:0;display:grid;gap:36px 28px;grid-template-columns:repeat(auto-fill,minmax(250px,1fr))}
.cover{aspect-ratio:16/10;border-radius:14px;overflow:hidden;background:var(--surface);border:1px solid var(--line);display:grid;place-items:center;font:600 2.6rem var(--display);color:var(--muted);transition:transform .25s,box-shadow .25s}
.cover img{width:100%;height:100%;object-fit:cover}
.cards a:hover .cover{transform:translateY(-3px);box-shadow:0 14px 30px rgb(0 0 0/.10)}
.cards time{display:block;color:var(--muted);font-size:14px;margin-top:14px}
.cards h3{font:600 1.15rem/1.35 var(--sans);letter-spacing:-.01em;margin:2px 0 0}`,
  },
  broadsheet: {
    label: "Broadsheet",
    home: newspaper,
    css: `:root{--bg:#f3eee3;--surface:#fbf8f1;--ink:#151311;--muted:#5f574c;--line:#cfc4b0;--accent:#9c1c1f;--on-accent:#fff;--radius:2px;--sans:Inter,ui-sans-serif,system-ui,sans-serif;--display:Fraunces,Georgia,serif}
body{font-family:Newsreader,Georgia,serif;font-size:18px}
.masthead{text-align:center;padding-top:20px}
.topline{display:flex;justify-content:space-between;gap:12px;font:500 12px/1 var(--sans);letter-spacing:.12em;text-transform:uppercase;color:var(--muted);padding-bottom:14px;border-bottom:1px solid var(--ink)}
.topline a{color:var(--accent);font-weight:700}
.seal{width:56px;height:56px;border-radius:50%;object-fit:cover;margin:28px auto 0;filter:grayscale(.2)}
.nameplate{font:900 clamp(3rem,11vw,8rem)/.9 var(--display);letter-spacing:-.04em;margin:22px 0 14px;font-variation-settings:"opsz" 144,"SOFT" 0,"WONK" 1}
.dateline{font:italic 400 1.15rem Newsreader,Georgia,serif;color:var(--muted);margin:0;padding:12px 0;border-top:3px double var(--ink);border-bottom:3px double var(--ink)}
.lead a{display:grid;grid-template-columns:1.35fr 1fr;gap:36px;align-items:center;padding:40px 0;border-bottom:1px solid var(--ink)}
.lead-art{aspect-ratio:4/3;overflow:hidden}.lead-art img{width:100%;height:100%;object-fit:cover;filter:contrast(1.05) saturate(.9)}
.kicker{font:700 12px/1 var(--sans);letter-spacing:.16em;text-transform:uppercase;color:var(--accent);margin:0 0 12px}
.lead h2{font:800 clamp(2rem,4.4vw,3.4rem)/1.02 var(--display);letter-spacing:-.025em;margin:0 0 16px}
.lead time,.columns time{font:500 13px var(--sans);color:var(--muted);letter-spacing:.04em;text-transform:uppercase}
.more{display:block;margin-top:18px;font:600 15px var(--sans);color:var(--accent)}
.lead a:hover h2,.columns a:hover h3{text-decoration:underline;text-decoration-thickness:1px;text-underline-offset:4px}
.columns{display:grid;grid-template-columns:repeat(3,1fr);border-bottom:1px solid var(--ink)}
.columns article{padding:28px 24px;border-left:1px solid var(--line)}
.columns article:nth-child(3n+1){border-left:0;padding-left:0}.columns article:nth-child(3n){padding-right:0}
.columns img{aspect-ratio:3/2;width:100%;object-fit:cover;margin-bottom:16px;filter:saturate(.9)}
.columns h3{font:700 1.35rem/1.15 var(--display);letter-spacing:-.015em;margin:6px 0 0}
.desk{display:grid;grid-template-columns:1fr 1fr;gap:32px;align-items:center;margin:48px 0 72px;padding:36px;background:var(--surface);border:1px solid var(--ink)}
.desk h2{font:800 2rem/1.05 var(--display);letter-spacing:-.02em;margin:0}
.desk .fine{font-family:var(--sans)}
footer.foot{border-top:3px double var(--ink);font-family:var(--sans)}
.reader h1{font-weight:800}
@media (max-width:820px){.lead a,.desk{grid-template-columns:1fr}.columns{grid-template-columns:1fr}.columns article{border-left:0;padding:24px 0;border-top:1px solid var(--line)}}`,
  },
  midnight: {
    label: "Midnight",
    home: cards,
    css: `:root{--bg:#07090d;--surface:rgb(255 255 255/.04);--ink:#eef2f6;--muted:#8b96a3;--line:rgb(255 255 255/.09);--accent:#6ee7a8;--on-accent:#04140b;--sans:Inter,ui-sans-serif,system-ui,sans-serif;--display:Inter,ui-sans-serif,system-ui,sans-serif}
body{background:radial-gradient(1200px 520px at 50% -140px,rgb(110 231 168/.16),transparent 70%),radial-gradient(900px 480px at 85% 0,rgb(120 119 255/.14),transparent 70%),var(--bg);background-repeat:no-repeat}
.hero{display:flex;flex-direction:column;align-items:center;text-align:center;padding-top:88px;padding-bottom:80px}
.hero .logo{width:76px;height:76px;border-radius:20px;object-fit:cover;box-shadow:0 0 0 1px var(--line),0 18px 50px rgb(110 231 168/.18)}
.hero h1{font:700 clamp(2.8rem,8vw,5rem)/1 var(--display);letter-spacing:-.05em;margin:28px 0 16px;background:linear-gradient(180deg,#fff,#b9c4cf);-webkit-background-clip:text;background-clip:text;color:transparent}
.hero .lede{color:var(--muted);font-size:1.2rem;max-width:560px;margin:0 0 34px}
form.sub{backdrop-filter:blur(10px);box-shadow:0 0 0 1px var(--line),0 20px 60px rgb(0 0 0/.4)}
.issues{padding-bottom:96px}.issues .label{margin:0 0 28px}
.cards{list-style:none;margin:0;padding:0;display:grid;gap:22px;grid-template-columns:repeat(auto-fill,minmax(260px,1fr))}
.cards a{display:block;padding:12px 12px 20px;border-radius:20px;background:var(--surface);border:1px solid var(--line);transition:border-color .2s,transform .2s}
.cards a:hover{border-color:rgb(110 231 168/.45);transform:translateY(-2px)}
.cover{aspect-ratio:16/10;border-radius:12px;overflow:hidden;display:grid;place-items:center;background:linear-gradient(135deg,rgb(110 231 168/.18),rgb(120 119 255/.18));font:700 2.4rem var(--display);color:var(--ink)}
.cover img{width:100%;height:100%;object-fit:cover}
.cards time{display:block;color:var(--muted);font-size:13px;margin:16px 6px 0}
.cards h3{font:600 1.1rem/1.35 var(--sans);margin:4px 6px 0}
.reader iframe{border-color:transparent}`,
  },
  atelier: {
    label: "Atelier",
    home: magazine,
    css: `:root{--bg:#ede6dc;--surface:#f7f2ea;--ink:#231b16;--muted:#76685c;--line:#d6cbbd;--accent:#8a3b22;--on-accent:#fff;--radius:4px;--sans:Inter,ui-sans-serif,system-ui,sans-serif;--display:Fraunces,Georgia,serif}
.spread{display:grid;grid-template-columns:minmax(300px,5fr) 7fr;gap:64px;padding-top:56px;padding-bottom:72px}
.intro{position:sticky;top:48px;align-self:start}
.intro .logo{width:64px;height:64px;border-radius:50%;object-fit:cover;margin-bottom:28px}
.intro h1{font:400 clamp(3.2rem,7vw,5.6rem)/.92 var(--display);font-style:italic;letter-spacing:-.04em;margin:14px 0 22px;font-variation-settings:"opsz" 144,"SOFT" 100}
.intro .lede{font:400 1.25rem/1.5 Newsreader,Georgia,serif;color:var(--muted);margin:0 0 32px}
.index ol{list-style:none;margin:0;padding:0}
.index li{border-top:1px solid var(--ink)}.index li:first-child{border-top:0}
.index a{display:grid;grid-template-columns:1fr;gap:14px;padding:32px 0}
.thumb{aspect-ratio:16/9;overflow:hidden;background:var(--surface);display:grid;place-items:center;font:italic 400 4rem var(--display);color:var(--muted)}
.thumb img{width:100%;height:100%;object-fit:cover;transition:transform .6s}
.index a:hover .thumb img{transform:scale(1.03)}
.meta{display:flex;justify-content:space-between;font:500 12px/1 var(--sans);letter-spacing:.14em;text-transform:uppercase;color:var(--muted)}
.no{color:var(--accent);font-weight:700}
.index h3{font:500 clamp(1.6rem,3vw,2.3rem)/1.1 var(--display);letter-spacing:-.02em;margin:0}
.index a:hover h3{color:var(--accent)}
.reader h1{font-weight:400;font-style:italic}
@media (max-width:900px){.spread{grid-template-columns:1fr;gap:24px}.intro{position:static}}`,
  },
  letter: {
    label: "Letter",
    home: letter,
    css: `:root{--bg:#fffdf8;--surface:#fff;--ink:#28241f;--muted:#857c71;--line:#e9e2d6;--accent:#2f5d50;--on-accent:#fff;--radius:10px;--sans:Inter,ui-sans-serif,system-ui,sans-serif;--display:Newsreader,Georgia,serif}
body{font:400 19px/1.7 Newsreader,Georgia,serif}
.sheet{max-width:660px;margin:0 auto;padding:88px 24px 56px}
.sheet .logo{width:56px;height:56px;border-radius:50%;object-fit:cover;margin-bottom:28px}
.sheet h1{font:500 clamp(2.4rem,6vw,3.4rem)/1.05 var(--display);letter-spacing:-.025em;margin:0 0 18px}
.sheet .lede{font-size:1.3rem;color:var(--muted);margin:0 0 30px;font-style:italic}
form.sub{max-width:100%}
.sheet section{margin-top:72px}
.sheet h2{font:600 13px/1 var(--sans);letter-spacing:.16em;text-transform:uppercase;color:var(--muted);margin:0 0 12px}
.rows{list-style:none;margin:0;padding:0}
.rows a{display:flex;align-items:baseline;gap:12px;padding:14px 0;border-bottom:1px solid var(--line)}
.rows span{font-size:1.15rem}.rows a:hover span{color:var(--accent);text-decoration:underline;text-underline-offset:4px;text-decoration-thickness:1px}
.rows i{flex:1;border-bottom:1px dotted var(--line);transform:translateY(-6px)}
.rows time{font:400 14px var(--sans);color:var(--muted);white-space:nowrap}
footer.foot{max-width:660px;border-top:0}
.bar{max-width:780px}
.reader h1{font-weight:500}
@media (max-width:560px){.rows a{flex-direction:column;gap:2px}.rows i{display:none}}`,
  },
  signal: {
    label: "Signal",
    home: cards,
    css: `:root{--bg:#f5f5f7;--surface:#fff;--ink:#0d0d12;--muted:#636370;--line:#e4e4ea;--accent:#ff4f2e;--on-accent:#fff;--radius:14px;--sans:Inter,ui-sans-serif,system-ui,sans-serif;--display:Inter,ui-sans-serif,system-ui,sans-serif}
.bar{position:relative;z-index:1;color:#fff}.bar .btn{background:#fff;color:var(--ink)}
.band{margin-top:-74px;padding-top:74px;background:radial-gradient(800px 400px at 15% 0,#ffb547 0,transparent 60%),radial-gradient(900px 500px at 90% 20%,#ff2e93 0,transparent 60%),linear-gradient(135deg,#ff4f2e,#d9258f);color:#fff}
.hero{display:flex;flex-direction:column;align-items:flex-start;padding-top:72px;padding-bottom:160px}
.hero .logo{width:72px;height:72px;border-radius:18px;object-fit:cover;box-shadow:0 10px 30px rgb(0 0 0/.25)}
.hero h1{font:800 clamp(3rem,9vw,6.2rem)/.95 var(--display);letter-spacing:-.055em;margin:28px 0 18px;max-width:900px}
.hero .lede{font-size:1.3rem;max-width:600px;margin:0 0 34px;opacity:.92}
.hero form.sub{background:#fff;border-color:transparent;box-shadow:0 20px 50px rgb(80 0 40/.35)}
.hero .fine{color:#fff;opacity:.85}.hero .note{color:var(--ink)}
.issues{margin-top:-104px;padding-bottom:96px}.issues .label{color:#fff;margin:0 0 20px}
.cards{list-style:none;margin:0;padding:0;display:grid;gap:22px;grid-template-columns:repeat(auto-fill,minmax(270px,1fr))}
.cards a{display:block;height:100%;padding:12px 12px 22px;border-radius:22px;background:var(--surface);box-shadow:0 1px 2px rgb(0 0 0/.06),0 18px 40px rgb(13 13 18/.08);transition:transform .2s,box-shadow .2s}
.cards a:hover{transform:translateY(-4px);box-shadow:0 1px 2px rgb(0 0 0/.06),0 26px 50px rgb(13 13 18/.14)}
.cover{aspect-ratio:16/10;border-radius:14px;overflow:hidden;display:grid;place-items:center;background:linear-gradient(135deg,#ffe1d6,#ffd3ea);font:800 2.6rem var(--display);color:var(--accent)}
.cover img{width:100%;height:100%;object-fit:cover}
.cards time{display:block;color:var(--muted);font-size:13px;font-weight:500;margin:16px 6px 0}
.cards h3{font:700 1.2rem/1.3 var(--sans);letter-spacing:-.02em;margin:4px 6px 0}
.reader .back,.reader time{color:var(--muted)}
.end h2,.panel h1{font-weight:800;letter-spacing:-.04em}
@media (max-width:560px){.hero{padding-bottom:140px}}`,
  },
};

export const template = (name: string) => templates[name] || templates.paper;

export const stylesheet = (name: string) => base + template(name).css;
