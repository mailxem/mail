# Email Flow Builder

A React component for building email marketing campaign flows using React Flow.

## Installation

```bash
npm install @your-org/email-flow-builder @xyflow/react react-icons
```

## Usage

```tsx
import { EmailFlowBuilder } from '@your-org/email-flow-builder';
import '@xyflow/react/dist/base.css';
import '@your-org/email-flow-builder/dist/styles.css';

export default function CampaignPage() {
  return (
    <div style={{ width: '100vw', height: '100vh' }}>
      <EmailFlowBuilder />
    </div>
  );
}
```

## Features

- Visual email campaign flow builder
- Drag and drop interface
- Pre-built node types for common email campaign elements:
  - Segment (audience targeting)
  - Email Campaign
  - Time Delay
  - Split Test (A/B testing)
  - Trigger actions
  - Alerts
  - Filters
  - Tag Management
  - Goal tracking

## Customization

You can customize the appearance using CSS variables:

```css
.react-flow {
  --bg-color: #000000;
  --text-color: #ffffff;
  --node-border-radius: 12px;
  --node-box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.3);
}
```

## Types

The package exports TypeScript types for use in your application:

```tsx
import type { 
  TurboNodeData, 
  NodeConfig, 
  EmailFlowState 
} from '@your-org/email-flow-builder';
```

## License

MIT 