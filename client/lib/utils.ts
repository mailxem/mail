import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import jwt, { JwtPayload } from "jsonwebtoken";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: Date): string {
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(date));
}

export function isEmpty(value: string | undefined | null | boolean): boolean {
  if (typeof value === "boolean") {
    return !value;
  }
  return (
    !value ||
    value === undefined ||
    value === null ||
    value === "" ||
    value.length === 0
  );
}

export function extractVariables(html: string): string[] {
  const regex = /{{([^}]+)}}/g;
  const variables = new Set<string>();
  let match: string[];
  while ((match = regex.exec(html)) !== null) {
    variables.add(match[1].trim());
  }
  return Array.from(variables);
}

export function removeUndefined<T>(obj: T): T {
  return Object.fromEntries(
    Object.entries(obj).filter(([_, v]) => !isEmpty(v))
  ) as T;
}

export function createJwtFromSecret(
  payload: Record<string, string>,
  secret: string
) {
  const jwt = require("jsonwebtoken");
  return jwt.sign(payload, secret);
}

export function deepCompare(a: any, b: any): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

export function encodeToBase64(str: string) {
  return Buffer.from(str).toString("base64");
}

export function decodeFromBase64(str: string) {
  return Buffer.from(str, "base64").toString("utf-8");
}

function jwtDecode(jwtToken: string): JwtPayload {
  const decoded = jwt.decode(jwtToken);
  return decoded as JwtPayload;
}

export function isJwtExpired(jwtToken: string) {
  const decoded = jwtDecode(jwtToken);
  return decoded.exp < Date.now() / 1000;
}

export function generateKey() {
  const key = "23uj429u49uyhh28h4rj23ijri23jruh23hrio34jr34r34r";
  // we need to jumble the key
  const jumble = key.split("").sort(() => Math.random() - 0.5).join("");
  return jumble;
}