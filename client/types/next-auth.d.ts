// next-auth.d.ts
import { DefaultSession, DefaultUser } from "next-auth";
import { JWT } from "next-auth/jwt";

declare module "next-auth" {
  interface Session extends DefaultSession {
    accessToken?: string;
    error?: string;
    user: {
      id: string;
      role: string;
      teamId: string;
    } & DefaultSession["user"];
  }

  interface User extends DefaultUser {
    role?: string;
    teamId?: string;
    accessToken?: string;
    refreshToken?: string;
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    accessToken?: string;
    refreshToken?: string;
    exp?: number;
    role?: string;
    teamId?: string;
    error?: string;
  }
}