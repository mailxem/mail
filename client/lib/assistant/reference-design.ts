import { templateStarters, type TemplateStarter } from "@/lib/template-starters/registry";

/**
 * Reference-based design: keep a library template's layout, fonts and colours,
 * and let the AI rewrite only the words. The model never sees or returns
 * markup; it gets numbered plain-text slots and answers with plain text.
 */

export type Slot = { id: string; kind: "heading" | "text" | "button"; text: string };
export type SlotAnswer = {
  subject: string;
  preheader?: string;
  slots: Record<string, string>;
  remove?: string[];
};
type Content = { id?: string; type: string; values: Record<string, unknown> };
type Row = { id?: string; columns: { contents: Content[] }[] };
type Design = { body: { rows: Row[] } };

const STOP = new Set(
  "a an and are as at be but by for from has have i in is it its me my of on or our so that the this to us we with you your email emails want make".split(" "),
);
const words = (s: string) =>
  s.toLowerCase().match(/[a-z0-9]+/g)?.filter((w) => w.length > 2 && !STOP.has(w)) ?? [];

// Which kind of email a brief asks for, so a coffee newsletter isn't built on a
// money-newsletter welcome just because both say "newsletter".
const KINDS: [string, RegExp][] = [
  ["Transactional", /\b(password|reset|sign[- ]?in|log ?in|verify|verification|confirm|code|receipt|invoice|order|security|privacy|magic link|account)\b/],
  ["Welcome", /\b(welcome|onboard\w*|sign(ed)? ?up|new (users?|members?|subscribers?)|getting started|first steps?)\b/],
  ["Newsletters", /\b(newsletter|digest|roundup|weekly|monthly|edition|issue)\b/],
  ["Events", /\b(event|webinar|conference|meetup|pop-?up|invite|invitation|rsvp|workshop)\b/],
  ["Commerce", /\b(sale|discount|shop|store|cart|collection|deal|black friday)\b/],
  ["Product", /\b(announce\w*|launch\w*|release|feature|update|changelog|pro plan|paid tier|upgrade)\b/],
];
const kindOf = (brief: string) => KINDS.find(([, re]) => re.test(brief.toLowerCase()))?.[0];

/** Library templates closest to a brief, best first, favouring the Studio collection. */
export function rankReferences(brief: string, limit = 6): TemplateStarter[] {
  const terms = new Set(words(brief));
  const kind = kindOf(brief);
  return [...templateStarters]
    .map((t) => {
      const hay = words([t.name, t.description, t.category, ...t.tags].join(" "));
      let score = 0;
      for (const w of hay) if (terms.has(w)) score += 1;
      if (kind && t.category === kind) score += 6;
      // Studio templates are the house style the designer should produce.
      // Reference designs can carry real brands in their images, which the AI
      // cannot change, so they only win on a much closer match.
      if (t.collection === "Studio collection") score += 6;
      else if (t.collection === "Reference designs") score += 1;
      return { t, score };
    })
    .sort((a, b) => b.score - a.score)
    .slice(0, limit)
    .map((x) => x.t);
}

// The hand-built templates published at xem.email/templates (website/lib/template-gallery.json).
// They are the house style, so the designer's example screenshots come from here.
export const GALLERY = new Set(
  "field-notes sunday-letter signal-digest community-roundup warm-welcome first-steps founders-note release-notes feature-spotlight new-collection weekend-edit saved-for-you gather-together live-session order-confirmed reset-password verify-email open-door".split(" "),
);
export const galleryUrl = (key: string) => (GALLERY.has(key) ? `https://xem.email/templates/${key}` : null);

// The reference sheet: Studio templates picked as the look the designer should
// aim for (from PR #50; absent keys are simply skipped until they land).
export const REFERENCE_SHEET = new Set(
  (
    // Welcome and milestones
    "studio-smooth-searching studio-on-the-list studio-you-did-it studio-first-commit studio-good-taste studio-ask-anything studio-speak-it studio-our-story " +
    // Account and transactional
    "studio-blog-welcome-survey studio-tldr-welcome studio-new-sign-in studio-shared-data studio-login-changed studio-confirm-code studio-id-on-phone studio-email-changed studio-privacy-update studio-almost-there-code studio-reset-camper studio-confirm-personal studio-magic-link studio-just-signed-in studio-reset-geometric"
  ).split(" "),
);
const isExample = (key: string) => REFERENCE_SHEET.has(key) || GALLERY.has(key);

export type DesignImage = { role: "reference" | "example"; data: string; source?: string };

async function previewData(key: string, signal?: AbortSignal) {
  try {
    const response = await fetch(`/assets/template-starters/${key}/preview.webp`, { signal });
    if (!response.ok) return null;
    const blob = await response.blob();
    if (blob.size > 290_000) return null;
    return await new Promise<string>((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(String(reader.result));
      reader.onerror = reject;
      reader.readAsDataURL(blob);
    });
  } catch {
    return null;
  }
}

