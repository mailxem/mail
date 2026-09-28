"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useTeam } from "@/app/providers/team-provider";
import { toast } from "sonner";
import {
  Select,
  SelectItem,
  SelectTrigger,
  SelectContent,
  SelectValue,
} from "@/components/ui/select";
import { useApi } from "@/hooks/use-api";

const inviteSchema = z.object({
  email: z.string().email("Invalid email address"),
  name: z.string().default(""),
  role: z.enum(["ADMIN", "MEMBER", "SUPER_ADMIN"]),
});

type InviteFormData = z.infer<typeof inviteSchema>;

export function InviteTeamMember() {
  const { refreshTeam } = useTeam();
  const { apiFetch } = useApi();
  const [loading, setLoading] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<InviteFormData>({
    resolver: zodResolver(inviteSchema),
  });

  const onSubmit = async (data: InviteFormData) => {
    try {
      setLoading(true);
      const response = await apiFetch("users/invite", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });

      if (!response.ok) throw new Error("Failed to send invite");

      toast.success("Invite sent successfully");
      reset();
      refreshTeam();
    } catch (error) {
      toast.error("Failed to send invite");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div>
        <Input
          {...register("name")}
          type="text"
          placeholder="Name"
          disabled={loading}
        />
      </div>
      <div>
        <Input
          {...register("email")}
          type="email"
          placeholder="Email address"
          disabled={loading}
        />
        {errors.email && (
          <p className="text-sm text-red-500 mt-1">{errors.email.message}</p>
        )}
      </div>

      <div>
        <Select
          {...register("role")}
          disabled={loading}
          onValueChange={(value) => {
            setValue("role", value as "ADMIN" | "MEMBER" | "SUPER_ADMIN");
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select Role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="MEMBER">Team Member</SelectItem>
            <SelectItem value="ADMIN">Admin</SelectItem>
            <SelectItem value="SUPER_ADMIN">Super Admin</SelectItem>
          </SelectContent>
        </Select>
        {errors.role && (
          <p className="text-sm text-red-500 mt-1">{errors.role.message}</p>
        )}
      </div>

      <Button type="submit" disabled={loading}>
        {loading ? "Sending..." : "Send Invite"}
      </Button>
    </form>
  );
}
