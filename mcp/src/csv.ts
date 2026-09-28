import { parse } from "csv-parse/sync";
import { z } from "zod";

export const contactSchema = z
  .object({
    email: z
      .string()
      .trim()
      .email()
      .max(254)
      .transform((v) => v.toLowerCase()),
    firstName: z.string().max(500).optional(),
    lastName: z.string().max(500).optional(),
    phone: z.string().max(500).optional(),
    company: z.string().max(500).optional(),
    country: z.string().max(500).optional(),
    city: z.string().max(500).optional(),
  })
  .strict();
export const mappingSchema = z
  .object({
    email: z.string().min(1),
    firstName: z.string().min(1).optional(),
    lastName: z.string().min(1).optional(),
    phone: z.string().min(1).optional(),
    company: z.string().min(1).optional(),
    country: z.string().min(1).optional(),
    city: z.string().min(1).optional(),
  })
  .strict();
export function mapCSV(csv: string, mappings: z.infer<typeof mappingSchema>) {
  if (Buffer.byteLength(csv) > 1024 * 1024)
    throw new Error("CSV exceeds 1 MiB");
  let rows: string[][];
  try {
    rows = parse(csv, {
      bom: true,
      skip_empty_lines: true,
      max_record_size: 16_384,
      to: 502,
    });
  } catch {
    throw new Error(
      "Invalid CSV: check quoting, column counts and record size",
    );
  }
  if (rows.length < 2 || rows.length > 501)
    throw new Error("CSV must contain headers and 1 to 500 rows");
  const headers = rows[0].map((h) => h.trim());
  if (headers.some((h) => !h) || new Set(headers).size !== headers.length)
    throw new Error("CSV headers must be nonempty and unique");
  const entries = Object.entries(mappings);
  if (new Set(entries.map(([, h]) => h)).size !== entries.length)
    throw new Error("Map each CSV column only once");
  for (const [, header] of entries)
    if (!headers.includes(header))
      throw new Error("A mapped column is missing from CSV headers");
  return rows.slice(1).map((row, index) => {
    const candidate = Object.fromEntries(
      entries.map(([field, header]) => [
        field,
        row[headers.indexOf(header)].trim(),
      ]),
    );
    const result = contactSchema.safeParse(candidate);
    if (
      !result.success ||
      Object.values(candidate).some((v) => v.includes("\0"))
    )
      throw new Error(`Invalid contact at CSV record ${index + 2}`);
    return result.data;
  });
}

// Quote every field and neutralize spreadsheet formula prefixes, including whitespace.
export function encodeCSV(rows: Record<string, unknown>[]): string {
  const columns = [
    "id",
    "email",
    "firstName",
    "lastName",
    "phone",
    "company",
    "country",
    "city",
    "status",
    "listId",
  ];
  const cell = (value: unknown) => {
    let text = typeof value === "string" ? value : "";
    if (/^[\s]*[=+\-@]/.test(text) || /^[\t\r\n]/.test(text)) text = "'" + text;
    return '"' + text.replace(/"/g, '""') + '"';
  };
  return (
    [columns, ...rows.map((row) => columns.map((key) => row[key]))]
      .map((row) => row.map(cell).join(","))
      .join("\r\n") + "\r\n"
  );
}