/**
 * Screenshots the designer looks at: the chosen reference first (when there is
 * one), then the reference sheet and gallery templates closest to the brief as
 * house-style examples. At most four; missing previews are skipped.
 */
export async function designImages(brief: string, reference: TemplateStarter | null, signal?: AbortSignal) {
  const examples = rankReferences(brief, Infinity)
    .filter((t) => isExample(t.key) && t.key !== reference?.key)
    // Examples of the same kind as the reference come first.
    .sort((a, b) => Number(b.category === reference?.category) - Number(a.category === reference?.category))
    .slice(0, reference ? 2 : 3);
  const picks = [...(reference ? [{ role: "reference" as const, t: reference }] : []), ...examples.map((t) => ({ role: "example" as const, t }))];
  const loaded = await Promise.all(picks.map(async (p) => ({ role: p.role, data: await previewData(p.t.key, signal) })));
  return loaded
    .map((x, i) => {
      const url = galleryUrl(picks[i].t.key);
      return { ...x, source: url ? `${picks[i].t.name}, ${url}` : picks[i].t.name };
    })
    .filter((x): x is DesignImage & { source: string } => !!x.data);
}

function plain(html: string) {
  const doc = new DOMParser().parseFromString(html, "text/html");
  doc.querySelectorAll("br").forEach((br) => br.replaceWith("\n"));
  doc.querySelectorAll("p,div,li,h1,h2,h3,h4").forEach((el) => el.append("\n"));
  return (doc.body.textContent ?? "").replace(/\n{3,}/g, "\n\n").trim();
}
// Blocks with images, tables or lists are part of the design, not copy.
const structural = (html: string) => /<(img|table|ul|ol|iframe)\b/i.test(html);

function* contents(design: Design) {
  for (const row of design.body.rows)
    for (const column of row.columns) for (const content of column.contents) yield { row, content };
}

/** Writable copy in a design, as numbered plain-text slots. */
export function extractSlots(design: Design): Slot[] {
  const slots: Slot[] = [];
  for (const { content } of contents(design)) {
    if (!["heading", "text", "button"].includes(content.type) || !content.id) continue;
    const html = String(content.values.text ?? "");
    if (!html || structural(html)) continue;
    const text = plain(html);
    if (text) slots.push({ id: content.id, kind: content.type as Slot["kind"], text: text.slice(0, 1200) });
    if (slots.length >= 80) break;
  }
  return slots;
}

/** Row ids with their slot ids, so the model can drop a section that doesn't fit. */
export function rowsOf(design: Design) {
  return design.body.rows
    .filter((r) => r.id)
    .map((r) => ({
      id: r.id as string,
      slots: r.columns.flatMap((c) => c.contents.map((x) => x.id).filter(Boolean)) as string[],
    }));
}

const escape = (s: string) =>
  s.replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]!);

/** Rewrites a block's text inside its original first element, so inline styling survives. */
function rewrite(html: string, text: string) {
  const doc = new DOMParser().parseFromString(`<div id="x">${html}</div>`, "text/html");
  const root = doc.getElementById("x")!;
  const first = root.firstElementChild;
  const paragraphs = text.split(/\n{2,}/).map((p) => escape(p.trim()).replace(/\n/g, "<br />"));
  if (!first) return paragraphs.join("<br /><br />");
  const tag = first.tagName.toLowerCase();
  const attrs = [...first.attributes].map((a) => ` ${a.name}="${escape(a.value)}"`).join("");
  // Keep one inline wrapper (e.g. a styled <span>) when the block uses one.
  const inner = first.firstElementChild?.tagName === "SPAN" && first.childElementCount === 1 ? first.firstElementChild : null;
  const wrap = (p: string) =>
    inner ? `<span${[...inner.attributes].map((a) => ` ${a.name}="${escape(a.value)}"`).join("")}>${p}</span>` : p;
  if (["p", "h1", "h2", "h3", "h4", "div"].includes(tag))
    return paragraphs.map((p) => `<${tag}${attrs}>${wrap(p)}</${tag}>`).join("");
  return `<${tag}${attrs}>${paragraphs.join("<br /><br />")}</${tag}>`;
}

/** A copy of the reference with the model's copy applied. Unknown ids are ignored. */
export function applySlots<T extends Design>(design: T, answer: SlotAnswer): T {
  const next = structuredClone(design);
  const remove = new Set(answer.remove ?? []);
  next.body.rows = next.body.rows.filter((r) => !r.id || !remove.has(r.id));
  for (const { content } of contents(next)) {
    const text = content.id ? answer.slots[content.id] : undefined;
    if (typeof text !== "string" || !text.trim()) continue;
    content.values.text = rewrite(String(content.values.text ?? ""), text.trim());
  }
  return next;
}
