import * as React from "react";
import { renderToStaticMarkup } from "react-dom/server";
import {
  inferMailProvider,
  MailProviderIcon,
} from "@/components/marketing/mail-provider-icon";

describe("mail provider identity", () => {
  test.each([
    [{ host: "imap.gmail.com" }, "gmail", "Gmail"],
    [{ host: "IMAP.GOOGLEMAIL.COM." }, "gmail", "Gmail"],
    [{ provider: "GOOGLE_OAUTH", host: "custom.example" }, "gmail", "Gmail"],
    [{ host: "imap.mail.me.com" }, "icloud", "iCloud Mail"],
    [{ host: "imap.mail.icloud.com" }, "icloud", "iCloud Mail"],
    [{ provider: "CLOUDFLARE" }, "cloudflare", "Cloudflare"],
    [{ provider: "MANAGED", host: "example.com" }, "managed", "Xem inbox"],
  ])("recognizes trusted provider metadata %#", (input, kind, label) => {
    expect(inferMailProvider(input)).toEqual({ kind, label });
  });

  test.each([
    "evil-imap.gmail.com",
    "imap.gmail.com.attacker.example",
    "cloudflare.example",
    "imap.cloudflare.com",
    "mail.me.com.attacker.example",
  ])("does not infer a provider from an untrusted hostname: %s", (host) => {
    expect(inferMailProvider({ host })).toEqual({
      kind: "imap",
      label: "IMAP",
    });
  });

  it("renders an accessible icon and optional provider label", () => {
    const markup = renderToStaticMarkup(
      <MailProviderIcon host="imap.mail.me.com" showLabel />,
    );
    expect(markup).toContain('aria-label="iCloud Mail provider"');
    expect(markup).toContain("iCloud Mail");
    expect(markup).toContain("#1687F8");
  });

  it("uses the generic mail glyph for an ordinary IMAP mailbox", () => {
    const markup = renderToStaticMarkup(
      <MailProviderIcon host="mail.example.com" />,
    );
    expect(markup).toContain('aria-label="IMAP provider"');
    expect(markup).not.toContain("Cloudflare");
  });
});
