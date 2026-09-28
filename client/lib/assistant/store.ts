import Redis from "ioredis";
import { randomUUID } from "node:crypto";
import type { Conversation, ConversationSummary } from "./types";

export const RETENTION_SECONDS = 7 * 24 * 60 * 60;
const memory = new Map<string, { value: string; until: number }>();
let client: Redis | undefined;
function redis() {
  const url = process.env.ASSISTANT_REDIS_URL;
  if (!url) {
    if (process.env.NODE_ENV === "production")
      throw new Error("Assistant storage is not configured");
    return undefined;
  }
  if (!client) {
    client = new Redis(url, {
      maxRetriesPerRequest: 1,
      connectTimeout: 5000,
      commandTimeout: 5000,
      retryStrategy: (n) => (n < 2 ? 100 : null),
    });
    client.on("error", () => {});
  }
  return client;
}
async function get(key: string): Promise<string | null> {
  const db = redis();
  if (db) return db.get(key);
  const row = memory.get(key);
  if (!row || row.until < Date.now()) {
    memory.delete(key);
    return null;
  }
  return row.value;
}
async function put(key: string, value: string, ttl = RETENTION_SECONDS) {
  const db = redis();
  if (db) {
    await db.set(key, value, "EX", ttl);
    return;
  }
  for (const [k, v] of memory) if (v.until < Date.now()) memory.delete(k);
  if (memory.size > 5000) throw new Error("Development storage is full");
  memory.set(key, { value, until: Date.now() + ttl * 1000 });
}
const key = (owner: string, id: string) =>
  `xem:assistant:v1:${owner}:chat:${id}`;
export async function loadConversation(
  owner: string,
  id: string,
): Promise<Conversation | null> {
  const value = await get(key(owner, id));
  return value ? JSON.parse(value) : null;
}
export async function listConversations(
  owner: string,
): Promise<ConversationSummary[]> {
  const ids: string[] = JSON.parse(
    (await get(`xem:assistant:v1:${owner}:index`)) || "[]",
  );
  const rows = await Promise.all(ids.map((id) => loadConversation(owner, id)));
  return rows
    .filter((r): r is Conversation => !!r)
    .map(({ id, title, updatedAt }) => ({ id, title, updatedAt }))
    .sort((a, b) => b.updatedAt - a.updatedAt);
}
export async function saveConversation(owner: string, row: Conversation) {
  if (JSON.stringify(row).length > 500_000)
    throw new Error("Start a new conversation to continue");
  await put(key(owner, row.id), JSON.stringify(row));
  const indexKey = `xem:assistant:v1:${owner}:index`;
  const ids: string[] = JSON.parse((await get(indexKey)) || "[]");
  const all = [row.id, ...ids.filter((id) => id !== row.id)];
  for (const id of all.slice(30)) await deleteConversation(owner, id);
  await put(indexKey, JSON.stringify(all.slice(0, 30)));
}
export async function deleteConversation(owner: string, id: string) {
  const db = redis();
  if (db) await db.del(key(owner, id));
  else memory.delete(key(owner, id));
}
// A distributed user/workspace lock serializes both chat turns and approval
// execution. Compare-and-delete prevents an expired worker unlocking its successor.
export async function acquireLock(
  owner: string,
): Promise<(() => Promise<void>) | null> {
  const lockKey = `xem:assistant:v1:${owner}:lock`;
  const nonce = randomUUID();
  const db = redis();
  if (db) {
    if ((await db.set(lockKey, nonce, "EX", 120, "NX")) !== "OK") return null;
    return async () => {
      await db.eval(
        "if redis.call('get',KEYS[1]) == ARGV[1] then return redis.call('del',KEYS[1]) else return 0 end",
        1,
        lockKey,
        nonce,
      );
    };
  }
  const existing = memory.get(lockKey);
  if (existing && existing.until > Date.now()) return null;
  memory.set(lockKey, { value: nonce, until: Date.now() + 120_000 });
  return async () => {
    if ((await get(lockKey)) === nonce) memory.delete(lockKey);
  };
}
export async function allowRequest(
  identity: string,
  limit: number,
  windowSeconds = 60,
) {
  const bucket = `xem:assistant:v1:rate:${identity}:${Math.floor(Date.now() / (windowSeconds * 1000))}`;
  const db = redis();
  if (db)
    return (
      Number(
        await db.eval(
          "local n=redis.call('incr',KEYS[1]); if n==1 then redis.call('expire',KEYS[1],ARGV[1]) end; return n",
          1,
          bucket,
          windowSeconds + 60,
        ),
      ) <= limit
    );
  const existing = memory.get(bucket);
  const n =
    Number(existing && existing.until > Date.now() ? existing.value : 0) + 1;
  memory.set(bucket, {
    value: String(n),
    until: Date.now() + (windowSeconds + 60) * 1000,
  });
  return n <= limit;
}
