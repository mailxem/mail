// Xem Cloud plans, shared by /pricing, /compare and the cost calculator.
export const xemPlans = [
  { name: "Free", price: 0, included: 3000, overage: null },
  { name: "Starter", price: 9, included: 20000, overage: 0.5 },
  { name: "Growth", price: 29, included: 100000, overage: 0.3 },
  { name: "Scale", price: 99, included: 500000, overage: 0.2 },
] as const;

/** Cheapest Xem Cloud plan for a month's emails, with its cost in USD. */
export function xemMonthlyCost(emails: number) {
  return xemPlans
    .filter((p) => p.overage !== null || emails <= p.included)
    .map((p) => ({
      plan: p.name,
      cost:
        p.price +
        (p.overage ? (Math.max(0, emails - p.included) / 1000) * p.overage : 0),
    }))
    .sort((a, b) => a.cost - b.cost)[0];
}

/** Emails a month for a weekly newsletter: four sends to the whole list. */
export const weeklyEmails = (subscribers: number) => subscribers * 4;

/** Self-hosting through Amazon SES at $0.10 per 1,000 emails, before server costs. */
export const sesMonthlyCost = (emails: number) => (emails / 1000) * 0.1;
