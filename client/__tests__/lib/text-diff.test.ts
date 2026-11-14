import { diffLines, readableLines } from "@/lib/text-diff";

describe("issue web page diff", () => {
  it("reads an email as lines of text and images", () => {
    expect(
      readableLines(
        `<html><head><style>p{color:red}</style></head><body><h1>We shipped</h1><p>Tom &amp; Ana   fixed <b>it</b>.</p><img src="https://cdn.example.com/a.png" width="600"><table><tr><td>Thanks</td></tr></table></body></html>`,
      ),
    ).toEqual(["We shipped", "Tom & Ana fixed it.", "[Image: https://cdn.example.com/a.png]", "Thanks"]);
  });

  it("marks what was added and removed", () => {
    expect(diffLines(["Hello", "We shipped", "Bye"], ["Hello", "We shipped, and fixed a typo", "Bye"])).toEqual([
      { op: " ", text: "Hello" },
      { op: "-", text: "We shipped" },
      { op: "+", text: "We shipped, and fixed a typo" },
      { op: " ", text: "Bye" },
    ]);
    expect(diffLines([], ["New"])).toEqual([{ op: "+", text: "New" }]);
  });
});
