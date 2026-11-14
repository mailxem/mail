/** One line of a diff: kept, added (+) or removed (-). */
export type DiffLine = { op: " " | "+" | "-"; text: string };

const entities: Record<string, string> = {
  amp: "&",
  lt: "<",
  gt: ">",
  quot: '"',
  "#39": "'",
  nbsp: " ",
};

/**
 * The readable lines of an email or web page: its text, a line per block,
 * with images as "[Image: address]" so a swapped picture shows up too.
 */
export function readableLines(html: string): string[] {
  return html
    .replace(/<(script|style|head)\b[\s\S]*?<\/\1>/gi, "")
    .replace(/<img\b[^>]*?\ssrc\s*=\s*["']([^"']+)["'][^>]*>/gi, "\n[Image: $1]\n")
    .replace(/<\/?(p|div|h[1-6]|li|tr|td|th|br|table|section|article|blockquote|ul|ol)\b[^>]*>/gi, "\n")
    .replace(/<[^>]+>/g, "")
    .replace(/&(amp|lt|gt|quot|#39|nbsp);/g, (_, e: string) => entities[e])
    .split("\n")
    .map((l) => l.replace(/\s+/g, " ").trim())
    .filter(Boolean);
}

/** A line diff of a and b, from their longest common subsequence. */
export function diffLines(a: string[], b: string[]): DiffLine[] {
  // ponytail: O(n·m) table; fine for an issue's few hundred lines.
  const n = a.length;
  const m = b.length;
  const lcs = Array.from({ length: n + 1 }, () => new Array<number>(m + 1).fill(0));
  for (let i = n - 1; i >= 0; i--)
    for (let j = m - 1; j >= 0; j--)
      lcs[i][j] = a[i] === b[j] ? lcs[i + 1][j + 1] + 1 : Math.max(lcs[i + 1][j], lcs[i][j + 1]);
  const out: DiffLine[] = [];
  let i = 0;
  let j = 0;
  while (i < n && j < m) {
    if (a[i] === b[j]) {
      out.push({ op: " ", text: a[i] });
      i++;
      j++;
    } else if (lcs[i + 1][j] >= lcs[i][j + 1]) out.push({ op: "-", text: a[i++] });
    else out.push({ op: "+", text: b[j++] });
  }
  while (i < n) out.push({ op: "-", text: a[i++] });
  while (j < m) out.push({ op: "+", text: b[j++] });
  return out;
}
