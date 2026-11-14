import { contentChecks, visibleText } from "@/lib/email/content-checks";

const words = (n: number) => Array.from({ length: n }, (_, i) => `word${i}`).join(" ");
const body = (inner: string) =>
  `<html><head><style>p{color:red}</style></head><body>${inner}</body></html>`;
const ids = (subject: string, html: string) =>
  contentChecks({ subject, html }).map((w) => w.id);
const clean = body(`<p>${words(60)}</p><img src="a.png" alt="Our new shop">`);

test("a plain, well-written email has no warnings", () => {
  expect(ids("Your April update from the studio", clean)).toEqual([]);
});

test("subject mostly in capitals, ignoring contact fields", () => {
  expect(ids("HUGE SALE TODAY ONLY", clean)).toContain("caps-subject");
  expect(ids("Big news, {{FIRST_NAME}}", clean)).not.toContain("caps-subject");
  expect(ids("FAQ for May", clean)).not.toContain("caps-subject");
});

test("three or more exclamation marks in a row, in subject or body", () => {
  expect(ids("Hurry!!!", clean)).toContain("exclamations");
  expect(ids("Hi", body(`<p>Wow!!!! ${words(60)}</p>`))).toContain("exclamations");
  expect(ids("Hi! Hello!!", clean)).not.toContain("exclamations");
});

test("link shorteners", () => {
  const html = body(`<p>${words(60)} <a href="https://bit.ly/x1">a</a> <a href="https://www.tinyurl.com/y">b</a> <a href="https://example.com/bit.ly">c</a></p>`);
  const w = contentChecks({ subject: "Hi", html }).find((w) => w.id === "shorteners");
  expect(w?.title).toContain("bit.ly");
  expect(w?.title).toContain("www.tinyurl.com");
  expect(w?.title).not.toContain("example.com");
  expect(ids("Hi", body(`<p>${words(60)} <a href="https://notbit.ly/x">a</a></p>`))).not.toContain("shorteners");
});

test("image-only email and very little text", () => {
  const imageOnly = contentChecks({ subject: "Hi", html: body(`<img src="a.png" alt="Sale">`) });
  expect(imageOnly.find((w) => w.id === "little-text")?.title).toBe("The email is only images");
  expect(ids("Hi", body(`<img src="a.png" alt="x"><p>${words(30)}</p>`))).toContain("little-text");
  expect(ids("Hi", body(`<p>${words(10)}</p>`))).toContain("little-text");
  expect(ids("Hi", body(`<p>${words(25)}</p>`))).not.toContain("little-text");
  expect(ids("Hi", "")).toEqual([]);
});

test("images without alt text", () => {
  const html = body(`<p>${words(60)}</p><img src="a.png"><img src="b.png" alt=""><img alt="Logo" src="c.png"><img src=d.png alt=Logo>`);
  expect(contentChecks({ subject: "Hi", html }).find((w) => w.id === "alt-text")?.title).toBe("2 images have no alt text");
});

test("very long subject", () => {
  expect(ids("x".repeat(71), clean)).toContain("long-subject");
  expect(ids("x".repeat(70), clean)).not.toContain("long-subject");
});

test("every warning says what to change", () => {
  for (const w of contentChecks({ subject: "BUY NOW!!! ".repeat(8), html: body(`<img src="a.png"> <a href="https://bit.ly/a">go</a>`) }))
    expect(w.fix.length).toBeGreaterThan(20);
});

test("visible text skips head, styles and comments", () => {
  expect(visibleText(body(`<!-- hidden --><p>Hello&nbsp;there &amp; you</p>`))).toBe("Hello there you");
});
