# Assistant design notes — 12 September 2026

The owner supplied a Cloudflare account-home screenshot: a restrained sidebar, generous canvas, one main prompt, and a small set of useful routes. Ask Xem follows that hierarchy in the existing Xem light palette. It uses one composer, four task-oriented starters, a compact history menu and inline review cards; the conversation does not introduce another permanent sidebar.

Sources reviewed:

- https://ui.shadcn.com/docs/helpers/ai-sdk.md — real `useChat` lifecycle driven by scripted transport, useful for reliable demos rather than production inference.
- https://resend.com/blog/ai-email-editor — begin with existing content, keep AI available without obscuring the work, review before sending.
- https://resend.com/blog/one-more-ai-thing — small contextual writing improvements and preservation of the user's voice.
- https://resend.com/blog/agent-experience — discoverable capabilities, clear failures and useful documentation for agent workflows.
- https://ai-sdk.dev/docs/ai-sdk-core/mcp-tools — authenticated HTTP MCP, transport lifecycle and tool-definition drift.

The home screen uses existing Geist typography, a quiet off-white canvas and one violet accent. At the owner's request, the active transcript follows Claude's reading layout: a centered 768px column, warm paper canvas, 17px Georgia serif answers, sans-serif user messages in neutral rounded bubbles, no assistant avatar, icon-only copy and a compact expanding composer. Georgia is an available fallback, not Claude's proprietary typeface. Claude's native chat was inspected directly and a design conversation confirmed the hierarchy and recommended collapsing resolved action cards while keeping their details accessible. Tool results are collapsible; mutations display the actual arguments before approval. Markdown is semantic and HTML/images are disabled. The composer respects IME composition, Shift+Enter and reduced motion; scrolling follows new text only while the reader stays near the bottom.

The website uses Xem's EB Garamond editorial typography and lavender section background around a code-rendered animated chat. It explicitly labels example data and development status. Play/pause/replay controls are keyboard accessible; reduced motion shows the completed example without autoplay. No competitor assets or proprietary font files are redistributed; no invented performance results are used.
