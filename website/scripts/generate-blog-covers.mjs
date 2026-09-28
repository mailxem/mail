import fs from "node:fs/promises";
import path from "node:path";
import { chromium } from "playwright";
import sharp from "sharp";

// Original Xem editorial artwork. SVG source is retained for future editing.
const C = {
  cream: "#ffffef",
  iris: "#5b3cc4",
  ink: "#22251f",
  lemon: "#edf09b",
  forest: "#164b3f",
  lavender: "#e7d8fa",
  muted: "#66695e",
};
const esc = (s) => String(s).replaceAll("&", "&amp;").replaceAll("<", "&lt;");
const rect = (x, y, w, h, fill, r = 0, stroke = "none") =>
  `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="${r}" fill="${fill}" stroke="${stroke}" stroke-width="2"/>`;
const line = (x1, y1, x2, y2, color = C.ink, width = 2) =>
  `<path d="M${x1} ${y1}L${x2} ${y2}" fill="none" stroke="${color}" stroke-width="${width}" stroke-linecap="round"/>`;
const circle = (x, y, r, fill, stroke = "none") =>
  `<circle cx="${x}" cy="${y}" r="${r}" fill="${fill}" stroke="${stroke}" stroke-width="2"/>`;
const text = (x, y, s, size = 24, color = C.ink, serif = false, extra = "") =>
  `<text x="${x}" y="${y}" fill="${color}" font-family="${serif ? "EB Garamond" : "DM Sans"}" font-size="${size}" ${extra}>${esc(s)}</text>`;
const title = (x, y, lines, size = 88, color = C.ink) =>
  lines
    .map((s, i) =>
      text(x, y + i * size * 0.91, s, size, color, true, 'letter-spacing="-3"'),
    )
    .join("");
const label = (x, y, s, color = C.ink) =>
  text(x, y, s, 15, color, false, 'letter-spacing="2.3" font-weight="500"');
const group = (x, y, angle, body) =>
  `<g transform="translate(${x} ${y}) rotate(${angle})">${body}</g>`;
const star = (x, y, r = 35, fill = C.iris) =>
  `<path d="M${x} ${y - r}Q${x} ${y} ${x + r} ${y}Q${x} ${y} ${x} ${y + r}Q${x} ${y} ${x - r} ${y}Q${x} ${y} ${x} ${y - r}" fill="${fill}"/>`;
const envelope = (x, y, w, fill = C.cream, stroke = C.ink) =>
  group(
    x,
    y,
    0,
    rect(0, 0, w, w * 0.65, fill, 12, stroke) +
      `<path d="M2 4L${w / 2} ${w * 0.36}L${w - 2} 4M2 ${w * 0.65 - 4}L${w * 0.34} ${w * 0.32}M${w - 2} ${w * 0.65 - 4}L${w * 0.66} ${w * 0.32}" fill="none" stroke="${stroke}" stroke-width="2"/>`,
  );
const mail = (
  x,
  y,
  w,
  h,
  accent = C.iris,
  heading = "A little good in your inbox.",
) =>
  group(
    x,
    y,
    0,
    rect(9, 12, w, h, C.ink + "16", 12) +
      rect(0, 0, w, h, C.cream, 12, C.ink) +
      circle(22, 22, 4, accent) +
      text(35, 27, "xem", 17, C.ink, false, 'font-weight="600"') +
      line(18, 42, w - 18, 42, C.ink + "33") +
      text(20, 78, heading, 22, C.ink, true) +
      rect(20, 100, w - 40, Math.max(35, h - 190), accent, 5) +
      line(20, h - 66, w - 35, h - 66, C.ink + "66", 3) +
      line(20, h - 52, w * 0.63, h - 52, C.ink + "44", 3) +
      rect(20, h - 30, 76, 14, accent, 7),
  );
