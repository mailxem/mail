'use client';
import { Badge } from "@/components/ui/badge";
import { DataTable } from "@/components/ui/data-table";
import { ColumnDef } from "@tanstack/react-table";
import { Clock, Users } from "lucide-react";

type Automation = {
  id: string;
  name: string;
  status: string;
  triggers: string;
  actions: string[];
  lastRun: string;
}

const automations = [
  {
    id: "1",
    name: "Welcome Email Sequence",
    status: "Active", 
    triggers: "New Contact Added",
    actions: ["Send Email", "Wait 2 Days", "Send Follow-up"],
    lastRun: "2 hours ago",
  },
  {
    id: "2",
    name: "Lead Nurturing",
    status: "Draft",
    triggers: "Tag Added", 
    actions: ["Send Email Series", "Update Contact"],
    lastRun: "Never",
  },
];

const columns: ColumnDef<Automation>[] = [
  {
    accessorKey: "name",
    header: "Name",
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const status = row.getValue("status") as string;
      return (
        <Badge variant={status === "Active" ? "default" : "secondary"}>
          {status}
        </Badge>
      );
    },
  },
  {
    accessorKey: "triggers",
    header: "Triggers When",
  },
  {
    accessorKey: "actions",
    header: "Actions",
    cell: ({ row }) => {
      const actions = row.original.actions;
      return (
        <div className="flex items-center gap-2">
          <Users className="h-4 w-4" />
          <span>{actions.length} actions</span>
        </div>
      );
    },
  },
  {
    accessorKey: "lastRun",
    header: "Last Run",
    cell: ({ row }) => {
      const lastRun = row.getValue("lastRun") as string;
      return (
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4" />
          <span>{lastRun}</span>
        </div>
      );
    },
  },
];

export function AutomationsList() {
  return <DataTable columns={columns} data={automations} />;
}