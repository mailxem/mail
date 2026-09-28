"use client";
import "@maily-to/core/style.css";
import { Editor } from "@maily-to/core";
import type { Editor as TiptapEditor } from "@tiptap/core";

// Reuse the original compose editor and its email-safe renderer. This is a
// personal-message canvas; the template builder continues to use Unlayer.
export function RichMessageEditor({ onReady, onChange, editable = true }: {
  onReady: (editor: TiptapEditor) => void;
  onChange: (editor: TiptapEditor) => void;
  editable?: boolean;
}) {
  return <div className="overflow-hidden rounded-xl border border-input bg-white">
    <Editor
      contentJson={{ type: "doc", content: [{ type: "paragraph", content: [] }] }}
      onCreate={onReady}
      onUpdate={onChange}
      editable={editable}
      config={{ immediatelyRender: false, hasMenuBar: true, spellCheck: true,
        wrapClassName: "!max-w-none !border-0 !rounded-none",
        bodyClassName: "!m-0 !bg-white !p-0",
        contentClassName: "min-h-64 !px-6 !py-4 text-base leading-relaxed",
      }}
    />
    <p className="border-t border-border px-4 py-2 text-xs text-muted-foreground">Select text to format it. Type / to add an image, list, or other content.</p>
  </div>;
}
