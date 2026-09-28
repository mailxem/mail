import fs from "node:fs/promises";
import sharp from "sharp";

// Original vector illustrations; no browser or external image service is needed.
const c = {
  cream: "#ffffef",
  ink: "#22251f",
  forest: "#164b3f",
  iris: "#5b3cc4",
  lavender: "#e7d8fa",
  lemon: "#edf09b",
  muted: "#5c655b",
};
const esc = (s) => s.replaceAll("&", "&amp;").replaceAll("<", "&lt;");
const rect = (x, y, w, h, fill, r = 18, stroke = "none") =>
  `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="${r}" fill="${fill}" stroke="${stroke}" stroke-width="2"/>`;
const text = (x, y, s, size = 26, fill = c.ink, serif = false) =>
  `<text x="${x}" y="${y}" font-family="${serif ? "Georgia, serif" : "Arial, sans-serif"}" font-size="${size}" fill="${fill}">${esc(s)}</text>`;
const line = (x1, y1, x2, y2, color = c.forest) =>
  `<path d="M${x1} ${y1}L${x2} ${y2}" fill="none" stroke="${color}" stroke-width="3"/>`;
const arrow = (x, y, color = c.forest) =>
  `<path d="M${x} ${y}v28m-7-8 7 8 7-8" fill="none" stroke="${color}" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>`;
const svg = (w, h, title, desc, body) =>
  `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}" role="img" aria-labelledby="title desc"><title id="title">${esc(title)}</title><desc id="desc">${esc(desc)}</desc>${body}</svg>`;
const heading = (kicker, title) =>
  text(36, 48, kicker, 17, c.forest) + text(36, 96, title, 36, c.ink, true);
const node = (y, n, title, detail, fill = c.cream) =>
  rect(116, y, 488, 102, fill, 18, c.forest) +
  text(137, y + 35, n, 16, c.forest) +
  text(175, y + 39, title, 27) +
  text(137, y + 77, detail, 22, c.muted);

const flow = svg(
  640,
  1020,
  "One message. A path you can inspect.",
  "Campaigns and SMTP clients pass Xem checks, enter PostgreSQL, and go to SES and the recipient server. SES events return through SNS to delivery history and suppression.",
  rect(0, 0, 640, 1020, "#eef0e7") +
    heading("02 / THE DELIVERY PATH", "One message. A clear path.") +
    node(
      128,
      "01",
      "Campaign or SMTP client",
      "Your content, your verified sender",
    ) +
    arrow(360, 230) +
    node(268, "02", "Xem policy checks", "Workspace · domain · limits") +
    arrow(360, 370) +
    node(
      408,
      "03",
      "PostgreSQL outbox",
      "Store first. Then acknowledge.",
      c.lemon,
    ) +
    arrow(360, 510) +
    node(
      548,
      "04",
      "Amazon SES API",
      "The delivery infrastructure",
      c.lavender,
    ) +
    arrow(360, 650) +
    node(688, "05", "Recipient mail server", "Acceptance ≠ inbox placement") +
    `<path d="M116 599H63V883H116m-9-7 9 7-9 7" fill="none" stroke="${c.iris}" stroke-width="3" stroke-dasharray="7 6"/>` +
    rect(116, 820, 488, 126, c.iris) +
    text(138, 859, "SES events → Amazon SNS", 26, c.cream) +
    text(138, 899, "Verify signatures. Record outcomes.", 22, c.cream) +
    text(138, 929, "Update suppression before later sends.", 22, c.cream) +
    text(
      36,
      984,
      "Illustrative architecture · not measured traffic",
      19,
      c.muted,
    ),
);

const domain = svg(
  640,
  840,
  "One brand. Three different jobs.",
  "The From address is visible to readers. A dedicated bounce subdomain uses SES MAIL FROM records. Replies and the existing receiving MX stay with the mailbox provider.",
  rect(0, 0, 640, 840, "#eee8f5") +
    heading("01 / YOUR DOMAIN", "One brand. Three different jobs.") +
    rect(36, 130, 568, 170, c.cream, 22) +
    text(60, 166, "WHAT READERS SEE", 17, c.iris) +
    text(60, 215, "hello@example.com", 34, c.ink, true) +
    text(60, 255, "Your visible From address.", 25) +
    text(60, 284, "DKIM + DMARC help authenticate it.", 22, c.muted) +
    rect(36, 330, 568, 170, c.cream, 22) +
    text(60, 366, "WHERE DELIVERY FAILURES GO", 17, c.iris) +
    text(60, 415, "bounce.example.com", 34, c.ink, true) +
    text(60, 455, "Dedicated custom MAIL FROM domain.", 25) +
    text(60, 484, "SES MX + SPF live on this subdomain.", 22, c.muted) +
    rect(36, 530, 568, 170, c.forest, 22) +
    text(60, 566, "WHERE THE CONVERSATION CONTINUES", 17, c.lemon) +
    text(60, 615, "Your existing mailbox", 34, c.cream, true) +
    text(60, 655, "Keep your receiving MX records.", 25, c.cream) +
    text(60, 684, "Replies follow your Reply-To address.", 22, c.cream) +
    text(36, 757, "SMTP host = where a tool submits mail.", 24) +
    text(36, 794, "It does not have to match your From domain.", 22, c.muted),
);

