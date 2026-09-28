import type { NextAuthConfig, Session } from "next-auth";
import type { JWT } from "next-auth/jwt";
import { User } from "./lib";
import { APIService } from "./lib/services";
import { API_URL } from "./auth";

export const authConfig = {
  callbacks: {
    async signIn({ user, account, profile }) {
      // Check if the provider is Google and if email is verified
      if (account?.provider === "google" && profile?.email && profile?.email_verified) {
        const authService = new APIService("auth", {
          accessToken: account.access_token,
        });

        const response = await authService.get<{
          token: string;
          refreshToken: string;
        }>("google/callback");

        // Get user details from our backend
        const userResponse = await fetch(`${API_URL}/users/me`, {
          headers: {
            Authorization: `Bearer ${response.token}`,
          },
        });

        const userData = await userResponse.json();

        if (!userResponse.ok) {
          throw new Error("Failed to get user details");
        }

        user.id = userData.id;
        user.email = userData.email;
        user.name = userData["firstName"] + " " + userData["lastName"];
        user.image = userData.image;
        user.role = userData.role;
        user.teamId = userData.teamId;
        user.accessToken = response.token;
        user.refreshToken = response["refresh_token"];
        return true;
      }
      return true;
    },
    // @ts-ignore
    session({
      session,
      user,
      token,
    }: {
      session: Session;
      user: User;
      token: JWT;
    }) {
      return {
        ...session,
        user: {
          ...session.user,
          id: token.sub,
          teamId: token.teamId || session.user.teamId,
        },
      };
    },
    // @ts-ignore
    jwt({ token, user }: { token: JWT; user: CustomAuthUser }) {
      if (user) {
        token.teamId = user.teamId;
      }
      return token;
    },
    authorized({ auth, request: { nextUrl } }) {
      const isLoggedIn = !!auth?.user;
      const isApiRoute = nextUrl.pathname.startsWith("/api");
      const isOnDashboard = nextUrl.pathname.startsWith("/");
      const isOnLogin = nextUrl.pathname.startsWith("/login");

      // Allow API routes to handle their own auth
      if (isApiRoute) return true;

      if (isOnLogin) {
        if (isLoggedIn) return Response.redirect(new URL("/", nextUrl));
        return true;
      }

      if (isOnDashboard) {
        if (isLoggedIn) return true;
        return false;
      }

      return true;
    },
  },
  providers: [],
} satisfies NextAuthConfig;
