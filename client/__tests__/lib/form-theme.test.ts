import {
  formHTMLSnippet,
  resolveFormTheme,
  formPresets,
} from "@/lib/marketing/form-theme";

describe("hosted form branding", () => {
  it("keeps existing forms readable and fills partial dark themes", () => {
    expect(resolveFormTheme()).toEqual(formPresets.light);
    expect(
      resolveFormTheme({ preset: "dark", buttonColor: "#123456" }),
    ).toMatchObject({
      cardColor: formPresets.dark.cardColor,
      buttonColor: "#123456",
    });
  });
  it("rejects unsafe style values and logo protocols", () => {
    expect(
      resolveFormTheme({
        buttonColor: "red;display:none",
        logoUrl: "javascript:alert(1)",
      }),
    ).toMatchObject({
      buttonColor: formPresets.light.buttonColor,
      logoUrl: "",
    });
    expect(
      resolveFormTheme({ logoUrl: "https://user:pass@example.com/logo.png" })
        .logoUrl,
    ).toBe("");
  });
});
describe("custom HTML form example", () => {
  it("uses native field names, consent, and honeypot without a shared submission ID", () => {
    const html = formHTMLSnippet(
      "https://api.example.com/public/forms/test",
      [
        {
          Label: "Email <script>",
          FieldType: "EMAIL",
          Required: true,
          mapToContactField: "email",
        },
        {
          Label: "Message",
          FieldType: "TEXTAREA",
          Required: false,
          mapToContactField: "message",
        },
      ],
      "Join <now>",
    );
    expect(html).toContain('method="post"');
    expect(html).toContain(
      'type="email" name="email" maxlength="2000" required',
    );
    expect(html).toContain('<textarea name="message"');
    expect(html).toContain('name="consent" value="true" required');
    expect(html).toContain('name="website"');
    expect(html).toContain("Email &lt;script&gt;");
    expect(html).toContain("Join &lt;now&gt;");
    expect(html).not.toContain("requestId");
  });
});


test("native HTML forms preserve optional and absent consent", () => {
  const fields = [{ Label: "Email", FieldType: "EMAIL", Required: true, mapToContactField: "email" }];
  const optional = formHTMLSnippet("https://forms.example/public/forms/signup", fields, "Send", { mode: "optional", label: "Product <news>" });
  expect(optional).toContain('name="consent" value="true">');
  expect(optional).toContain("Product &lt;news&gt;");
  expect(optional).not.toContain('name="consent" value="true" required');
  expect(formHTMLSnippet("https://forms.example/public/forms/signup", fields, "Send", { mode: "none", label: "" })).not.toContain('name="consent"');
});


test("native HTML examples include revision and permit fractional number answers", () => {
  const fields = [{ Label: "Amount", FieldType: "NUMBER", Required: true, mapToContactField: "amount" }];
  const html = formHTMLSnippet("https://forms.example/public/forms/signup", fields, "Send", { mode: "none", label: "" }, 7);
  expect(html).toContain('<input type="hidden" name="version" value="7">');
  expect(html).toContain('type="number" step="any"');
});
