import { z } from "zod";

export const IMAPConfigSchema = z.object({
  host: z.string().min(1, { message: "Host is required" }),
  port: z.number().min(1, { message: "Port is required" }),
  username: z.string().min(1, { message: "Username is required" }),
  password: z.string().min(1, { message: "Password is required" }),
  id: z.any().optional(),
});

export type IMAPConfig = z.infer<typeof IMAPConfigSchema>;