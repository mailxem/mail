import { Card } from "@/components/ui/card";
import { LucideIcon } from "lucide-react";

interface StatCardProps {
  name: string;
  value: string;
  icon: LucideIcon;
  change: string;
}

export function StatCard({ name, value, icon: Icon, change }: StatCardProps) {
  return (
    <Card className="p-4">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-muted-foreground">{name}</p>
          <h3 className="text-2xl font-semibold">{value}</h3>
        </div>
        <div className=" p-2 bg-primary/10">
          <Icon className="h-4 w-4 text-primary" />
        </div>
      </div>
      <div className="mt-2 text-sm text-green-600">{change}</div>
    </Card>
  );
}
