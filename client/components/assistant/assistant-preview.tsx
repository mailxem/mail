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
            "Let’s make it worth opening. Your **Xem community** is a good place to start.\n\nHere’s a simple starting point:\n\n- **One useful idea.** Give readers something they can put to work today.\n- **One clear next step.** A single link makes it easier to act.\n- **Your voice.** Keep it warm, direct, and recognizably you.\n\nWhat’s the one thing you’d like readers to take away from this edition?",
          );
        })
        .user(
          "Share our new managed sending guide. Create a draft, don't send it.",
        )
        .assistant(({ writer }) => {
          writer.text(
            "A short introduction, one useful guide, one clear link. Here’s the draft to review.",
          );
          writer
            .tool("create_campaign", {
              input: {
                name: "A simpler way to send",
                subject: "Your domain. A clearer path.",
              },
            })
            .sleep(400)
            .output({
              kind: "xem-action",
              actionId: "preview-draft",
              tool: "create_campaign",
              status: "pending",
              input: {
                name: "A simpler way to send",
                subject: "Your domain. A clearer path.",
                audience: "The Xem community",
                status: "DRAFT",
              },
            });
          writer.text("Approve to save the draft. Nothing will be sent.");
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
        chat.setMessages(
          (messages) =>
            messages.map((message) => ({
              ...message,
              parts: message.parts.map((part) => {
                if (
                  "output" in part &&
                  (part.output as ActionOutput)?.actionId === id
                )
                  return {
                    ...part,
                    output: {
                      ...(part.output as ActionOutput),
                      status: approved ? "completed" : "declined",
                      result: approved
                        ? {
                            notice:
                              "Preview draft saved. No real campaign was created.",
                          }
                        : undefined,
                    },
                  };
                return part;
              }),
            })) as UIMessage[],
        );
      }}
    />
  );
}
