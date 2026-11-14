import { fieldQuery, insertField, matchFields, usedFields } from "@/lib/email/fields";

test("{{ before the caret opens field suggestions, narrowed as you type", () => {
  expect(fieldQuery("Hi {{", 5)).toBe("");
  expect(fieldQuery("Hi {{ fir", 9)).toBe("fir");
  expect(fieldQuery("Hi {{first_name}} x", 19)).toBeUndefined();
  expect(fieldQuery("Hi there", 8)).toBeUndefined();
  expect(matchFields("fir").map((f) => f.key)).toEqual(["first_name"]);
  expect(matchFields("c").map((f) => f.key)).toEqual(["company", "city", "country"]);
  expect(matchFields("").length).toBe(9);
});

test("picking a field completes the token and puts the caret after it", () => {
  expect(insertField("Hi {{fi, welcome", 7, "first_name")).toEqual({ text: "Hi {{first_name}}, welcome", caret: 17 });
  expect(insertField("Hi {{fi}}!", 7, "first_name")).toEqual({ text: "Hi {{first_name}}!", caret: 17 });
});

test("used fields keep a fallback only when every use has one", () => {
  expect(usedFields("Hi {{ first_name | there }}", "<p>{{company}} {{city|}} {{first_name|friend}}</p>")).toEqual([
    { key: "first_name", fallback: "there" },
    { key: "company", fallback: undefined },
    { key: "city", fallback: "" },
  ]);
  expect(usedFields("{{name|you}}", "{{name}}")).toEqual([{ key: "name", fallback: undefined }]);
});
