import type { FormDefinition, FormQuestion } from "./form-definition";
import type { LeadForm } from "./types";

const email: FormQuestion = {
  key: "email",
  label: "Email address",
  type: "EMAIL",
  required: true,
};
const name: FormQuestion = {
  key: "first_name",
  label: "First name",
  type: "TEXT",
  required: false,
};
const consentLabel =
  "I agree to receive emails and understand I can unsubscribe at any time.";

export type FormStarter = {
  key: string;
  name: string;
  description: string;
  button: string;
  success: string;
  nextStep: string;
  definition: FormDefinition;
};

export const formStarters: FormStarter[] = [
  {
    key: "newsletter",
    name: "Newsletter",
    description: "A simple invitation to stay in touch.",
    button: "Subscribe",
    success: "You’re on the list. Thanks for joining us!",
    nextStep: "Connect a form-completion journey to send your welcome email.",
    definition: {
      schemaVersion: 1,
      pages: [{ id: "signup", title: "Join the list", fields: [email, name] }],
      consent: { mode: "required", label: consentLabel },
      saveProgress: false,
    },
  },
  {
    key: "lead-magnet",
    name: "Resource request",
    description: "Deliver the resource that matches each reader’s interests.",
    button: "Request the guide",
    success: "Thanks! Your resource request has been received.",
    nextStep:
      "Create a form-completion journey with your resource link for visitors who opt into emails.",
    definition: {
      schemaVersion: 1,
      pages: [
        {
          id: "resource",
          title: "What would help you most?",
          fields: [
            {
              key: "interest",
              label: "Choose a topic",
              type: "RADIO",
              required: true,
              options: [
                "Getting started",
                "Growing my audience",
                "Email automation",
              ],
            },
          ],
        },
        {
          id: "delivery",
          title: "Where should we send it?",
          fields: [email, name],
        },
      ],
      consent: { mode: "optional", label: consentLabel },
      saveProgress: false,
    },
  },
  {
    key: "waitlist",
    name: "Waitlist",
    description: "Learn what early customers want to build.",
    button: "Join the waitlist",
    success: "You’re on the waitlist. Thank you for your interest!",
    nextStep: "Use the interest answer to branch your welcome journey.",
    definition: {
      schemaVersion: 1,
      pages: [
        { id: "signup", title: "Be the first to know", fields: [email, name] },
        {
          id: "interests",
          title: "Make it yours",
          fields: [
            {
              key: "interest",
              label: "What are you interested in?",
              type: "SELECT",
              required: true,
              options: [
                "Newsletters",
                "Transactional email",
                "Marketing automation",
              ],
            },
          ],
        },
      ],
      consent: { mode: "required", label: consentLabel },
      saveProgress: true,
    },
  },
  {
    key: "demo",
    name: "Demo request",
    description: "Ask the right follow-up questions before a conversation.",
    button: "Request a demo",
    success: "Thanks! Your demo request has been received.",
    nextStep:
      "Connect this form to a team notification and a request confirmation.",
    definition: {
      schemaVersion: 1,
      pages: [
        {
          id: "contact",
          title: "Let’s meet",
          fields: [
            email,
            name,
            { key: "company", label: "Company", type: "TEXT", required: true },
          ],
        },
        {
          id: "needs",
          title: "Tell us a little more",
          fields: [
            {
              key: "team_size",
              label: "Team size",
              type: "RADIO",
              required: true,
              options: ["Just me", "2–10", "11–50", "51+"],
            },
            {
              key: "requirements",
              label: "Any requirements we should know about?",
              type: "TEXTAREA",
              required: false,
              condition: {
                field: "team_size",
                operator: "not_equals",
                value: "Just me",
              },
            },
          ],
        },
      ],
      consent: { mode: "optional", label: consentLabel },
      saveProgress: true,
    },
  },
  {
    key: "feedback",
    name: "Feedback",
    description: "Make room for useful, specific customer feedback.",
    button: "Send feedback",
    success: "Thank you. Your feedback has been received.",
    nextStep:
      "Review completed feedback in Submissions. This form does not subscribe people to marketing.",
    definition: {
      schemaVersion: 1,
      pages: [
        {
          id: "feedback",
          title: "How can we improve?",
          fields: [
            email,
            {
              key: "rating",
              label: "How was your experience?",
              type: "RADIO",
              required: true,
              options: ["Great", "Good", "Could be better"],
            },
            {
              key: "message",
              label: "Tell us more",
              type: "TEXTAREA",
              required: true,
            },
          ],
        },
      ],
      consent: { mode: "none", label: "" },
      saveProgress: false,
    },
  },
  {
    key: "preferences",
    name: "Preferences",
    description: "Understand which updates your readers want.",
    button: "Send preferences",
    success: "Thanks for sharing your preferences.",
    nextStep:
      "Use these answers in a form-completion journey. Sending this form alone does not change existing subscriptions.",
    definition: {
      schemaVersion: 1,
      pages: [
        {
          id: "preferences",
          title: "What would you like to hear about?",
          fields: [
            email,
            {
              key: "interest",
              label: "Favorite topic",
              type: "SELECT",
              required: true,
              options: [
                "Product updates",
                "Tips and guides",
                "Community stories",
              ],
            },
            {
              key: "frequency",
              label: "Preferred frequency",
              type: "RADIO",
              required: true,
              options: ["Weekly", "Monthly"],
            },
          ],
        },
      ],
      consent: { mode: "none", label: "" },
      saveProgress: false,
    },
  },
];

/** Each editor gets its own copy; starter edits must never affect another form. */
export function copyFormDefinition(definition: FormDefinition): FormDefinition {
  return JSON.parse(JSON.stringify(definition)) as FormDefinition;
}

/** Keep the full persisted shape when duplicating or changing publication state. */
export function formSavePayload(form: LeadForm) {
  return {
    name: form.Name,
    description: form.description,
    listId: form.AddToListID,
    status: form.Status,
    successMessage: form.successMessage,
    buttonText: form.SubmitButtonText,
    theme: form.theme,
    ...(form.definition ? { definition: form.definition } : {}),
    fields: form.fields.map((field) => ({
      label: field.Label,
      type: field.FieldType,
      required: field.Required,
      key: field.mapToContactField,
    })),
  };
}

/** Renaming an identifier preserves its dependent conditions. */
export function renameFormField(
  definition: FormDefinition,
  oldKey: string,
  newKey: string,
  target: { pageIndex: number; fieldIndex: number },
): FormDefinition {
  const unique =
    definition.pages
      .flatMap((page) => page.fields)
      .filter((field) => field.key === oldKey).length === 1;
  return {
    ...definition,
    pages: definition.pages.map((page, pageIndex) => ({
      ...page,
      condition:
        unique && page.condition?.field === oldKey
          ? { ...page.condition, field: newKey }
          : page.condition,
      fields: page.fields.map((field, fieldIndex) => ({
        ...field,
        key:
          target.pageIndex === pageIndex && target.fieldIndex === fieldIndex
            ? newKey
            : field.key,
        condition:
          unique && field.condition?.field === oldKey
            ? { ...field.condition, field: newKey }
            : field.condition,
      })),
    })),
  };
}

export function fieldIsReferenced(
  definition: FormDefinition,
  key: string,
): boolean {
  return definition.pages.some(
    (page) =>
      page.condition?.field === key ||
      page.fields.some((field) => field.condition?.field === key),
  );
}
