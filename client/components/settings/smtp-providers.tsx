"use client";

import { useState } from "react";
import { UseFormReturn } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Eye,
  EyeOff,
} from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { formSchema, SMTPConfig } from "@/lib/validations/smtp-provider";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { SMTPProviderType } from "@/lib";
import {
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  Sheet,
  SheetFooter,
} from "../ui/sheet";

export function SMTPProviders({
  onTestConnection,
  onSaveProvider,
  isDialogOpen,
  setIsDialogOpen,
  form,
  providers,
  onProviderChange,
}: {
  onTestConnection: (data: z.infer<typeof formSchema>) => void;
  onSaveProvider: (data: z.infer<typeof formSchema>) => void;
  isDialogOpen: boolean;
  setIsDialogOpen: (open: boolean) => void;
  providers: Record<SMTPProviderType, SMTPConfig>;
  onProviderChange: (provider: SMTPProviderType) => void;
  form: UseFormReturn<z.infer<typeof formSchema>>;
}) {
  const [showPassword, setShowPassword] = useState(false);
  return (
    <Sheet open={isDialogOpen} onOpenChange={setIsDialogOpen}>
      <SheetContent className="overflow-y-auto w-auto p-0">
        <SheetHeader className="p-4">
          <SheetTitle>Add SMTP Provider</SheetTitle>
          <SheetDescription>
            Add a new SMTP service provider configuration
          </SheetDescription>
        </SheetHeader>
        <form onSubmit={form.handleSubmit(onSaveProvider)}>
          <div className="p-4 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="provider">Provider Type</Label>
              <Select
                onValueChange={(value) => {
                  const provider = providers[value as SMTPProviderType];
                  onProviderChange(provider.provider as SMTPProviderType);
                }}
                defaultValue={Object.keys(providers)[0]}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select a provider" />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(providers).map(([provider, config]) => (
                    <SelectItem key={provider} value={provider}>
                      {provider}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <span className="text-sm text-muted-foreground flex items-center gap-1">
                Need help? Check out our documentation.{" "}
                <a
                  href="https://docs.xem.email/docs/smtp-providers"
                  target="_blank"
                  className="text-blue-500 hover:text-blue-600"
                >
                  Click here
                </a>
              </span>
            </div>
            <div className="space-y-2">
              <Label htmlFor="host">Host</Label>
              <Input
                id="host"
                placeholder="smtp.example.com"
                {...form.register("host")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="port">Port</Label>
              <Input
                id="port"
                type="number"
                placeholder="587"
                {...form.register("port")}
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Requires Authentication</Label>
                <p className="text-xs text-muted-foreground">
                  Provider requires username and password
                </p>
              </div>
              <Switch
                checked={form.watch("requiresAuth")}
                onCheckedChange={(checked) =>
                  form.setValue("requiresAuth", checked)
                }
              />
            </div>
            {form.watch("requiresAuth") && (
              <>
                <div className="space-y-2">
                  <Label htmlFor="username">Username</Label>
                  <Input
                    id="username"
                    placeholder="john@example.com"
                    {...form.register("username")}
                  />
                </div>
                <div className="space-y-2 relative">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    placeholder="********"
                    {...form.register("password")}
                  />
                  {!showPassword ? (
                    <Button
                      variant="ghost"
                      type="button"
                      size="icon"
                      className="absolute right-1 top-6"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      <Eye className="h-4 w-4" />
                    </Button>
                  ) : (
                    <Button
                      variant="ghost"
                      type="button"
                      size="icon"
                      className="absolute right-1 top-6"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      <EyeOff className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </>
            )}
            <div className="space-y-2">
              <Label htmlFor="fromEmail">From Address</Label>
              <Input
                id="fromEmail"
                type="text"
                placeholder="noreply@example.com"
                {...form.register("fromEmail")}
              />
              <p className="text-xs text-muted-foreground">
                Example: Human from Company &lt;noreply@example.com&gt;
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="maxSendRate">Max Send Rate</Label>
              <Input
                id="maxSendRate"
                placeholder="1,000"
                type="number"
                {...form.register("maxSendRate")}
              />
              <p className="text-xs text-muted-foreground">
                This is the maximum number of emails that can be sent per
                second.
              </p>
            </div>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Supports TLS</Label>
                <p className="text-xs text-muted-foreground">
                  Provider supports TLS encryption
                </p>
              </div>
              <Switch
                checked={form.watch("supportsTls")}
                onCheckedChange={(checked) =>
                  form.setValue("supportsTls", checked)
                }
              />
            </div>
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Default Provider</Label>
                <p className="text-xs text-muted-foreground">
                  Make this the default provider
                </p>
              </div>
              <Switch
                checked={form.watch("isDefault")}
                onCheckedChange={(checked) =>
                  form.setValue("isDefault", checked)
                }
              />
            </div>
            <span className="text-xs text-muted-foreground">
              Note: The default provider will be used for all email
              notifications.
            </span>
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
      </SheetContent>
    </Sheet>
  );
}
