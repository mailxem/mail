/**
 * Email provider options
 */
export type EmailProvider = 'CUSTOM' | string;

/**
 * Email send request parameters
 */
export interface SendEmailParams {
    /**
     * Email recipient address
     */
    to: string;

    /**
     * Email subject
     */
    subject: string;

    /**
     * HTML content of the email
     */
    html?: string;

    /**
     * Template ID to use (alternative to html)
     */
    templateId?: string;

    /**
     * Template data for dynamic content
     */
    data?: any[];

    /**
     * BCC recipients
     */
    bcc?: string;

    /**
     * CC recipients
     */
    cc?: string;

    /**
     * Reply-to address
     */
    replyTo?: string;

    /**
     * Email provider to use
     */
    provider?: EmailProvider;

    /**
     * Schedule email for future delivery (ISO 8601 format)
     */
    scheduleAt?: string;

    /**
     * Test mode - email won't actually be sent
     */
    test?: boolean;
}

/**
 * Email send response
 */
export interface SendEmailResponse {
    success: boolean;
    messageId?: string;
    [key: string]: any;
}

/**
 * API error response
 */
export interface ApiError {
    message: string;
    statusCode?: number;
    details?: any;
}

/**
 * SDK configuration options
 */
export interface XemConfig {
    /**
     * API key for authentication
     */
    apiKey?: string;

    /**
     * Base URL for the API (default: https://api.xem.email/api/v1)
     */
    baseUrl?: string;

    /**
     * Request timeout in milliseconds (default: 30000)
     */
    timeout?: number;
}
