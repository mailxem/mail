import fs from "node:fs";
import path from "node:path";
import { templateStarters, filterTemplateStarters, loadStarterDesign, resolveStarterAssets } from "@/lib/template-starters/registry";

const asset = (url: string) => path.join(process.cwd(), "public", url);

describe("native starter library", () => {
  it("ships distinct editable compositions with consistent IDs, assets and a footer", () => {
    expect(templateStarters.length).toBeGreaterThanOrEqual(200);
    const keys = new Set<string>();
    const structures = new Set<string>();
    for (const starter of templateStarters) {
      expect(keys.has(starter.key)).toBe(false); keys.add(starter.key);
      const design = JSON.parse(fs.readFileSync(asset(starter.designUrl), "utf8"));
      const preview = fs.readFileSync(asset(starter.previewUrl), "utf8");
      expect(design.schemaVersion).toBe(18);
      expect(preview).toContain("Built with Xem");
      expect(preview).not.toMatch(/<script|onerror\s*=|onload\s*=|javascript:/i);
      const ids = new Set<string>(); const counts: Record<string, number> = {};
      const check = (item: any, type: string) => {
        expect(ids.has(item.id)).toBe(false); ids.add(item.id);
        counts[type] = (counts[type] || 0) + 1;
        expect(item.values._meta.htmlID).toBe(`${type}_${counts[type]}`);
      };
      for (const row of design.body.rows) {
        check(row, "u_row"); expect(row.cells).toHaveLength(row.columns.length);
        for (const column of row.columns) {
          check(column, "u_column");
          for (const block of column.contents) {
            check(block, `u_content_${block.type}`);
            expect(["heading", "text", "image", "button", "divider"]).toContain(block.type);
            if (block.type === "image") {
              expect(block.values.altText).toBeTruthy();
              const url = block.values.src.url;
              if (url.startsWith("/assets/")) expect(fs.existsSync(asset(url))).toBe(true);
              else expect(url.startsWith("https://")).toBe(true);
            }
          }
        }
      }
      expect(design.counters).toEqual(counts);
      const structure = JSON.stringify(design.body.rows.map((row: any) => [row.cells, row.columns.map((c: any) => c.contents.map((b: any) => b.type))]));
      expect(structures.has(structure)).toBe(false); structures.add(structure);
    }
  });
  it("filters by category, content and collection together", () => {
    const found = filterTemplateStarters("Newsletters", "coffee", "Editorial collection");
    expect(found.length).toBeGreaterThan(0);
    expect(found.every(s => s.category === "Newsletters" && s.collection === "Editorial collection")).toBe(true);
    expect(filterTemplateStarters("All templates", "  MYGREENHOUSE  ").map(s => s.key)).toContain("greenhouse-welcome");
    expect(filterTemplateStarters("Commerce", "", "Editorial collection")).toHaveLength(0);
  });
  it("resolves only bundled assets without mutating a cached design", () => {
    const design = { image: "/assets/template-starters/media/example.png", link: "https://example.com/x", text: "Hello" };
    expect(resolveStarterAssets(design, "https://xem.email")).toEqual({ ...design, image: "https://xem.email/assets/template-starters/media/example.png" });
    expect(design.image).toBe("/assets/template-starters/media/example.png");
  });
  it("never fetches arbitrary or unknown starter URLs", async () => {
    const fetchMock = jest.fn(); global.fetch = fetchMock;
    await expect(loadStarterDesign("https://evil.example/design.json")).rejects.toThrow("could not be found");
    expect(fetchMock).not.toHaveBeenCalled();
  });
  it("rejects failed and empty design responses", async () => {
    global.fetch = jest.fn().mockResolvedValue({ ok: false });
    await expect(loadStarterDesign("greenhouse-welcome")).rejects.toThrow("could not be loaded");
    global.fetch = jest.fn().mockResolvedValue({ ok: true, json: async () => ({ schemaVersion: 18, body: { rows: [] } }) });
    await expect(loadStarterDesign("greenhouse-welcome")).rejects.toThrow("invalid");
  });
});
