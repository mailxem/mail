import { ApiClient } from '../core/api-client';
import { SendEmailParams, SendEmailResponse } from '../types';

/**
 * Email resource for sending emails
 */
export class EmailResource {
    constructor(private client: ApiClient) { }

    /**
     * Send an email
     * 
     * @param params - Email parameters
     * @returns Promise resolving to the send response
     * 
     * @example
     * ```typescript
     * const result = await xem.email.send({
     *   to: 'user@example.com',
     *   subject: 'Hello',
     *   html: '<h1>Welcome!</h1>',
     * });
     * ```
     */
    async send(params: SendEmailParams): Promise<SendEmailResponse> {
        // Validate required fields
        if (!params.to) {
            throw new Error('Email recipient (to) is required');
        }

        if (!params.subject) {
            throw new Error('Email subject is required');
        }

        if (!params.html && !params.templateId) {
            throw new Error('Either html or templateId must be provided');
        }

        // Prepare request body
        const body: any = {
            to: params.to,
            subject: params.subject,
        };

        // Add optional fields if provided
        if (params.html) body.html = params.html;
        if (params.templateId) body.templateId = params.templateId;
        if (params.data) body.data = params.data;
        if (params.bcc) body.bcc = params.bcc;
        if (params.cc) body.cc = params.cc;
        if (params.replyTo) body.replyTo = params.replyTo;
        if (params.provider) body.provider = params.provider;
        if (params.scheduleAt) body.scheduleAt = params.scheduleAt;
        if (params.test !== undefined) body.test = params.test;

        return this.client.post<SendEmailResponse>('/email', body);
    }
}