const states = svg(
  640,
  870,
  "A status is a promise. Keep it precise.",
  "Queued proves database acceptance, accepted proves SES acceptance, delivered proves recipient-server acceptance. Inbox placement and reading are not proved by these events.",
  rect(0, 0, 640, 870, "#eef0e7") +
    heading("03 / WHAT THE STATUS TELLS YOU", "Three signals. No guesswork.") +
    line(66, 162, 66, 585, c.forest) +
    [
      [
        145,
        "1",
        "QUEUED",
        "Xem stored the message.",
        "SMTP acceptance follows the DB commit.",
      ],
      [
        310,
        "2",
        "ACCEPTED",
        "SES accepted the submission.",
        "Delivery is still a separate event.",
      ],
      [
        475,
        "3",
        "DELIVERED",
        "The recipient server accepted it.",
        "Based on authenticated SES feedback.",
      ],
    ]
      .map(
        ([y, n, label, detail, note]) =>
          `<circle cx="66" cy="${y + 26}" r="22" fill="${c.forest}"/>` +
          text(59, y + 34, n, 24, c.cream) +
          text(112, y + 17, label, 18, c.forest) +
          text(112, y + 59, detail, 27, c.ink, true) +
          text(112, y + 96, note, 21, c.muted),
      )
      .join("") +
    rect(36, 650, 568, 162, c.lavender, 20, c.iris) +
    text(60, 687, "NOT PROVED BY THESE EVENTS", 17, c.iris) +
    text(60, 734, "Inbox placement. Human attention.", 29, c.ink, true) +
    text(60, 776, "A delivery event cannot tell you either.", 23) +
    text(
      36,
      847,
      "Illustrative states · not a conversion-rate chart",
      19,
      c.muted,
    ),
);

const cover = svg(
  1200,
  800,
  "Your domain. A clear path.",
  "Xem managed sending connects a verified sender to a durable queue, Amazon SES, and signed feedback.",
  rect(0, 0, 1200, 800, c.cream, 0) +
    text(75, 69, "xem / THE JOURNAL", 23, c.forest) +
    text(967, 69, "BUILD NOTES", 18, c.forest) +
    line(75, 95, 1125, 95, "#164b3f40") +
    text(75, 291, "Your domain.", 76, c.ink, true) +
    text(75, 382, "A clear path.", 76, c.ink, true) +
    text(78, 445, "Inside Xem managed SMTP", 26, c.muted) +
    rect(78, 490, 272, 44, c.lavender, 22) +
    text(99, 519, "CUSTOM DOMAINS + SES", 18, c.iris) +
    rect(654, 156, 472, 494, "#e3e9d9", 36) +
    rect(702, 193, 376, 125, c.cream, 16, c.forest) +
    `<path d="M705 199L890 273L1074 199" fill="none" stroke="${c.forest}" stroke-width="3"/>` +
    text(744, 300, "hello@example.com", 23, c.forest) +
    arrow(890, 322) +
    rect(713, 370, 354, 72, c.lemon, 16, c.forest) +
    text(746, 415, "Xem · durable outbox", 27) +
    arrow(890, 450) +
    rect(713, 497, 354, 72, c.iris, 16) +
    text(789, 542, "Amazon SES", 28, c.cream) +
    `<path d="M713 534H680V614H744m-8-6 8 6-8 6" fill="none" stroke="${c.iris}" stroke-width="3" stroke-dasharray="7 6"/>` +
    text(772, 620, "Signed feedback", 24, c.iris) +
    line(75, 711, 1125, 711, "#164b3f40") +
    text(75, 753, "DELIVERABILITY / OPEN SOURCE", 18, c.forest) +
    text(769, 753, "Every email, a little more human.", 19, c.muted),
);

await fs.mkdir("design/blog/svg", { recursive: true });
for (const [name, art] of [
  ["managed-smtp-flow", flow],
  ["managed-smtp-domain-map", domain],
  ["managed-smtp-states", states],
]) {
  await fs.writeFile(`public/images/blog/${name}.svg`, art);
}
await fs.writeFile("design/blog/svg/managed-smtp-custom-domains.svg", cover);
await sharp(Buffer.from(cover))
  .resize(1600, 1067)
  .webp({ quality: 91, effort: 6 })
  .toFile("public/images/blog/managed-smtp-custom-domains.webp");
console.log("Rendered managed SMTP cover and three original diagrams.");
