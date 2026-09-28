# @xem.email/sdk

Official Node.js TypeScript SDK for the XEM Email API.

## Installation

```bash
npm install @xem.email/sdk
```

## Quick Start

```typescript
import { Xem } from '@xem.email/sdk';

// Initialize with API key
const xem = new Xem({ apiKey: 'your-api-key' });

// Send an email
const result = await xem.email.send({
  to: 'user@example.com',
  subject: 'Welcome to XEM!',
  html: '<h1>Hello World</h1>',
});
```

## Usage

### Initialization

There are multiple ways to initialize the SDK:

#### Option 1: Pass API key in constructor

```typescript
const xem = new Xem({ apiKey: 'your-api-key' });
```

#### Option 2: Login after initialization

```typescript
const xem = new Xem();
xem.login('your-api-key');
```

#### Option 3: Custom configuration

```typescript
const xem = new Xem({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.xem.email/api/v1', // Optional
  timeout: 30000, // Optional, in milliseconds
});
```

### Sending Emails

#### Basic email

```typescript
await xem.email.send({
  to: 'recipient@example.com',
  subject: 'Hello!',
  html: '<p>This is a test email</p>',
});
```

#### Email with CC and BCC

```typescript
await xem.email.send({
  to: 'recipient@example.com',
  cc: 'cc@example.com',
  bcc: 'bcc@example.com',
  subject: 'Team Update',
  html: '<p>Important team announcement</p>',
});
```

#### Using templates

```typescript
await xem.email.send({
  to: 'user@example.com',
  subject: 'Welcome!',
  templateId: 'welcome-template',
  data: [
    { name: 'John', email: 'john@example.com' }
  ],
});
```

#### Scheduled emails

```typescript
await xem.email.send({
  to: 'user@example.com',
  subject: 'Reminder',
  html: '<p>This is your scheduled reminder</p>',
  scheduleAt: '2026-03-01T10:00:00Z', // ISO 8601 format
});
```

#### Test mode

```typescript
await xem.email.send({
  to: 'test@example.com',
  subject: 'Test Email',
  html: '<p>Testing</p>',
  test: true, // Email won't actually be sent
});
```

#### Custom provider

```typescript
await xem.email.send({
  to: 'user@example.com',
  subject: 'Custom Provider',
  html: '<p>Using custom email provider</p>',
  provider: 'CUSTOM',
});
```

#### Reply-to address

```typescript
await xem.email.send({
  to: 'customer@example.com',
  subject: 'Support Ticket #123',
  html: '<p>Your support ticket has been updated</p>',
  replyTo: 'support@example.com',
});
```

## API Reference

### `Xem`

Main SDK client class.

#### Constructor

```typescript
new Xem(config?: XemConfig)
```

**Parameters:**
- `config.apiKey` (optional) - Your XEM API key
- `config.baseUrl` (optional) - Custom API base URL (default: `https://api.xem.email/api/v1`)
- `config.timeout` (optional) - Request timeout in milliseconds (default: `30000`)

#### Methods

##### `login(apiKey: string): void`

Authenticate with your API key.

```typescript
xem.login('your-api-key');
```

##### `isAuthenticated(): boolean`

Check if the client is authenticated.

```typescript
if (xem.isAuthenticated()) {
  // Ready to make API calls
}
```

### `xem.email`

Email resource for sending emails.

#### `send(params: SendEmailParams): Promise<SendEmailResponse>`

Send an email.

**Parameters:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `to` | `string` | ✅ | Recipient email address |
| `subject` | `string` | ✅ | Email subject |
| `html` | `string` | * | HTML content (required if `templateId` not provided) |
| `templateId` | `string` | * | Template ID (required if `html` not provided) |
| `data` | `any[]` | ❌ | Template data for dynamic content |
| `cc` | `string` | ❌ | CC recipients |
| `bcc` | `string` | ❌ | BCC recipients |
| `replyTo` | `string` | ❌ | Reply-to address |
| `provider` | `string` | ❌ | Email provider (default: `CUSTOM`) |
| `scheduleAt` | `string` | ❌ | Schedule for future delivery (ISO 8601) |
| `test` | `boolean` | ❌ | Test mode - email won't be sent |

## TypeScript Support

This SDK is written in TypeScript and includes full type definitions. All types are exported from the main package:

```typescript
import { Xem, SendEmailParams, SendEmailResponse, XemConfig } from '@xem.email/sdk';
```

## Error Handling

The SDK throws errors for invalid parameters and API errors. Always wrap API calls in try-catch blocks:

```typescript
try {
  await xem.email.send({
    to: 'user@example.com',
    subject: 'Test',
    html: '<p>Hello</p>',
  });
} catch (error) {
  console.error('Failed to send email:', error.message);
}
```

## Development

### Building

```bash
npm run build
```

### Watch mode

```bash
npm run dev
```

## License

MIT
