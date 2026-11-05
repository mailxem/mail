# Email design skill

How Xem's designer adapts a finished template to a new brief. The design itself
(layout, fonts, colours, illustrations) comes from the reference; this guide is
about making the words fit that design the way a good editorial designer would.

## Read the reference first

Every Studio template has a rhythm. Find it before writing:

- **Masthead**: brand name, sometimes a date or edition line. Keep it short and
  factual. Never invent a date: use `[date]` or keep the original pattern.
- **Kicker**: a small label above a headline ("COVER STORY", "WHAT'S NEW",
  "IN THIS ISSUE"). Two to four words, all caps if the original is.
- **Headline**: the one idea of the email. Three to nine words. Concrete nouns and
  verbs, no puns that need explaining, no colons unless the original uses one.
- **Dek**: one or two sentences under the headline that say what's inside and why
  it matters to this reader. Up to about 25 words.
- **Hero**: an illustration or photo. Its alt text should describe what's shown.
- **Sections**: each section has one job (a story, a feature, a step, a stat, an
  event). Title of three to eight words, one or two sentences of body, then one
  action.
- **Actions**: a filled button is the main action (one to four words, verb first:
  "Read the essay", "Claim your key"). Secondary actions are text links ending in
  " →" ("Read more →", "See the schedule →").
- **Sign-off**: a first name and role when the original has one; use `[Name]` and
  `[Role]` placeholders rather than inventing a person.
- **Footer**: company · `[address]` · unsubscribe. Keep it plain and short.

## Match the email's purpose

- **Newsletter or digest**: lead story first, then two to five shorter items.
  Each item gets a specific headline and one line on why to read it. An
  "In this issue" list mirrors the section titles.
- **Release notes / product update**: one headline naming the biggest change.
  Each feature card: what it is, what the reader can now do, one link. Prefer
  "You can now…" over adjectives.
- **Welcome**: thank them in one line, set expectations (what they'll get and how
  often), then one first step. No feature dump.
- **Event or webinar**: what, when, where, who it's for, one action. Use `[date]`,
  `[time]`, `[location]` placeholders when the brief doesn't give them.
- **Offer or commerce**: the product or offer, the one reason it matters, the
  terms placeholder (`[price]`, `[ends date]`), one action. Never invent discounts.
- **Transactional** (receipt, key, reset): the fact first, the action second,
  reassurance third. No marketing language.
- **Announcement**: the news in the headline, what changes for the reader in the
  dek, details in sections, one action.

## Writing rules

- Keep each slot within about half to one and a half times its original length,
  so the layout holds. Headlines and buttons especially.
- Write for one reader. "You" more than "we".
- Specific beats general: "One 3,000-word essay every Sunday" beats "in-depth
  content regularly".
- Never invent facts: prices, dates, numbers, names, customers, quotes, results.
  Use a short placeholder in brackets instead.
- Keep merge variables such as `{{first_name}}` exactly as written.
- Subject line: under about 60 characters, the email's one idea, no clickbait.
  Preheader: the next most useful sentence, not a repeat of the subject.

## Removing sections

Drop a row only when its whole purpose doesn't apply: a product grid in a plain
announcement, a webinar block in a receipt, a stats panel when the brief has no
numbers. Keep the masthead, the main headline section, at least one action and
the footer.
