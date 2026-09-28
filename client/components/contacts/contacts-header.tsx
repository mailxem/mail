"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Search } from "lucide-react";

export function ContactsHeader() {
  return (
    <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 className="text-3xl font-semibold">Contacts</h1>
        <p className="text-muted-foreground">Manage your contacts and lists</p>
      </div>
      <div className="flex gap-4 items-center">
        <div className="relative">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input placeholder="Search contacts..." className="pl-8 bg-card" />
        </div>
        <Button>
          <Plus className="mr-2 h-4 w-4" /> Add Contact
        </Button>
      </div>
    </div>
  );
}