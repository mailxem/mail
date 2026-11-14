// Conditional sections: a template editor row kept for some readers, or for
// the email or the newsletter website only. The server decides who gets each
// one when it renders (server/internal/utils/sections.go); the rules live in
// the design as `xemSections`, by row ID.

export type SectionRule = {
  name?: string;
  audience?: "" | "only" | "except";
  tags?: string[];
  segments?: string[];
  show?: "" | "email" | "web";
};

export type SectionRules = Record<string, SectionRule>;

type Design = {
  body?: { rows?: { values?: { _meta?: { htmlID?: string } }; columns?: { contents?: { type?: string; values?: Record<string, unknown> }[] }[] }[] };
  xemSections?: SectionRules;
  [key: string]: unknown;
};

/** Whether a rule changes anything: a row for everyone, everywhere, needs none. */
export const isLimited = (rule?: SectionRule) =>
  !!rule && (!!rule.audience || !!rule.show);

/** The rows of a design, each with a short description of what it holds. */
export function designRows(design: unknown) {
  const rows = (design as Design | undefined)?.body?.rows ?? [];
  return rows.map((row, index) => {
    const id = row.values?._meta?.htmlID ?? "";
    const parts: string[] = [];
    for (const column of row.columns ?? [])
      for (const content of column.contents ?? []) {
        const text = content.values?.text;
        if (typeof text === "string") parts.push(text);
        else if (content.type === "image") parts.push("Image");
        else if (content.type) parts.push(content.type[0].toUpperCase() + content.type.slice(1));
      }
    const summary = parts
      .join(" · ")
      .replace(/<[^>]*>/g, " ")
      .replace(/&nbsp;/g, " ")
      .replace(/\s+/g, " ")
      .trim();
    return { id, index, summary: summary.slice(0, 90) || "Empty row" };
  });
}

/** The rules worth keeping: limited ones, on rows the design still has. */
export function keepRules(rules: SectionRules, design: unknown): SectionRules {
  const ids = new Set(designRows(design).map((row) => row.id));
  const kept: SectionRules = {};
  for (const [id, rule] of Object.entries(rules)) {
    if (!ids.has(id) || !isLimited(rule)) continue;
    const { name, audience, tags, segments, show } = rule;
    kept[id] = {
      ...(name?.trim() ? { name: name.trim() } : {}),
      ...(audience ? { audience, tags: tags ?? [], segments: segments ?? [] } : {}),
      ...(show ? { show } : {}),
    };
  }
  return kept;
}

/** A design with its section rules, as it is saved. */
export function withRules<T>(design: T, rules: SectionRules): T {
  return { ...(design as object), xemSections: keepRules(rules, design) } as T;
}

/** A design split into what the editor loads and its section rules. */
export function splitRules(design: unknown) {
  const { xemSections, ...rest } = (design ?? {}) as Design;
  return { design: rest, rules: xemSections ?? {} };
}

/** What a rule means, in a sentence. */
export function describeRule(
  rule: SectionRule,
  tagNames: string[] = rule.tags ?? [],
  segmentNames: string[] = [],
) {
  const who = [...tagNames.map((t) => `tag ${t}`), ...segmentNames.map((s) => `segment ${s}`)].join(" or ");
  const audience =
    rule.audience === "only"
      ? `Only readers with ${who}`
      : rule.audience === "except"
        ? `Everyone except readers with ${who}`
        : "Everyone";
  const place =
    rule.show === "email" ? ", in the email only" : rule.show === "web" ? ", on the website only" : "";
  return audience + place;
}

/** Why a rule can't be saved, or "". */
export function ruleProblem(rule: SectionRule) {
  if (rule.audience && !(rule.tags?.length || rule.segments?.length))
    return "Pick at least one tag or segment.";
  if ((rule.name ?? "").length > 80) return "Names are up to 80 characters.";
  return "";
}
