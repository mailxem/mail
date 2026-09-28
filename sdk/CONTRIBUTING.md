# Contributing to @xem.email/sdk

Thank you for your interest in contributing to the XEM Email SDK!

## Development Setup

1. Clone the repository
2. Install dependencies:
   ```bash
   npm install
   ```
3. Build the project:
   ```bash
   npm run build
   ```
4. Watch for changes during development:
   ```bash
   npm run dev
   ```

## Project Structure

```
src/
├── index.ts              # Main entry point
├── client.ts             # Main Xem client class
├── core/
│   └── api-client.ts     # HTTP client with authentication
├── resources/
│   └── email.ts          # Email resource
└── types/
    └── index.ts          # TypeScript type definitions
```

## Adding New Resources

To add a new resource (e.g., `sms`, `webhook`):

1. Create a new file in `src/resources/[resource-name].ts`
2. Define the resource class with methods
3. Add types to `src/types/index.ts`
4. Initialize the resource in `src/client.ts`
5. Export types from `src/index.ts`
6. Update README.md with documentation
7. Add examples to `examples/` directory

Example:

```typescript
// src/resources/sms.ts
import { ApiClient } from '../core/api-client';
import { SendSmsParams, SendSmsResponse } from '../types';

export class SmsResource {
  constructor(private client: ApiClient) {}

  async send(params: SendSmsParams): Promise<SendSmsResponse> {
    return this.client.post('/sms', params);
  }
}

// src/client.ts
export class Xem {
  public readonly email: EmailResource;
  public readonly sms: SmsResource; // Add new resource

  constructor(config: XemConfig = {}) {
    // ...
    this.email = new EmailResource(this.apiClient);
    this.sms = new SmsResource(this.apiClient); // Initialize
  }
}
```

## Code Style

- Use TypeScript strict mode
- Follow existing code patterns
- Add JSDoc comments for public APIs
- Include examples in documentation
- Validate required parameters
- Handle errors appropriately

## Testing

Before submitting a PR:

1. Build the project without errors: `npm run build`
2. Test with the examples in `examples/`
3. Ensure TypeScript types are correct

## Pull Request Process

1. Create a feature branch
2. Make your changes
3. Update documentation
4. Update CHANGELOG.md
5. Submit a pull request

## Questions?

Feel free to open an issue for any questions or concerns.
