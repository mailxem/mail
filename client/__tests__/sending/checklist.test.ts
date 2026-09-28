import { state } from "@/test/fixtures/sending";
import { checklist, onboardingNavigation } from "@/lib/sending/types";

test("earlier BYO mail and another domain's test do not complete managed setup", () => {
  const s = state();
  s.journey.testedDomainIds = ["domain-b"];
  expect(checklist(s)[3].done).toBe(false);
  s.journey.testedDomainIds.push("domain-a");
  expect(checklist(s)[3].done).toBe(true);
  expect(checklist(s).every((step) => step.done)).toBe(false);
  s.journey.campaigns = 1;
  expect(checklist(s).every((step) => step.done)).toBe(true);
});
test("a fresh workspace offers a neutral choice; disabled managed sending preserves BYO progress", () => {
  const s = state();
  s.account.sendingMode = "";
  expect(checklist(s)).toHaveLength(1);
  expect(checklist(s)[0].done).toBe(false);
  s.enabled = false;
  s.account.sendingMode = "BYO";
  expect(checklist(s)).toHaveLength(4);
  expect(checklist(s)[1].done).toBe(true);
});

test("managed onboarding stays on DNS until every required check passes", () => {
  const s = state();
  s.domains[0].ready = false;
  for (const status of ["OWNERSHIP_REQUIRED", "AWAITING_APPROVAL", "PENDING"]) {
    s.domains[0].identityStatus = status;
    const steps = checklist(s);
    expect(onboardingNavigation(steps, null).active).toBe(2);
    expect(onboardingNavigation(steps, 3).active).toBe(2);
    expect(onboardingNavigation(steps, 4).unlocked).toBe(2);
    expect(onboardingNavigation(steps, 1).active).toBe(1);
  }
  s.domains[0].ready = true;
  expect(onboardingNavigation(checklist(s), 3).active).toBe(3);
  expect(onboardingNavigation(checklist(s), 2).active).toBe(2);
  // A queued message is not provider acceptance.
  expect(onboardingNavigation(checklist(s), 4).active).toBe(3);
  s.journey.testedDomainIds = [s.domains[0].id];
  expect(onboardingNavigation(checklist(s), 4).active).toBe(4);
  // DNS invalidation must bring even a previously selected later step back.
  s.domains[0].ready = false;
  expect(onboardingNavigation(checklist(s), 4).active).toBe(2);
});

test("BYO setup requires a provider, then a draft, before the first send", () => {
  const s = state();
  s.account.sendingMode = "BYO";
  s.journey = {
    byoSenders: 0,
    campaigns: 0,
    acceptedEmails: 0,
    testedDomainIds: [],
  };
  expect(onboardingNavigation(checklist(s), 3).active).toBe(1);
  s.journey.byoSenders = 1;
  expect(onboardingNavigation(checklist(s), 3).active).toBe(2);
  s.journey.campaigns = 1;
  expect(onboardingNavigation(checklist(s), 3).active).toBe(3);
  s.journey.acceptedEmails = 1;
  expect(checklist(s).every((step) => step.done)).toBe(true);
});
