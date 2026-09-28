import { SMTPProviderType } from "@/lib";
import { z } from "zod";

export const formSchema = z.object({
  id: z.any().optional(),
  provider: z.nativeEnum(SMTPProviderType),
  host: z.string().min(1, "Host is required"),
  port: z
    .union([z.number(), z.string()])
    .transform((val) => (typeof val === "string" ? parseInt(val) : val)),
  username: z.string().min(1, "Username is required"),
  requiresAuth: z.boolean().default(true),
  fromEmail: z.string().min(1, "From Email is required"),
  password: z.string().min(1, "Password is required"),
  isActive: z.boolean().default(true),
  supportsTls: z.boolean().default(true),
  maxSendRate: z
    .union([z.number(), z.string()])
    .transform((val) => (typeof val === "string" ? parseInt(val) : val)),
  documentation: z.string().optional(),
  isDefault: z.boolean().default(false),
});

export type SMTPConfig = z.infer<typeof formSchema>;
