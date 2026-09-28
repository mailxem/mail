"use client";

import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";

export function NotificationSettings() {
  return (
    <div className="space-y-6">
      <h2 className="text-lg font-medium">Notification Settings</h2>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Label htmlFor="email-notifications">Email Notifications</Label>
          <Switch id="email-notifications" />
        </div>
        <div className="flex items-center justify-between">
          <Label htmlFor="marketing-emails">Marketing Emails</Label>
          <Switch id="marketing-emails" />
        </div>
        <div className="flex items-center justify-between">
          <Label htmlFor="security-alerts">Security Alerts</Label>
          <Switch id="security-alerts" defaultChecked />
        </div>
      </div>
    </div>
  );
}