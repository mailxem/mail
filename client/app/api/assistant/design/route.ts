import { editorProxy } from "@/lib/assistant/editor-proxy";
export const runtime = "nodejs";
export const maxDuration = 120;
export const POST = (request: Request) => editorProxy(request, "draft");
