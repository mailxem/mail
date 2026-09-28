"use client";

import { useSession } from "next-auth/react";

type ApiFetchOptions = RequestInit & { requireAuth?: boolean };

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "";

export function useApi() {
  const { data: session, status } = useSession();

  const apiFetch = async (path: string, options: ApiFetchOptions = {}) => {
    const requireAuth = options.requireAuth ?? true;
    const token = session?.accessToken;

    if (requireAuth && !token) {
      throw new Error("Not authenticated");
    }

    const headers = new Headers(options.headers);
    if (requireAuth && token && !headers.has("authorization")) {
      headers.set("authorization", `Bearer ${token}`);
    }

    if (
      options.body &&
      !(options.body instanceof FormData) &&
      !headers.has("content-type")
    ) {
      headers.set("content-type", "application/json");
    }

    const normalizedPath = path.replace(/^\/+/, "");
    const url = `${API_BASE_URL}/${normalizedPath}`;

    const response = await fetch(url, {
      ...options,
      headers,
    });

    return response;
  };

  return { apiFetch, session, status };
}
