import { ApiClient } from './core/api-client';
import { EmailResource } from './resources/email';
import { XemConfig } from './types';

/**
 * Main XEM SDK client
 * 
 * @example
 * ```typescript
 * // Initialize with API key
 * const xem = new Xem({ apiKey: 'your-api-key' });
 * 
 * // Or initialize and login later
 * const xem = new Xem();
 * xem.login('your-api-key');
 * 
 * // Send an email
 * await xem.email.send({
 *   to: 'user@example.com',
 *   subject: 'Hello',
 *   html: '<h1>Welcome!</h1>',
 * });
 * ```
 */
export class Xem {
    private apiClient: ApiClient;

    /**
     * Email resource for sending emails
     */
    public readonly email: EmailResource;

    /**
     * Create a new XEM client
     * 
     * @param config - Configuration options
     */
    constructor(config: XemConfig = {}) {
        const baseUrl = config.baseUrl || 'https://api.xem.email/api/v1';
        const timeout = config.timeout || 30000;

        this.apiClient = new ApiClient(baseUrl, timeout);

        // Set API key if provided in config
        if (config.apiKey) {
            this.apiClient.setApiKey(config.apiKey);
        }

        // Initialize resources
        this.email = new EmailResource(this.apiClient);
    }

    /**
     * Authenticate with API key
     * 
     * @param apiKey - Your XEM API key
     * 
     * @example
     * ```typescript
     * const xem = new Xem();
     * xem.login('your-api-key');
     * ```
     */
    login(apiKey: string): void {
        if (!apiKey || typeof apiKey !== 'string') {
            throw new Error('API key must be a non-empty string');
        }

        this.apiClient.setApiKey(apiKey);
    }

    /**
     * Check if the client is authenticated
     * 
     * @returns true if API key is set
     */
    isAuthenticated(): boolean {
        return this.apiClient.getApiKey() !== null;
    }
}
