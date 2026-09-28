"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Search } from "lucide-react";

export function CampaignsHeader() {
  return (
    <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 className="text-3xl font-semibold">Campaigns</h1>
        <p className="text-muted-foreground">Create and manage your email campaigns</p>
      </div>
      <div className="flex gap-4 items-center">
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input placeholder="Search campaigns..." className="pl-8" />
        </div>
        <Button>
          <Plus className="mr-2 h-4 w-4" /> Create Campaign
        </Button>
      </div>
    </div>
  );
}