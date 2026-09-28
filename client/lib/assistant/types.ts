import type { UIMessage } from "ai";
export type ActionStatus =
  | "pending"
  | "running"
  | "completed"
  | "declined"
  | "unknown"
  | "failed";
export interface AssistantAction {
  id: string;
  tool: string;
  input: Record<string, unknown>;
  schemaHash: string;
  status: ActionStatus;
  createdAt: number;
  result?: unknown;
}
export interface Conversation {
  id: string;
  title: string;
  createdAt: number;
  updatedAt: number;
  messages: UIMessage[];
  actions: AssistantAction[];
}
export interface ConversationSummary {
  id: string;
  title: string;
  updatedAt: number;
}
export interface AssistantScope {
  userId: string;
  teamId: string;
  canWrite: boolean;
  accessToken: string;
}
export interface ActionOutput {
  kind: "xem-action";
  actionId: string;
  tool: string;
  status: ActionStatus;
  input: Record<string, unknown>;
  result?: unknown;
}
