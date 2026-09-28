/**
 * @xem.email/sdk - Official Node.js SDK for XEM Email API
 * 
 * @example
 * ```typescript
 * import { Xem } from '@xem.email/sdk';
 * 
 * const xem = new Xem({ apiKey: 'your-api-key' });
 * 
 * await xem.email.send({
 *   to: 'user@example.com',
 *   subject: 'Welcome!',
 *   html: '<h1>Hello World</h1>',
 * });
 * ```
 */

export { Xem } from './client';
export * from './types';
