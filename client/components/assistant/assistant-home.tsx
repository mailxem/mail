"use client";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import { ChatView, messageText } from "./chat-view";
import type { ConversationSummary } from "@/lib/assistant/types";

async function responseJSON(response: Response) {
  const data = await response.json();
  if (!response.ok)
    throw new Error(data.error || "Couldn’t connect. Please try again.");
  return data;
}
export function AssistantHome() {
  const [id, setId] = useState("");
  const current = useRef("");
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(true);
  const [enabled, setEnabled] = useState(true);
  const [canWrite, setCanWrite] = useState(false);
  const [problem, setProblem] = useState("");
  const [acting, setActing] = useState(false);
  const [conversations, setConversations] = useState<ConversationSummary[]>([]);
  const refresh = useCallback(async () => {
    const data = await responseJSON(
      await fetch("/api/assistant", { cache: "no-store" }),
    );
    setEnabled(data.enabled);
    setCanWrite(data.canWrite);
    setConversations(data.conversations);
  }, []);
  const transport = useMemo(
    () =>
      new DefaultChatTransport({
        api: "/api/assistant",
        prepareSendMessagesRequest: ({ messages }) => ({
          body: {
            conversationId: current.current,
            text: messageText(messages[messages.length - 1]),
          },
        }),
      }),
    [],
  );
  const chat = useChat({
    id: "xem-workspace",
    transport,
    onFinish: () => {
      void refresh().catch(() => {});
    },
  });
  const busy =
    chat.status === "submitted" || chat.status === "streaming" || acting;
  const open = useCallback(
    async (next: string) => {
      setLoading(true);
      setProblem("");
      try {
        const data = await responseJSON(
          await fetch(`/api/assistant?id=${encodeURIComponent(next)}`, {
            cache: "no-store",
          }),
        );
        current.current = next;
        setId(next);
        chat.setMessages(data.conversation.messages);
        window.history.replaceState(null, "", `/?chat=${next}`);
      } catch (e) {
        setProblem((e as Error).message);
      } finally {
        setLoading(false);
      }
    },
    [chat.setMessages],
  );
  useEffect(() => {
    const next = crypto.randomUUID();
    current.current = next;
    setId(next);
    void refresh()
      .then(async () => {
        const saved = new URLSearchParams(window.location.search).get("chat");
        if (saved) await open(saved);
      })
      .catch((e) => setProblem(e.message))
      .finally(() => setLoading(false));
  }, []);
  function fresh() {
    const next = crypto.randomUUID();
    current.current = next;
    setId(next);
    chat.setMessages([]);
    chat.clearError();
    setInput("");
    setProblem("");
    window.history.replaceState(null, "", "/");
  }
  async function approve(actionId: string, approved: boolean) {
    setActing(true);
    setProblem("");
    try {
      const data = await responseJSON(
        await fetch("/api/assistant/actions", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            conversationId: current.current,
            actionId,
            approved,
          }),
        }),
      );
      chat.setMessages(data.conversation.messages);
    } catch (e) {
      setProblem((e as Error).message);
    } finally {
      setActing(false);
    }
  }
  async function remove(next: string) {
    try {
      const response = await fetch("/api/assistant", {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ conversationId: next }),
      });
      if (!response.ok) await responseJSON(response);
      if (next === current.current) fresh();
      await refresh();
    } catch (e) {
      setProblem((e as Error).message);
    }
  }
  let error = problem || chat.error?.message;
  if (error) {
    try {
      error = JSON.parse(error).error || error;
    } catch {}
  }
  return (
    <ChatView
      messages={chat.messages}
      input={input}
      setInput={setInput}
      busy={busy}
      loading={loading}
      enabled={enabled}
      canWrite={canWrite}
      error={error}
      onSend={(text) => {
        setInput("");
        setProblem("");
        chat.clearError();
        window.history.replaceState(null, "", `/?chat=${current.current}`);
        void chat.sendMessage({ text });
      }}
      onStop={() => void chat.stop()}
      onNew={fresh}
      onAction={(action, approved) => void approve(action, approved)}
      conversations={conversations}
      onOpen={(next) => void open(next)}
      onDelete={(next) => void remove(next)}
      onRecover={() => void open(current.current)}
    />
  );
}
