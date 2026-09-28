import axios, { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios';
import { ApiError } from '../types';

/**
 * Core API client for making HTTP requests
 */
export class ApiClient {
    private client: AxiosInstance;
    private apiKey: string | null = null;

    constructor(baseUrl: string, timeout: number = 30000) {
        this.client = axios.create({
            baseURL: baseUrl,
            timeout,
            headers: {
                'Content-Type': 'application/json',
            },
        });

        // Response interceptor for error handling
        this.client.interceptors.response.use(
            (response) => response,
            (error: AxiosError) => {
                return Promise.reject(this.handleError(error));
            }
        );
    }

    /**
     * Set the API key for authentication
     */
    setApiKey(apiKey: string): void {
        this.apiKey = apiKey;
    }

    /**
     * Get the current API key
     */
    getApiKey(): string | null {
        return this.apiKey;
    }

    /**
     * Make a POST request
     */
    async post<T = any>(
        path: string,
        data?: any,
        config?: AxiosRequestConfig
    ): Promise<T> {
        this.ensureAuthenticated();

        const response = await this.client.post<T>(path, data, {
            ...config,
            headers: {
                ...config?.headers,
                'X-API-KEY': this.apiKey!,
            },
        });

        return response.data;
    }

    /**
     * Make a GET request
     */
    async get<T = any>(
        path: string,
        config?: AxiosRequestConfig
    ): Promise<T> {
        this.ensureAuthenticated();

        const response = await this.client.get<T>(path, {
            ...config,
            headers: {
                ...config?.headers,
                'X-API-KEY': this.apiKey!,
            },
        });

        return response.data;
    }

    /**
     * Make a PUT request
     */
    async put<T = any>(
        path: string,
        data?: any,
        config?: AxiosRequestConfig
    ): Promise<T> {
        this.ensureAuthenticated();

        const response = await this.client.put<T>(path, data, {
            ...config,
            headers: {
                ...config?.headers,
                'X-API-KEY': this.apiKey!,
            },
        });

        return response.data;
    }

    /**
     * Make a DELETE request
     */
    async delete<T = any>(
        path: string,
        config?: AxiosRequestConfig
    ): Promise<T> {
        this.ensureAuthenticated();

        const response = await this.client.delete<T>(path, {
            ...config,
            headers: {
                ...config?.headers,
                'X-API-KEY': this.apiKey!,
            },
        });

        return response.data;
    }

    /**
     * Ensure API key is set before making requests
     */
    private ensureAuthenticated(): void {
        if (!this.apiKey) {
            throw new Error(
                'API key not set. Please call login() or pass apiKey in constructor.'
            );
        }
    }

    /**
     * Handle and normalize API errors
     */
    private handleError(error: AxiosError): ApiError {
        if (error.response) {
            // Server responded with error status
            return {
                message:
                    (error.response.data as any)?.message ||
                    error.message ||
                    'An error occurred',
                statusCode: error.response.status,
                details: error.response.data,
            };
        } else if (error.request) {
            // Request made but no response
            return {
                message: 'No response from server',
                details: error.message,
            };
        } else {
            // Error in request setup
            return {
                message: error.message || 'Request failed',
            };
        }
    }
}
