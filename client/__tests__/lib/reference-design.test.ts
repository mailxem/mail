/** @jest-environment jsdom */
// jsdom predates structuredClone; designs are plain JSON, so a JSON copy is equivalent here.
globalThis.structuredClone ??= ((v: unknown) => JSON.parse(JSON.stringify(v))) as typeof structuredClone;

import { applySlots, designImages, extractSlots, rankReferences, rowsOf } from "@/lib/assistant/reference-design";

const design = {
  body: {
    rows: [
      { id: "r1", columns: [{ contents: [{ id: "h1", type: "heading", values: { text: "Good things take time." } }] }] },
      {
        id: "r2",
        columns: [
          {
            contents: [
              { id: "t1", type: "text", values: { text: '<p style="line-height: 160%;"><span style="color: #333;">Old body</span></p>' } },
              { id: "logo", type: "text", values: { text: '<table><tr><td><img src="/x.png" /></td></tr></table>' } },
              { id: "b1", type: "button", values: { text: "Read more" } },
            ],
          },
        ],
      },
      { id: "r3", columns: [{ contents: [{ id: "t2", type: "text", values: { text: "<p>Products</p>" } }] }] },
    ],
  },
};

test("extracts writable copy and skips structural blocks", () => {
  const slots = extractSlots(design);
  expect(slots.map((s) => s.id)).toEqual(["h1", "t1", "b1", "t2"]);
  expect(slots[1]).toMatchObject({ kind: "text", text: "Old body" });
  expect(rowsOf(design)[1]).toEqual({ id: "r2", slots: ["t1", "logo", "b1"] });
});

test("applies new copy inside the original styling and drops removed rows", () => {
  const next = applySlots(design, {
    subject: "Hi",
    slots: { h1: "A deeper read", t1: "One <long> essay.\n\nEvery Sunday.", b1: "Read the essay", ghost: "ignored" },
    remove: ["r3"],
  });
  expect(next.body.rows.map((r) => r.id)).toEqual(["r1", "r2"]);
  const [t1, logo] = next.body.rows[1].columns[0].contents;
  expect(t1.values.text).toBe(
    '<p style="line-height: 160%;"><span style="color: #333;">One &lt;long&gt; essay.</span></p><p style="line-height: 160%;"><span style="color: #333;">Every Sunday.</span></p>',
  );
  expect(logo.values.text).toContain("<img");
  expect(design.body.rows[0].columns[0].contents[0].values.text).toBe("Good things take time.");
});

test("ranks library designs against the brief", () => {
  const [best] = rankReferences("weekly design newsletter editorial essay", 3);
  expect(best).toBeDefined();
});

test("sends the reference screenshot first, then examples, skipping missing ones", async () => {
  const [reference] = rankReferences("weekly newsletter", 6);
  const missing = rankReferences("weekly newsletter", Infinity).find((t) => t.collection === "Studio collection" && t.key !== reference.key)!.key;
  global.fetch = jest.fn(async (url: string) =>
    url.includes(missing)
      ? ({ ok: false } as Response)
      : ({ ok: true, blob: async () => new Blob(["x"], { type: "image/webp" }) } as Response),
  ) as never;
  const images = await designImages("weekly newsletter", reference);
  expect(images[0]).toMatchObject({ role: "reference", data: expect.stringMatching(/^data:image\/webp;base64,/) });
  // Examples are Studio templates and cite their public gallery page.
  for (const example of images.slice(1)) {
    expect(example.role).toBe("example");
    expect(example.source).toMatch(/, https:\/\/xem\.email\/templates\/[a-z-]+$/);
  }
  expect(images.length).toBeGreaterThan(1);
  expect(images.length).toBeLessThanOrEqual(4);
  expect((global.fetch as jest.Mock).mock.calls.some(([u]) => u.includes(`/${reference.key}/`))).toBe(true);
});

test("picks a reference of the kind of email the brief asks for", () => {
  expect(rankReferences("A password reset email. Link expires in 30 minutes.", 1)[0].category).toBe("Transactional");
  // An event mentioned inside a newsletter is still a newsletter.
  expect(rankReferences("Our monthly newsletter: a new coffee, a brewing tip and a pop-up on Saturday.", 1)[0].category).toBe("Newsletters");
  expect(rankReferences("A welcome email for new users with three first steps.", 1)[0].category).toBe("Welcome");
});
