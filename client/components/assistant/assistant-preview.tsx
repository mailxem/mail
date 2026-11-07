"use client";
import { useMemo, useState } from "react";
import { useChat } from "@ai-sdk/react";
import { createChat } from "@shadcn/helpers/ai-sdk";
import type { UIMessage } from "ai";
import { ChatView } from "./chat-view";
import type { ActionOutput } from "@/lib/assistant/types";

export function AssistantPreview() {
  const [input, setInput] = useState("");
  const [iteration, setIteration] = useState(0);
  const script = useMemo(
    () =>
      createChat()
        .user("Help me plan our next newsletter.")
        .assistant(({ writer }) => {
          writer
            .tool("get_marketing_options", { input: {} })
            .sleep(500)
            .output({
              audience: "The Xem community",
              subscribers: 2486,
              template: "The weekly edit",
            });
          writer.text(
            "Let’s make it worth opening. Your **Xem community** (2,486 readers) is a good place to start: one useful idea, one clear link, your voice.",
          );
          writer
            .tool("ask_user", { input: {} })
            .output({
              kind: "xem-ask",
              question: "What should this edition be about?",
              options: ["Our new managed sending", "Last month’s best tips", "Something else"],
            });
        })
        .user(
          "Email the people who created an API key but never sent. Draft it, don't send.",
        )
        .assistant(({ writer }) => {
          writer
            .tool("get_contact_lists", { input: {} })
            .sleep(300)
            .output({ data: [{ name: "All users" }, { name: "Beta" }], total: 2 });
          writer
            .tool("get_marketing_options", { input: {} })
            .sleep(300)
            .output({ senders: [{ fromEmail: "human@updates.xem.email" }], templates: [{ name: "Product update" }] });
          writer.text(
            "Seven people made a key and never sent. Here's who:\n\n| Name | Email | Key created |\n| --- | --- | --- |\n| Shubhojeet Bera | shubhobera98@gmail.com | 2 Nov |\n| Babbz V | sagarvagoni45@gmail.com | 1 Nov |\n| Five more | | Oct 28 to Nov 4 |\n\nI'll put them in their own list, then draft the email from **human@updates.xem.email** on your *Product update* template.",
          );
          writer
            .tool("add_contacts", { input: {} })
            .sleep(300)
            .output({
              kind: "xem-action",
              actionId: "preview-contacts",
              tool: "add_contacts",
              status: "pending",
              input: {
                listId: "8bb3d38d-6073-4a5a-b7e9-8f0b2429b1df",
                contacts: [
                  { email: "sagarvagoni45@gmail.com", firstName: "Babbz" },
                  { email: "shubhobera98@gmail.com", firstName: "Shubhojeet" },
                  { email: "ana@acme.dev" },
                  { email: "li@tiny.app" },
                  { email: "omar@ship.io" },
                  { email: "kit@north.co" },
                  { email: "eve@shop.in" },
                ],
                why: "So the email goes only to people who haven't sent yet.",
              },
            });
          writer
            .tool("create_campaign", { input: {} })
            .sleep(300)
            .output({
              kind: "xem-action",
              actionId: "preview-draft",
              tool: "create_campaign",
              status: "pending",
              input: {
                name: "Your first email is one step away",
                subject: "You made a key. Let's send something.",
                smtpConfigId: "01d17244-f026-45c4-af20-2a925056d576",
                why: "A draft you can read before anything goes out.",
              },
            });
          writer.text("Approve both and I'll check the draft and tell you who it will reach. Nothing is sent.");
        })
        .user("Continue")
        .assistant(({ writer }) => {
          writer
            .tool("get_campaign", { input: {} })
            .sleep(400)
            .output({ name: "Your first email is one step away", recipients: 7 });
          writer.text(
            "Done. **7 contacts** are in the list and the draft **Your first email is one step away** is saved for those 7 people. Nothing has been sent.",
          );
          writer
            .tool("ask_user", { input: {} })
            .output({
              kind: "xem-ask",
              question: "Want to send it?",
              fields: [
                { name: "when", label: "Send at", type: "date" },
                { name: "test", label: "Send a test to", type: "email", placeholder: "you@company.com" },
              ],
              submitLabel: "Schedule it",
              options: ["Open the draft", "Add a reminder in 3 days"],
            });
        }),
    [iteration],
  );
  const transport = useMemo(() => script.transport({ delayMs: 28 }), [script]);
  const chat = useChat({ id: `preview-${iteration}`, transport });
  return (
    <ChatView
      demo
      messages={chat.messages}
      input={input}
      setInput={setInput}
      busy={chat.status === "streaming" || chat.status === "submitted"}
      onSend={() => {
        const next = script.next(chat.messages);
        if (next) {
          setInput("");
          void chat.sendMessage(next);
        }
      }}
      onStop={() => void chat.stop()}
      onNew={() => {
        setIteration((n) => n + 1);
        chat.setMessages([]);
        setInput("");
      }}
      onAction={(id, approved) => {
        const messages = chat.messages.map((message) => ({
          ...message,
          parts: message.parts.map((part) =>
            "output" in part && (part.output as ActionOutput)?.actionId === id
              ? {
                  ...part,
                  output: {
                    ...(part.output as ActionOutput),
                    status: approved ? "completed" : "declined",
                    result: approved ? { notice: "Preview only. Nothing was changed." } : undefined,
                  },
                }
              : part,
          ),
        })) as UIMessage[];
        chat.setMessages(messages);
        // Like the real assistant: once nothing waits for review, it carries on.
        const waiting = messages.some((m) =>
          m.parts.some((p) => "output" in p && (p.output as ActionOutput)?.status === "pending"),
        );
        const next = script.next(messages);
        if (!waiting && next) setTimeout(() => void chat.sendMessage(next), 300);
      }}
    />
  );
}
