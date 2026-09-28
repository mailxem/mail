import type { SendingState } from "@/lib/sending/types";

export function state(): SendingState {
  return {
    enabled: true,
    account: {
      sendingMode: "MANAGED",
      approved: true,
      suspended: false,
      paused: false,
      onboardingDismissed: false,
      dailyLimit: 200,
      monthlyLimit: 1000,
      monthlyBudgetMicros: 1000000,
      dailyUsed: 0,
      monthlyUsed: 0,
      budgetUsedMicros: 0,
    },
    domains: [
      {
        id: "domain-a",
        name: "example.com",
        ready: true,
        smtpConfigId: "sender",
        fromEmail: "hello@example.com",
        token: "",
        ownership: true,
        provisioned: true,
        identityStatus: "SUCCESS",
        dkimStatus: "SUCCESS",
        mailFromStatus: "SUCCESS",
        dmarcStatus: "VALID",
        checkedAt: null,
        records: [],
      },
    ],
    credentials: [],
    messages: [],
    smtp: {
      enabled: true,
      host: "smtp.example.com",
      port: 587,
      security: "STARTTLS",
    },
    costPerRecipientMicros: 1000,
    maxMessageBytes: 5242880,
    journey: {
      byoSenders: 1,
      campaigns: 0,
      acceptedEmails: 20,
      testedDomainIds: [],
    },
  };
}
