import type { IMAPEmail } from "@/types/imap";

/** Development-only sample mail for designing the inbox and outbox (`?fixtures=1`). */
export function mailFixtures(outbox: boolean): (IMAPEmail & { status?: string })[] {
  const day = (offset: number, h: number, m: number) => {
    const d = new Date();
    d.setDate(d.getDate() - offset);
    d.setHours(h, m, 0, 0);
    return d.toISOString();
  };
  const body = (greeting: string, lines: string[], sign: string) =>
    `<div style="font-family:system-ui,sans-serif;font-size:15px;line-height:1.7;color:#3f3f46"><p>${greeting}</p>${lines.map((l) => `<p>${l}</p>`).join("")}<p>Best regards,</p><p>${sign}</p></div>`;
  const mail = [
    {
      from: "Rico Oktananda <rico.oktananda1@gmail.com>",
      subject: "We need your UX feedback",
      text: "We’re refining our product and need your insights on our user experience (UX) design before the next release.",
      at: day(0, 5, 16),
      seen: false,
      sign: "Rico",
    },
    {
      from: "Alicia from Deel <alicia@deel.support>",
      subject: "Your withdrawal is on the way",
      text: "Your money is on the way! Your Deel withdrawal is all set, and your money is on its way to your bank account.",
      at: day(3, 10, 2),
      seen: true,
      sign: "Alicia",
    },
    {
      from: "Substack Read <read@substack.com>",
      subject: "Research as a leisure activity",
      text: "Research as a leisure activity, collecting curiosities, and other reads picked for you this week.",
      at: day(3, 8, 40),
      seen: true,
      sign: "The Substack team",
    },
    {
      from: "Maya Chen <maya@northwind.studio>",
      subject: "Re: Spring campaign assets",
      text: "Attached are the final banner sizes for the spring campaign. Let me know if the hero copy needs another pass.",
      at: day(5, 16, 25),
      seen: true,
      sign: "Maya",
    },
    {
      from: "Stripe <receipts@stripe.com>",
      subject: "Your receipt from Xem #2381-0042",
      text: "Thanks for your payment. Your Growth plan renews on the 12th of next month. View your invoice any time from billing.",
      at: day(8, 9, 12),
      seen: true,
      sign: "Stripe",
    },
  ];
  return mail.map((m, i) => ({
    id: `fixture-${i}`,
    from: outbox ? "Team Xem <team@xem.email>" : m.from,
    to: outbox ? m.from : "Team Xem <team@xem.email>",
    cc: "",
    bcc: "",
    subject: m.subject,
    body: body("Hi Team,", [m.text, "Your feedback is crucial in helping us get this right."], m.sign),
    date: m.at,
    messageId: `<fixture-${i}@xem.test>`,
    flags: m.seen ? ["\\Seen"] : [],
    attachments: [],
    ...(outbox ? { status: ["DELIVERED", "OPENED", "CLICKED", "ACCEPTED", "BOUNCED"][i] } : {}),
  }));
}
