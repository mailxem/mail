"use client";

import { useEffect } from "react";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import { toast } from "sonner";
import { QueryState } from "@/components/marketing/shared";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useSession } from "next-auth/react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

const formSchema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Invalid email address"),
  bio: z.string().optional(),
});

type FormValues = z.infer<typeof formSchema>;

export function ProfileSettings() {
  const session = useSession();
  const query = useMarketingQuery<{firstName:string;lastName:string;email:string;bio?:string}>("users/me");
  const { request } = useMarketing();

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: session.data?.user.name || "",
      email: session.data?.user.email || "",
      bio: "",
    },
  });

  const onSubmit = async (data: FormValues) => {
    try {
      const [firstName, ...rest] = data.name.trim().split(/\s+/);
      await request("marketing/profile", "PUT", { firstName, lastName: rest.join(" "), bio: data.bio || "" });
      form.reset(data);
      await query.refetch();
      toast.success("Profile updated");
    } catch (error) { toast.error((error as Error).message); }
  };

  useEffect(() => {
    if (query.data && !form.formState.isDirty) form.reset({name: [query.data.firstName,query.data.lastName].filter(Boolean).join(" "), email: query.data.email, bio: query.data.bio || ""});
  }, [query.data, form]);
  if (query.isPending || query.error) return <QueryState loading={query.isPending} error={query.error} retry={() => void query.refetch()}/>;
  return (
    <div className="space-y-6">
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <FormField
            control={form.control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="email"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl>
                  <Input type="email" readOnly {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="bio"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Bio</FormLabel>
                <FormControl>
                  <Textarea {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button type="submit" disabled={form.formState.isSubmitting}>{form.formState.isSubmitting ? "Saving…" : "Save changes"}</Button>
        </form>
      </Form>
    </div>
  );
}