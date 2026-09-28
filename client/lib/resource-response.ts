/** Single-record endpoints return a record; older adapters may wrap it in data. */
export function resourceEntity<T>(payload: unknown): T {
  const record = payload && typeof payload === "object" && "data" in payload ? payload.data : payload;
  if (!record || typeof record !== "object" || Array.isArray(record) || !("id" in record) || typeof record.id !== "string") {
    throw new Error("The server returned an invalid record. Please try again.");
  }
  return record as T;
}
