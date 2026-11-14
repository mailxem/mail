import {
  describeRule,
  designRows,
  keepRules,
  ruleProblem,
  splitRules,
  withRules,
} from "@/lib/sections";

const design = {
  body: {
    rows: [
      { values: { _meta: { htmlID: "u_row_1" } }, columns: [{ contents: [{ type: "image", values: {} }] }] },
      { values: { _meta: { htmlID: "u_row_2" } }, columns: [{ contents: [{ type: "text", values: { text: "<p>VIP&nbsp;offer</p>" } }] }] },
    ],
  },
};

test("rows are listed with what they hold", () => {
  expect(designRows(design)).toEqual([
    { id: "u_row_1", index: 0, summary: "Image" },
    { id: "u_row_2", index: 1, summary: "VIP offer" },
  ]);
});

test("only limited rules on rows that still exist are saved, cleaned up", () => {
  const rules = {
    u_row_1: { name: "  ", tags: ["vip"] }, // for everyone: nothing to save
    u_row_2: { name: " Offer ", audience: "only" as const, tags: ["vip"], show: "" as const },
    u_row_9: { audience: "except" as const, tags: ["vip"] }, // row deleted
  };
  expect(keepRules(rules, design)).toEqual({
    u_row_2: { name: "Offer", audience: "only", tags: ["vip"], segments: [] },
  });
  const saved = withRules<Record<string, unknown>>(design, { u_row_1: { show: "web", tags: ["left over"] } });
  expect(saved.xemSections).toEqual({ u_row_1: { show: "web" } });
  expect(splitRules(saved)).toEqual({ design, rules: { u_row_1: { show: "web" } } });
});

test("rules read as sentences and need someone to be for", () => {
  expect(describeRule({ audience: "except", tags: ["vip"], show: "email" }, ["vip"], ["Churned"])).toBe(
    "Everyone except readers with tag vip or segment Churned, in the email only",
  );
  expect(describeRule({ show: "web" })).toBe("Everyone, on the website only");
  expect(ruleProblem({ audience: "only" })).toMatch(/tag or segment/);
  expect(ruleProblem({ audience: "only", segments: ["s"] })).toBe("");
});
