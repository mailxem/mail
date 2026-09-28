# Quick Reference Guide

## Installation

```bash
npm install @xem.email/sdk
```

## Basic Usage

```typescript
import { Xem } from '@xem.email/sdk';

// Method 1: Login after initialization
const xem = new Xem();
xem.login('your-api-key');

// Method 2: Pass API key in constructor
const xem = new Xem({ apiKey: 'your-api-key' });

// Send email
await xem.email.send({
  to: 'user@example.com',
  subject: 'Hello!',
  html: '<h1>Welcome</h1>',
});
```

## API Methods

### `xem.login(apiKey: string)`
Authenticate with your API key.

### `xem.isAuthenticated(): boolean`
Check if authenticated.

### `xem.email.send(params): Promise<SendEmailResponse>`
Send an email.

## Email Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `to` | string | ✅ Yes | Recipient email |
| `subject` | string | ✅ Yes | Email subject |
| `html` | string | * | HTML content |
| `templateId` | string | * | Template ID |
| `data` | any[] | No | Template data |
| `cc` | string | No | CC recipients |
| `bcc` | string | No | BCC recipients |
| `replyTo` | string | No | Reply-to address |
| `provider` | string | No | Email provider |
| `scheduleAt` | string | No | Schedule time (ISO 8601) |
| `test` | boolean | No | Test mode |

*Either `html` or `templateId` must be provided.

## Common Patterns

### Send HTML Email
```typescript
await xem.email.send({
  to: 'user@example.com',
  subject: 'Welcome!',
  html: '<h1>Hello World</h1>',
});
```

### Use Template
```typescript
await xem.email.send({
  to: 'user@example.com',
  subject: 'Welcome!',
  templateId: 'welcome-template',
  data: [{ name: 'John' }],
});
```

### Schedule Email
```typescript
await xem.email.send({
  to: 'user@example.com',
  subject: 'Reminder',
  html: '<p>Don\'t forget!</p>',
  scheduleAt: '2026-03-01T10:00:00Z',
});
```

### Test Mode
```typescript
await xem.email.send({
  to: 'test@example.com',
  subject: 'Test',
  html: '<p>Testing</p>',
  test: true, // Won't actually send
});
```

### With CC/BCC
```typescript
await xem.email.send({
  to: 'user@example.com',
  cc: 'manager@example.com',
  bcc: 'archive@example.com',
  subject: 'Report',
  html: '<p>Monthly report</p>',
});
```

## Error Handling

```typescript
try {
  await xem.email.send({...});
} catch (error) {
  console.error('Error:', error.message);
  console.error('Status:', error.statusCode);
  console.error('Details:', error.details);
}
```

## Configuration

```typescript
const xem = new Xem({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.xem.email/api/v1', // Optional
  timeout: 30000, // Optional (ms)
});
```

## Environment Variables

You can use environment variables for the API key:

```typescript
const xem = new Xem({
  apiKey: process.env.XEM_API_KEY,
});
```

## TypeScript Types

All types are exported:

```typescript
import {
  Xem,
  SendEmailParams,
  SendEmailResponse,
  XemConfig,
  EmailProvider,
  ApiError,
} from '@xem.email/sdk';
```

## Examples Location

Check the `examples/` directory for more:
- `basic-usage.ts` - Simple email sending
- `advanced-usage.ts` - All features demonstration
- `demo.ts` - Interactive demo script