const pill = (x, y, s, fill = C.lemon, color = C.ink, w = 160) =>
  rect(x, y, w, 38, fill, 19) +
  text(x + w / 2, y + 25, s, 14, color, false, 'text-anchor="middle"');
const pathStroke = (d, color = C.iris, width = 3) =>
  `<path d="${d}" fill="none" stroke="${color}" stroke-width="${width}" stroke-linecap="round"/>`;
const portrait = (x, y, color, r = 36) =>
  circle(x, y, r, color) +
  circle(x, y - 8, r * 0.28, C.cream) +
  `<path d="M${x - r * 0.55} ${y + r * 0.5}Q${x} ${y - r * 0.2} ${x + r * 0.55} ${y + r * 0.5}" fill="${C.cream}"/>`;
const designs = [
  {
    slug: "ai-email-marketing",
    bg: C.cream,
    category: "AI & WRITING",
    alt: "Cream Xem cover with a purple email draft, a human review card, and the words Less noise. More human.",
    body:
      title(76, 254, ["Less noise.", "More human."], 100) +
      text(80, 466, "A thoughtful way to write with AI.", 23) +
      pill(80, 526, "AI + YOUR VOICE", C.lavender, C.iris, 192) +
      circle(931, 397, 228, C.lavender) +
      group(
        713,
        174,
        -9,
        mail(0, 0, 290, 387, C.iris, "Something worth opening."),
      ) +
      group(
        825,
        449,
        8,
        rect(0, 0, 265, 147, C.lemon, 14, C.ink) +
          label(20, 35, "HUMAN REVIEW") +
          text(20, 81, "Sounds like us.", 34, C.ink, true) +
          pathStroke("M205 92l14 14 25-32", C.forest, 5),
      ) +
      star(702, 588, 43) +
      star(1108, 215, 25, C.forest),
  },
  {
    slug: "ai-email-prompts",
    bg: C.iris,
    light: true,
    category: "AI & WRITING",
    alt: "Iris purple cover with a large quotation mark and cream prompt cards labeled context, voice, and intent.",
    body:
      text(72, 350, "“", 350, C.lemon, true) +
      title(295, 234, ["Better briefs.", "Better emails."], 94, C.cream) +
      group(
        103,
        454,
        -5,
        rect(0, 0, 302, 174, C.lavender, 10) +
          label(24, 39, "01 / CONTEXT") +
          text(24, 90, "Who is reading?", 32, C.ink, true) +
          text(24, 131, "Start with the person.", 18),
      ) +
      group(
        448,
        438,
        3,
        rect(0, 0, 302, 174, C.cream, 10) +
          label(24, 39, "02 / VOICE") +
          text(24, 90, "Make it sound like you.", 28, C.ink, true) +
          text(24, 131, "Specific beats generic.", 18),
      ) +
      group(
        803,
        459,
        -4,
        rect(0, 0, 302, 174, C.lemon, 10) +
          label(24, 39, "03 / INTENT") +
          text(24, 90, "One useful next step.", 30, C.ink, true) +
          text(24, 131, "Give the draft a purpose.", 18),
      ) +
      star(1090, 166, 32, C.lavender),
  },
  {
    slug: "email-deliverability",
    bg: C.forest,
    light: true,
    category: "DELIVERABILITY",
    alt: "Forest green cover showing a cream envelope inside concentric delivery rings, with SPF, DKIM, and DMARC labels.",
    body:
      [110, 165, 220]
        .map((r) => circle(600, 314, r, "none", "#ffffef30"))
        .join("") +
      envelope(478, 236, 244, C.cream, C.lemon) +
      circle(725, 384, 31, C.lemon) +
      pathStroke("M710 384l10 10 20-24", C.forest, 4) +
      pill(217, 210, "SPF", C.lemon, C.forest, 100) +
      pill(897, 274, "DKIM", C.lavender, C.forest, 114) +
      pill(224, 405, "DMARC", C.cream, C.forest, 132) +
      star(955, 453, 30, C.lemon) +
      title(600, 621, ["Earn your place in the inbox."], 72, C.cream).replace(
        "<text ",
        '<text text-anchor="middle" ',
      ) +
      text(
        600,
        670,
        "Good sending starts long before you press send.",
        22,
        C.cream,
        false,
        'text-anchor="middle"',
      ),
  },
  {
    slug: "newsletter-growth",
    bg: C.lemon,
    category: "GROWTH",
    alt: "Lemon yellow cover with envelopes growing like leaves on a forest green stem beside the words Grow a readership.",
    body:
      pathStroke("M340 657C340 494 400 357 352 182", C.forest, 8) +
      pathStroke(
        "M356 550Q249 494 213 401M362 432Q486 350 518 272M367 323Q280 292 251 220",
        C.forest,
        6,
      ) +
      group(141, 352, -20, envelope(0, 0, 178, C.cream, C.forest)) +
      group(425, 209, 17, envelope(0, 0, 174, C.lavender, C.forest)) +
      group(184, 153, -14, envelope(0, 0, 143, C.iris, C.cream)) +
      circle(341, 655, 13, C.forest) +
      title(651, 295, ["Grow a", "readership."], 99, C.forest) +
      text(658, 502, "Not just a list.", 39, C.forest, true) +
      text(658, 553, "Permission. Consistency. Something good.", 18, C.forest) +
      star(1033, 156, 40, C.iris),
  },
  {
    slug: "email-design",
    bg: C.lavender,
    category: "DESIGN",
    alt: "Lavender editorial cover displaying three different email layouts with iris, green, and lemon blocks under Design for the inbox.",
    body:
      title(75, 208, ["Design for the inbox."], 91) +
      text(79, 256, "A little structure. A lot of personality.", 22) +
      group(103, 331, -6, mail(0, 0, 259, 307, C.iris, "The Sunday edit.")) +
      group(
        470,
        325,
        3,
        mail(0, 0, 254, 325, C.forest, "A fresh perspective."),
      ) +
      group(
        827,
        332,
        -4,
        mail(0, 0, 255, 298, C.lemon, "Meet your next idea."),
      ) +
      star(1120, 256, 32, C.iris),
  },
  {
    slug: "email-analytics",
    bg: C.cream,
    category: "ANALYTICS",
    alt: "Cream and forest green Xem cover with iris and lavender stacked bars, a rising line, and the headline Read the signals.",
    body:
      title(78, 204, ["Read the signals."], 98) +
      text(82, 254, "Find the story behind the send.", 23) +
      rect(78, 319, 682, 349, C.forest, 16) +
      label(105, 358, "ACTIVITY OVER TIME", C.cream) +
      [0, 1, 2, 3]
        .map((i) => line(111, 404 + i * 66, 724, 404 + i * 66, "#ffffef25"))
        .join("") +
      [105, 166, 146, 220, 193, 263, 242]
        .map(
          (h, i) =>
            rect(128 + i * 85, 631 - h, 42, h, C.lavender, 4) +
            rect(128 + i * 85, 631 - h, 42, h * 0.39, C.iris, 4),
        )
        .join("") +
      pathStroke(
        "M135 524L222 510L307 483L392 501L477 433L562 419L687 386",
        C.lemon,
        4,
      ) +
      circle(687, 386, 7, C.lemon) +
      circle(936, 433, 104, "none", C.lavender) +
      `<circle cx="936" cy="433" r="104" fill="none" stroke="${C.iris}" stroke-width="35" stroke-dasharray="420 654" transform="rotate(-90 936 433)"/>` +
      text(936, 446, "Clarity.", 41, C.ink, true, 'text-anchor="middle"') +
      label(851, 610, "BEYOND THE OPEN") +
      star(1075, 163, 36, C.iris),
  },
  {
    slug: "email-automation",
    bg: C.forest,
    light: true,
    category: "AUTOMATION",
    alt: "Forest green cover with a branching email journey from signup through welcome and follow-up, titled Good timing. By design.",
    body:
      title(76, 211, ["Good timing. By design."], 86, C.cream) +
      text(80, 264, "Build journeys that feel personal.", 23, C.cream) +
      pathStroke("M307 469H417V374H751M417 469V585H751", C.lemon, 3) +
      circle(417, 469, 8, C.lemon) +
      rect(80, 407, 227, 125, C.lemon, 16) +
      label(102, 439, "THE BEGINNING") +
      text(102, 491, "New subscriber", 28, C.forest, true) +
      rect(512, 328, 158, 87, C.cream, 44) +
      text(
        591,
        380,
        "Wait a little",
        18,
        C.forest,
        false,
        'text-anchor="middle"',
      ) +
      rect(512, 542, 158, 87, C.lavender, 44) +
      text(
        591,
        594,
        "If they click",
        18,
        C.forest,
        false,
        'text-anchor="middle"',
      ) +
      rect(751, 321, 350, 135, C.cream, 14) +
      envelope(775, 354, 77, C.lavender, C.forest) +
      text(877, 379, "A warm welcome.", 29, C.forest, true) +
      text(878, 410, "Make the first hello count.", 15, C.forest) +
      rect(751, 520, 350, 135, C.lavender, 14) +
      envelope(775, 553, 77, C.cream, C.forest) +
      text(877, 578, "The right next step.", 27, C.forest, true) +
      text(878, 609, "Continue the conversation.", 15, C.forest),
  },
  {
    slug: "email-personalization",
    bg: C.lavender,
    category: "GROWTH",
    alt: "Lavender cover with three individual reader portraits linked to a personalized email, titled More than a first name.",
    body:
      title(75, 219, ["More than", "a first name."], 96) +
      text(81, 446, "Relevance is the real personalization.", 21) +
      pill(81, 499, "PEOPLE, NOT FIELDS", C.cream, C.iris, 219) +
      circle(926, 379, 211, "none", "#5b3cc430") +
      circle(926, 379, 154, "none", "#5b3cc430") +
      pathStroke(
        "M775 210L926 380M1094 248L926 380M1068 584L926 380",
        C.iris,
        2,
      ) +
      portrait(775, 210, C.forest, 53) +
      portrait(1094, 248, C.iris, 47) +
      portrait(1068, 584, C.forest, 52) +
      group(790, 324, -7, envelope(0, 0, 253, C.cream, C.iris)) +
      pill(785, 505, "Picked for you", C.lemon, C.forest, 167) +
      star(640, 591, 34, C.iris),
  },
  {
    slug: "email-ab-testing",
    bg: C.cream,
    category: "ANALYTICS",
    alt: "Split cream and lavender cover showing two different email variants marked A and B under One question. Two possibilities.",
    body:
      rect(600, 99, 600, 620, C.lavender) +
      title(78, 203, ["One question. Two possibilities."], 76) +
      text(122, 451, "A", 171, C.iris, true) +
      text(1010, 451, "B", 171, C.iris, true) +
      group(
        288,
        316,
        -5,
        mail(0, 0, 230, 289, C.forest, "A thoughtful hello."),
      ) +
      group(682, 305, 5, mail(0, 0, 230, 289, C.iris, "Your next good read.")) +
      circle(600, 474, 40, C.iris) +
      text(600, 481, "vs", 24, C.cream, true, 'text-anchor="middle"') +
      text(79, 673, "Change one thing. Learn something useful.", 24) +
      star(1090, 631, 29, C.forest),
  },
  {
    slug: "welcome-email-sequence",
    bg: C.iris,
    light: true,
    category: "AUTOMATION",
    alt: "Purple cover with three rising cream and lavender email cards representing hello, value, and the next step.",
    body:
      title(74, 206, ["Start something good."], 91, C.cream) +
      text(80, 259, "The first hello is just the beginning.", 23, C.cream) +
      pathStroke("M139 650H415V565H728V472H1068", C.lemon, 3) +
      group(
        104,
        417,
        -5,
        mail(0, 0, 262, 236, C.lavender, "01 / A proper hello."),
      ) +
      group(
        469,
        358,
        3,
        mail(0, 0, 262, 236, C.lemon, "02 / Something useful."),
      ) +
      group(
        836,
        295,
        -4,
        mail(0, 0, 262, 236, C.forest, "03 / A little momentum."),
      ) +
      star(1080, 624, 41, C.lemon),
  },
];
const output = path.resolve("public/images/blog");
await fs.mkdir("design/blog/svg", { recursive: true });
const browser = await chromium.launch({ channel: "chrome", headless: true });
try {
  const page = await browser.newPage({
    viewport: { width: 1600, height: 1067 },
    deviceScaleFactor: 1,
  });
  for (let i = 0; i < designs.length; i++) {
    const d = designs[i],
      fg = d.light ? C.cream : C.ink;
    const mark = await fs.readFile("public/brand/xem-mark.png");
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1600" height="1067" viewBox="0 0 1200 800" role="img" aria-labelledby="title"><title id="title">${esc(d.alt)}</title>${rect(0, 0, 1200, 800, d.bg)}<image href="data:image/png;base64,${mark.toString("base64")}" x="76" y="40" width="29" height="29"/>${text(115, 65, "xem", 29, fg, false, 'font-weight="600" letter-spacing="-1"')}${label(207, 62, "THE JOURNAL", fg)}${text(1121, 62, String(i + 1).padStart(2, "0"), 18, fg, false, 'text-anchor="end"')}${line(77, 94, 1123, 94, fg + "44")}${d.body}${line(77, 722, 1123, 722, fg + "44")}${label(78, 759, d.category, fg)}${text(1122, 759, "Every email, a little more human.", 16, fg, false, 'text-anchor="end"')}</svg>`;
    await fs.writeFile(`design/blog/svg/${d.slug}.svg`, svg);
    await page.setContent(svg);
    await page.evaluate(
      async ({ sans, editorial }) => {
        for (const [name, data] of [
          ["DM Sans", sans],
          ["EB Garamond", editorial],
        ]) {
          const f = new FontFace(name, `url(data:font/woff2;base64,${data})`, {
            weight: "100 1000",
          });
          document.fonts.add(await f.load());
        }
        document.body.style.margin = "0";
        await document.fonts.ready;
      },
      {
        sans: (await fs.readFile("design/blog/fonts/dm-sans.woff2")).toString(
          "base64",
        ),
        editorial: (
          await fs.readFile("design/blog/fonts/eb-garamond.woff2")
        ).toString("base64"),
      },
    );
    const png = await page.locator("svg").screenshot();
    await sharp(png)
      .webp({ quality: 91, effort: 6 })
      .toFile(path.join(output, `${d.slug}.webp`));
    const file = `content/blog/${d.slug}.md`;
    const md = (await fs.readFile(file, "utf8"))
      .replace(/^coverAlt:.*$/m, `coverAlt: ${JSON.stringify(d.alt)}`)
      .replace(
        /^coverSource:.*$/m,
        'coverSource: "https://xem.email/blog/editorial"',
      );
    await fs.writeFile(file, md);
    console.log(`Created ${d.slug}`);
  }
  const tiles = await Promise.all(
    designs.map(async (d, i) => ({
      input: await sharp(path.join(output, `${d.slug}.webp`))
        .resize(480, 320)
        .toBuffer(),
      left: (i % 2) * 480,
      top: Math.floor(i / 2) * 320,
    })),
  );
  await sharp({
    create: { width: 960, height: 1600, channels: 3, background: C.cream },
  })
    .composite(tiles)
    .png()
    .toFile("design/blog/contact-sheet.png");
} finally {
  await browser.close();
}
