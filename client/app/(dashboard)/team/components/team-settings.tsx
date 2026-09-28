"use client";

import { useState, useEffect } from "react";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useTemplates } from "@/app/providers/templates-provider";
import { useTeam } from "@/app/providers/team-provider";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";


export function TeamSettings() {
  const [selectedTemplate, setSelectedTemplate] = useState<string>("");
  const { templates } = useTemplates();
  const { team } = useTeam();
  const { apiFetch } = useApi();

  useEffect(() => {
    if (team?.emailTemplateId) {
      setSelectedTemplate(team.emailTemplateId);
    }
  }, [team]);

  const handleTemplateChange = async (value: string) => {
    setSelectedTemplate(value);
    try {
      const response = await apiFetch("team/settings", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          inviteTemplateId: value,
        }),
      });

      if (!response.ok) throw new Error("Failed to update settings");

      toast.success("Invite template updated successfully");
    } catch (error) {
      toast.error("Failed to update invite template");
    }
  };

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="inviteTemplate">Invitation Email Template</Label>
        <Select
          disabled={templates.length === 0}
          value={selectedTemplate}
          onValueChange={handleTemplateChange}
        >
          <SelectTrigger id="inviteTemplate">
            <SelectValue placeholder="Select a template" />
          </SelectTrigger>
          <SelectContent>
            {templates.map((template) => (
              <SelectItem key={template.id} value={template.id}>
                {template.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </div>
  );
}
