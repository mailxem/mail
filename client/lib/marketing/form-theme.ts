export type FormTheme = {
  preset: "light" | "dark" | "warm";
  backgroundColor: string;
  cardColor: string;
  textColor: string;
  buttonColor: string;
  buttonTextColor: string;
  logoUrl: string;
  font: "sans" | "serif" | "mono";
  corners: "rounded" | "square";
};
export const formPresets: Record<FormTheme["preset"], FormTheme> = {
  light: {
    preset: "light",
    backgroundColor: "#f5f2fa",
    cardColor: "#ffffff",
    textColor: "#302839",
    buttonColor: "#7254e3",
    buttonTextColor: "#ffffff",
    logoUrl: "",
    font: "sans",
    corners: "rounded",
  },
  dark: {
    preset: "dark",
    backgroundColor: "#15151c",
    cardColor: "#242430",
    textColor: "#f4f2ff",
    buttonColor: "#c4b5fd",
    buttonTextColor: "#211736",
    logoUrl: "",
    font: "sans",
    corners: "rounded",
  },
  warm: {
    preset: "warm",
    backgroundColor: "#f8f0e4",
    cardColor: "#fffcf5",
    textColor: "#423929",
    buttonColor: "#466447",
    buttonTextColor: "#ffffff",
    logoUrl: "",
    font: "serif",
    corners: "rounded",
  },
};
export function resolveFormTheme(value?: Partial<FormTheme>): FormTheme {
  const base = formPresets[value?.preset] || formPresets.light;
  const theme = { ...base };
  for (const key of [
    "backgroundColor",
    "cardColor",
    "textColor",
    "buttonColor",
    "buttonTextColor",
  ] as const) {
    if (/^#[\da-f]{6}$/i.test(value?.[key] || "")) theme[key] = value[key];
  }
  if (["sans", "serif", "mono"].includes(value?.font)) theme.font = value.font;
  if (["rounded", "square"].includes(value?.corners))
    theme.corners = value.corners;
  try {
    const url = new URL(value?.logoUrl);
    if (url.protocol === "https:" && !url.username && !url.password)
      theme.logoUrl = url.href;
  } catch {}
  return theme;
}
export const publicFormAction = (slug: string) =>
  `${(process.env.NEXT_PUBLIC_API_URL || "").replace(/\/api\/v1\/?$/, "").replace(/\/$/, "")}/public/forms/${encodeURIComponent(slug)}`;
const escapeHTML = (text: string) =>
  text.replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ]!,
  );
export function formHTMLSnippet(
  action: string,
  fields: {
    Label: string;
    FieldType: string;
    Required: boolean;
    mapToContactField: string;
  }[],
  button: string,
  consent: { mode: "required" | "optional" | "none"; label: string } = { mode: "required", label: "I agree to receive emails and understand I can unsubscribe at any time." },
  version?: number,
) {
  return `<form action="${escapeHTML(action)}" method="post">${Number.isInteger(version) && version! > 0 ? `\n  <input type="hidden" name="version" value="${version}">` : ""}\n${fields.map((f) => `  <label>${escapeHTML(f.Label)}\n    ${f.FieldType === "TEXTAREA" ? `<textarea name="${escapeHTML(f.mapToContactField)}" maxlength="2000"${f.Required ? " required" : ""}></textarea>` : `<input type="${f.FieldType === "EMAIL" ? "email" : f.FieldType === "PHONE" ? "tel" : f.FieldType === "NUMBER" ? "number" : f.FieldType === "DATE" ? "date" : "text"}"${f.FieldType === "NUMBER" ? ' step="any"' : ""} name="${escapeHTML(f.mapToContactField)}" maxlength="2000"${f.Required ? " required" : ""}>`}\n  </label>`).join("\n")}${consent.mode === "none" ? "" : `\n  <label><input type="checkbox" name="consent" value="true"${consent.mode === "required" ? " required" : ""}> ${escapeHTML(consent.label)}</label>`}\n  <div hidden aria-hidden="true"><label>Website <input name="website" tabindex="-1" autocomplete="off"></label></div>\n  <button type="submit">${escapeHTML(button || "Subscribe")}</button>\n</form>`;
}
