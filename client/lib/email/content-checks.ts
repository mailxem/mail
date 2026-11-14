// Things spam filters and readers dislike, checked in the browser before
// sending. Deterministic and free: no AI. Each says what to change.
// No unsubscribe check: sending adds an unsubscribe footer to every campaign.

export type ContentWarning = { id: string; title: string; fix: string };

const SHORTENERS =
  /^(?:www\.)?(?:bit\.ly|tinyurl\.com|t\.co|goo\.gl|ow\.ly|is\.gd|buff\.ly|rebrand\.ly|cutt\.ly|shorturl\.at|rb\.gy|tiny\.cc|bl\.ink|s\.id|t\.ly|v\.gd|lnkd\.in)$/i;

export const LONG_SUBJECT = 70;

/** Visible words in the email body. */
export function visibleText(html: string) {
  return html
    .replace(/<!--[\s\S]*?-->/g, " ")
    .replace(/<(head|style|script|title)\b[\s\S]*?<\/\1>/gi, " ")
    .replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;|&#160;|&zwnj;|&#8204;/gi, " ")
    .replace(/&[a-z#0-9]+;/gi, "")
    .replace(/\s+/g, " ")
    .trim();
}

export function contentChecks({
  subject = "",
  html = "",
}: {
  subject?: string;
  html?: string;
}): ContentWarning[] {
  const warnings: ContentWarning[] = [];
  // Fields like {{first_name}} are filled in per reader; leave them out.
  const plainSubject = subject.replace(/{{[^{}]*}}/g, "").trim();
  const text = visibleText(html);

  const letters = plainSubject.replace(/[^\p{L}]/gu, "");
  const capitals = letters.replace(/[^\p{Lu}]/gu, "");
  if (letters.length >= 8 && capitals.length / letters.length > 0.6)
    warnings.push({
      id: "caps-subject",
      title: "The subject is mostly in capitals",
      fix: "Write it in sentence case. All capitals reads as shouting and spam filters count it against you.",
    });

  if (/!{3,}/.test(subject) || /!{3,}/.test(text))
    warnings.push({
      id: "exclamations",
      title: /!{3,}/.test(subject)
        ? "The subject has three or more exclamation marks in a row"
        : "The email has three or more exclamation marks in a row",
      fix: "Use one exclamation mark at most. Strings of them look like spam.",
    });

  const shortened = new Set<string>();
  for (const [, url] of html.matchAll(
    /\bhttps?:\/\/([^\s"'<>/?#]+)[^\s"'<>]*/gi,
  ))
    if (SHORTENERS.test(url)) shortened.add(url.toLowerCase());
  if (shortened.size)
    warnings.push({
      id: "shorteners",
      title: `Shortened links (${[...shortened].join(", ")})`,
      fix: "Link to the full address instead. Spammers use shorteners to hide where links go, so filters distrust them.",
    });

  const images = html.match(/<img\b[^>]*>/gi) ?? [];
  const words = text ? text.split(" ").length : 0;
  // Readers who block images need text to read, so images raise the bar.
  if (html.trim() && words < (images.length ? 50 : 20))
    warnings.push(
      images.length
        ? {
            id: "little-text",
            title:
              words < 5
                ? "The email is only images"
                : "The email is mostly images with little text",
            fix: "Add a few sentences of real text. Many inboxes block images, so those readers see an empty email, and image-only email looks like spam.",
          }
        : {
            id: "little-text",
            title: "The email has very little text",
            fix: "Say a bit more about why you're writing. Very short emails with a link are a common spam pattern.",
          },
    );

  const noAlt = images.filter(
    (img) =>
      !/\balt\s*=\s*("[^"]*\S[^"]*"|'[^']*\S[^']*'|[^\s"'>]+)/i.test(img),
  ).length;
  if (noAlt)
    warnings.push({
      id: "alt-text",
      title:
        noAlt === 1
          ? "1 image has no alt text"
          : `${noAlt} images have no alt text`,
      fix: "Describe each image in its alt text. Readers with images turned off, and screen readers, see that text instead.",
    });

  if (subject.trim().length > LONG_SUBJECT)
    warnings.push({
      id: "long-subject",
      title: `The subject is ${subject.trim().length} characters long`,
      fix: "Keep it under about 60 characters, or put the important words first. Longer subjects get cut off, most of all on phones.",
    });

  return warnings;
}
