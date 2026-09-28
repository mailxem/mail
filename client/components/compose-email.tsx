"use client";
import "@maily-to/core/style.css";
import { useState } from "react";
import { Send } from "lucide-react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Editor } from "@maily-to/core";
import type { Editor as TiptapEditor } from "@tiptap/core";
import { Textarea } from "./ui/textarea";
import { JSONContent } from "@tiptap/react";
import { Maily } from "@maily-to/render";

interface ComposeEmailProps {
  onSend?: (email: {
    to: string; 
    subject: string;
    html: string;
    cc: string;
    bcc: string;
  }) => void;
}

export function ComposeEmail({ onSend }: ComposeEmailProps) {
  const [to, setTo] = useState("");
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState<JSONContent>();
  const [cc, setCc] = useState("");
  const [bcc, setBcc] = useState("");
  const [data, setData] = useState<Record<string, any>>({});
  const [editor, setEditor] = useState<TiptapEditor>();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const maily = new Maily(editor?.getJSON());

    const html = await maily.render();

    if (onSend) {
      onSend({
        to: to,
        subject,
        html: html,
        cc: cc.trim().split(",").filter(Boolean).join(","),
        bcc: bcc.trim().split(",").filter(Boolean).join(","),
      });
    }

    // Reset form
    setTo("");
    setSubject("");
    setBody(undefined);
    setCc("");
    setBcc("");
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <div className="flex flex-col gap-2">
        <Label htmlFor="to">To</Label>
        <Input
          type="email"
          placeholder="To"
          value={to}
          onChange={(e) => setTo(e.target.value)}
          required
        />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="cc">CC</Label>
        <Input
          type="text"
          placeholder="Enter comma separated emails"
          value={cc}
          onChange={(e) => setCc(e.target.value)}
        />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="bcc">BCC</Label>
        <Input
          type="text"
          placeholder="Enter comma separated emails"
          value={bcc}
          onChange={(e) => setBcc(e.target.value)}
        />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="subject">Subject</Label>
        <Input
          type="text"
          placeholder="Subject"
          value={subject}
          onChange={(e) => setSubject(e.target.value)}
        />
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="variables">Template Variables (JSON)</Label>
        <Textarea
          placeholder='{"name": "John", "company": "Acme Inc"}'
          value={JSON.stringify(data)}
          onChange={(e) => {
            try {
              // Validate JSON format
              if (e.target.value) {
                JSON.parse(e.target.value);
              }
              setData(JSON.parse(e.target.value));
            } catch (err) {
              // Let user continue typing even if invalid JSON
              setData(JSON.parse(e.target.value));
            }
          }}
        />
        <p className="text-xs text-muted-foreground">
          Enter variables in JSON format to use in your email template. Example:{" "}
          {"{"}"name": "John"{"}"}
        </p>
      </div>
      <div className="flex flex-col gap-2">
        <Label htmlFor="body">Body</Label>
        <Editor
          onUpdate={(editor) => {
            setEditor(editor);
          }}
          onCreate={(editor) => {
            setEditor(editor);
          }}
          contentJson={body}
        />
      </div>

      <Button type="submit" className="self-end">
        <Send className="mr-2" />
        Send Email
      </Button>
    </form>
  );
}
