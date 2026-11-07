// Terminal: a newsletter for builders, read like a shell session.

import {
  esc,
  foot,
  href,
  no,
  note,
  part,
  reveal,
  type Home,
  type Template,
} from "./base";

function home(h: Home) {
  const slug =
    h.name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "") || "newsletter";
  const rows = h.issues
    .map(
      (i, n) =>
        `<li ${reveal(n)}><a href="${href(h, i)}"><span class="perm">-rw-r--r--</span><span class="dim">${part(i.sentAt, { year: "numeric", month: "2-digit", day: "2-digit" })}</span><span class="dim">#${no(h, n)}</span><span class="name">${esc(i.subject)}</span></a></li>`,
    )
    .join("");
  const meta = [
    h.cadence,
    h.readers >= 10 ? `${h.readers.toLocaleString("en")} readers` : "",
  ]
    .filter(Boolean)
    .join(" · ");
  return `<header class="nav wrap"><a class="brand" href="${esc(h.base || "/")}"><span class="g">~</span>/${esc(slug)}</a><nav><a href="#issues">ls</a><a href="${esc(h.base)}/feed.xml">rss</a><a href="#subscribe" class="btn sm">subscribe</a></nav></header>
<main class="wrap"><section class="hero" ${reveal(0)}><p class="dim">Last login: ${new Date().toUTCString().slice(0, 16)} on ttys001</p><p><span class="g">$</span> cat README.md</p><h1>${esc(h.name)}<span class="cursor"></span></h1>${h.description ? `<p class="lede"># ${esc(h.description)}</p>` : ""}${meta ? `<p class="dim">// ${esc(meta)}</p>` : ""}
<p class="cmd"><span class="g">$</span> subscribe --email</p>${note(h)}<form class="sub" id="subscribe" method="post" action="${esc(h.base)}/subscribe"><span class="g">›</span><input type="email" name="email" required autocomplete="email" placeholder="you@example.com" aria-label="Email address"><button class="btn">[ enter ]</button></form><p class="dim fine">→ confirms by email. unsubscribe anytime.</p></section>
<section class="ls" id="issues"><p class="cmd"><span class="g">$</span> ls -l issues/ <span class="dim">· total ${h.issues.length}</span></p>${rows ? `<ol>${rows}</ol>` : `<p class="dim">issues/ is empty. first issue soon.</p>`}</section></main>${foot(h.base, h.name, h.description)}`;
}

export const terminal: Template = {
  label: "Terminal",
  home,
  css: `:root{--bg:#0b0e0b;--surface:#121712;--field:transparent;--ink:#d6f5cf;--muted:#6f8a6b;--line:rgb(214 245 207/.14);--accent:#7dff6a;--on-accent:#061006;--radius:0;--sans:"Geist Mono",ui-monospace,monospace;--mono:"Geist Mono",ui-monospace,monospace;--head:"Geist Mono",monospace;--head-weight:500;--head-track:-.03em}
body{font:400 15px/1.7 var(--mono);background-image:repeating-linear-gradient(0deg,rgb(255 255 255/.018) 0 1px,transparent 1px 3px)}
.g{color:var(--accent)}.dim{color:var(--muted)}
.nav nav{font-family:var(--mono)}
.nav .btn{border:1px solid var(--accent);background:transparent;color:var(--accent)}
.nav .btn:hover{background:var(--accent);color:var(--on-accent);opacity:1}
.hero{padding:64px 0 72px;max-width:820px}
.hero p{margin-bottom:6px}
.hero h1{font:600 clamp(2.6rem,8vw,5.6rem)/1 var(--mono);letter-spacing:-.06em;color:var(--accent);margin:26px 0 22px;text-shadow:0 0 28px rgb(125 255 106/.35);word-break:break-word}
.cursor{display:inline-block;width:.5em;height:.82em;margin-left:.08em;background:var(--accent);vertical-align:-.05em;animation:blink 1.1s steps(1) infinite}
@keyframes blink{50%{opacity:0}}
@media (prefers-reduced-motion:reduce){.cursor{animation:none}}
.lede{font-size:1.05rem;color:var(--ink);max-width:62ch}
.cmd{margin-top:40px}
form.sub{gap:12px;max-width:620px;margin-top:10px;padding:10px 6px 10px 16px;border:1px dashed var(--line);border-radius:0;background:var(--surface)}
form.sub input{font:400 15px var(--mono);color:var(--accent)}
form.sub:focus-within{border-color:var(--accent);box-shadow:none}
form.sub .btn{background:transparent;color:var(--accent);height:34px;font:600 14px var(--mono);padding:0 10px}
form.sub .btn:hover{background:var(--accent);color:var(--on-accent);opacity:1}
.fine{margin-top:10px}
.ls{padding-bottom:112px}
.ls ol{list-style:none;margin:14px 0 0;padding:0;border-top:1px solid var(--line)}
.ls a{display:grid;grid-template-columns:auto auto auto 1fr;gap:24px;padding:12px 6px;border-bottom:1px solid var(--line)}
.ls a:hover{background:rgb(125 255 106/.07)}
.ls a:hover .name{color:var(--accent)}
.perm{color:#4f6b4c}
.name{color:var(--ink);text-decoration:underline;text-decoration-color:var(--line);text-underline-offset:4px}
.foot{font-family:var(--mono)}
.reader h1{color:var(--accent)}.reader iframe{border:1px dashed var(--line);border-radius:0}
@media (max-width:700px){.ls a{grid-template-columns:auto 1fr;gap:4px 14px}.perm{display:none}.name{grid-column:1/-1}}`,
};
