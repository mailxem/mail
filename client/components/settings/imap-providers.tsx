"use client";

import { useEffect, useState } from "react";
import { UseFormReturn } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  Sheet,
  SheetFooter,
} from "../ui/sheet";
import { Eye, EyeOff } from "lucide-react";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "../ui/form";
import { IMAPConfigSchema } from "@/lib/validations/imap-provider";
import { toast } from "sonner";

export function IMAPProviders({
  onTestConnection,
  onSaveProvider,
  isDialogOpen,
  setIsDialogOpen,
  form,
}: {
  onTestConnection: (data: z.infer<typeof IMAPConfigSchema>) => void;
  onSaveProvider: (data: z.infer<typeof IMAPConfigSchema>) => void;
  isDialogOpen: boolean;
  setIsDialogOpen: (open: boolean) => void;
  form: UseFormReturn<z.infer<typeof IMAPConfigSchema>>;
}) {
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    if (form.formState.errors && Object.keys(form.formState.errors).length > 0) {
      toast.error(
        "Please check the following fields:\n" +
          Object.values(form.formState.errors).map((error) => error.message)
      );
    }
  }, [form.formState.errors]);

  return (
    <Sheet open={isDialogOpen} onOpenChange={setIsDialogOpen}>
      <SheetContent className="overflow-y-auto w-auto p-0">
        <SheetHeader className="p-4">
          <SheetTitle>Add IMAP Provider</SheetTitle>
          <SheetDescription>
            Add a new IMAP service provider configuration
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSaveProvider)}>
            <div className="p-4 space-y-4">
              <div className="space-y-2">
                <Label htmlFor="host">Host</Label>
                <FormField
                  control={form.control}
                  name="host"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Input
                          id="host"
                          placeholder="imap.example.com"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="port">Port</Label>
                <FormField
                  control={form.control}
                  name="port"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Input
                          id="port"
                          type="number"
                          placeholder="993"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="username">Username</Label>
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Input
                          id="username"
                          placeholder="john@example.com"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <div className="space-y-2 relative">
                <Label htmlFor="password">Password</Label>
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Input
                          id="password"
                          type={showPassword ? "text" : "password"}
                          placeholder="********"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button
                  variant="ghost"
                  size="icon"
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 top-2/3 -translate-y-1/2"
                >
                  {showPassword ? (
                    <Eye className="w-4 h-4" />
                  ) : (
                    <EyeOff className="w-4 h-4" />
                  )}
                </Button>
              </div>
            </div>
            <SheetFooter className="flex justify-between sticky bottom-0 p-4 bg-background w-full">
              <Button
                type="button"
                variant="outline"
                className="!text-sm"
                onClick={() => onTestConnection(form.getValues())}
              >
                Test Connection
              </Button>
              <Button variant="default" type="submit" className="!text-sm">
                Save Provider
              </Button>
            </SheetFooter>
          </form>
        </Form>
      </SheetContent>
    </Sheet>
  );
}
