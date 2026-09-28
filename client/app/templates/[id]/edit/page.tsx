"use client";

import { TemplateEditor } from "@/components/templates/template-editor";
import { Button } from "@/components/ui/button";
import { use } from "react";
import { useRouter } from "next/navigation";

export default function EditTemplatePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();
  return (
    <div className="p-6 space-y-8">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          onClick={() => router.back()}
          className="text-gray-600 hover:"
        >
          ←
        </Button>
        <h1 className="text-xl font-medium ">Edit Template</h1>
      </div>
      <TemplateEditor templateId={id} />
    </div>
  );
}
