import { ApiError } from "@/lib";
import axios from "axios";

export class APIService {
  private readonly baseUrl: string;
  private readonly accessToken: any;

  constructor(endpoint: string, session?: any) {
    let baseUrl = "";
    if (typeof window === "undefined") {
      baseUrl = process.env.INTERNAL_API_URL as string;
    } else {
      const { API_BASE_URL } = require("@/hooks/use-api");
      baseUrl = API_BASE_URL as string;
    }
    this.baseUrl = `${baseUrl?.replace(/\/$/, "")}/${endpoint}`;
    this.accessToken = session?.accessToken;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const body = await response.text();
      let message = "API request failed";
      try { const error = JSON.parse(body); message = error.message || error.error || message; } catch {}
      throw new Error(message);
    }
    if (response.status === 204) return true as T;
    return response.json();
  }

  async get<T>(extraUrl: string, params: Record<string, any> = {}): Promise<T> {
    try {
      const requestUrl = `${this.baseUrl}${extraUrl ? `/${extraUrl.replace(/^\/+/, "")}` : ""}${
        Object.keys(params).length
          ? `?${new URLSearchParams(params).toString()}`
          : ""
      }`;
      const response = await fetch(requestUrl, {
        headers: {
          Authorization: `Bearer ${this.accessToken}`,
          "Content-Type": "application/json",
        },
      });
      return this.handleResponse<T>(response);
    } catch (error) {
      throw this.formatError(error);
    }
  }

  async post<T>(extraUrl: string, data: Record<string, any>): Promise<T> {
    try {
      const response = await fetch(
        `${this.baseUrl}${extraUrl ? `/${extraUrl.replace(/^\/+/, "")}` : ""}`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${this.accessToken}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify(data),
        }
      );
      return this.handleResponse<T>(response);
    } catch (error) {
      throw this.formatError(error);
    }
  }

  async update<T>(id: string, data: Record<string, any>): Promise<T> {
    try {
      const response = await fetch(`${this.baseUrl}/${id}`, {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${this.accessToken}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });
      return this.handleResponse<T>(response);
    } catch (error) {
      throw this.formatError(error);
    }
  }

  async delete<T>(id: string): Promise<T> {
    try {
      const response = await fetch(`${this.baseUrl}/${id}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${this.accessToken}`,
          "Content-Type": "application/json",
        },
      });
      return this.handleResponse<T>(response);
    } catch (error) {
      throw this.formatError(error);
    }
  }

  private formatError(error: unknown): ApiError {
    if (error instanceof Error) {
      return {
        message: error.message,
        code: "UNKNOWN_ERROR",
      };
    }
    return {
      message: "An unknown error occurred",
      code: "UNKNOWN_ERROR",
    };
  }

  async upload(file: File | Blob): Promise<{ file: string; url: string }> {
    const formData = new FormData();
    formData.append("file", file, file.name);
    const response = await axios.post(`${this.baseUrl}/upload`, formData, {
      headers: {
        Authorization: `Bearer ${this.accessToken}`,
        "Content-Type": "multipart/form-data",
      },
    });
    return response.data as { file: string; url: string };
  }
}
